package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"faridoon/service/buildinfo"
	faridoonv1 "faridoon/service/gen/faridoon/v1"
	"faridoon/service/internal/authpass"
	"faridoon/service/internal/quote"
	"faridoon/service/internal/store"
	"faridoon/service/internal/webhook"
)

func (s *FaridoonServer) Init(ctx context.Context, _ *connect.Request[faridoonv1.InitRequest]) (*connect.Response[faridoonv1.InitResponse], error) {
	su, err := s.loadSessionUser(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	pending := int32(0)
	if su != nil && su.hasPriv("APPROVE_QUOTES") {
		n, _ := s.store.CountPending(ctx)
		pending = int32(n)
	}
	return connect.NewResponse(&faridoonv1.InitResponse{
		Version: buildinfo.Version, SiteTitle: s.siteTitle(ctx), Features: s.featureFlags(ctx),
		User: s.toProtoUser(su), PendingApprovals: pending, WebhookEvents: webhook.SupportedEvents,
		HeaderLinks: s.loadEnabledHeaderLinks(ctx),
		Theme:       s.themeSettings(ctx),
	}), nil
}

func (s *FaridoonServer) featureFlags(ctx context.Context) *faridoonv1.Features {
	return &faridoonv1.Features{
		VotingEnabled:             s.votingEnabled(ctx),
		RegistrationEnabled:       s.registrationEnabled(ctx),
		GuestAddEnabled:           s.guestAddEnabled(ctx),
		SyntaxHighlightingEnabled: s.syntaxHighlightingEnabled(ctx),
		ShowPwaPrompt:             s.showPwaPrompt(ctx),
		MarkdownEnabled:           s.markdownEnabled(ctx),
	}
}

func (s *FaridoonServer) GetCurrentUser(ctx context.Context, _ *connect.Request[faridoonv1.GetCurrentUserRequest]) (*connect.Response[faridoonv1.GetCurrentUserResponse], error) {
	su, err := s.loadSessionUser(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&faridoonv1.GetCurrentUserResponse{User: s.toProtoUser(su)}), nil
}

func (s *FaridoonServer) maybeRehashPassword(ctx context.Context, row *store.UserRow, password string) {
	if !authpass.NeedsRehash(row.Password) {
		return
	}
	h, hashErr := authpass.Hash(password)
	if hashErr != nil {
		return
	}
	_ = s.store.UpdatePassword(ctx, row.ID, h)
}

func (s *FaridoonServer) authenticateLoginUser(ctx context.Context, username, password string) (*store.UserRow, error) {
	row, err := s.store.FindUserByUsername(ctx, username)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if row == nil || !authpass.Verify(password, row.Password) {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid username or password"))
	}
	return row, nil
}

func (s *FaridoonServer) loginResponse(ctx context.Context, row *store.UserRow, username string, privs []string) (*connect.Response[faridoonv1.LoginResponse], error) {
	if s.auth == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("auth not configured"))
	}
	sid := newSessionID()
	s.auth.RegisterUserSession("faridoon", sid, username, strings.Join(privs, ","))
	su := &sessionUser{Username: row.Username, GroupTitle: row.GroupTitle, Privileges: privs, ID: row.ID, GroupID: row.GroupID}
	s.audit(ctx, su, "user.login", "user", su.ID, "")
	res := connect.NewResponse(&faridoonv1.LoginResponse{User: s.toProtoUser(su)})
	s.attachSessionCookie(res.Header(), sid)
	return res, nil
}

func (s *FaridoonServer) Login(ctx context.Context, req *connect.Request[faridoonv1.LoginRequest]) (*connect.Response[faridoonv1.LoginResponse], error) {
	username := strings.TrimSpace(req.Msg.Username)
	password := req.Msg.Password
	if username == "" || password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("username and password required"))
	}
	row, authErr := s.authenticateLoginUser(ctx, username, password)
	if authErr != nil {
		return nil, authErr
	}
	s.maybeRehashPassword(ctx, row, password)
	privs, err := s.store.UserPrivileges(ctx, row.ID, row.GroupID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return s.loginResponse(ctx, row, username, privs)
}

func (s *FaridoonServer) Logout(ctx context.Context, _ *connect.Request[faridoonv1.LogoutRequest]) (*connect.Response[emptypb.Empty], error) {
	su, _ := s.loadSessionUser(ctx)
	s.audit(ctx, su, "user.logout", "user", 0, "")
	res := connect.NewResponse(&emptypb.Empty{})
	s.clearSessionCookie(res.Header())
	if r := s.httpReq(ctx); r != nil && s.auth != nil {
		if c, err := r.Cookie(s.auth.Config.GetLocalSessionCookieName()); err == nil {
			s.auth.DeleteUserSession("faridoon", c.Value)
		}
	}
	return res, nil
}

func validateRegisterRequest(req *faridoonv1.RegisterRequest) error {
	username := strings.TrimSpace(req.Username)
	password := req.Password
	if username == "" || len(password) < 4 {
		return fmt.Errorf("invalid username or password")
	}
	if password != req.PasswordConfirmation {
		return fmt.Errorf("password confirmation mismatch")
	}
	return nil
}

func registerGroupID(count int) int {
	if count == 0 {
		return 1
	}
	return 2
}

func (s *FaridoonServer) registerSession(ctx context.Context, id int, username string, groupID int) (*sessionUser, string, error) {
	row, _ := s.store.FindUser(ctx, id)
	privs, _ := s.store.UserPrivileges(ctx, id, groupID)
	sid := newSessionID()
	s.auth.RegisterUserSession("faridoon", sid, username, strings.Join(privs, ","))
	groupTitle := ""
	if row != nil {
		groupTitle = row.GroupTitle
	}
	su := &sessionUser{Username: username, GroupTitle: groupTitle, Privileges: privs, ID: id, GroupID: groupID}
	return su, sid, nil
}

func (s *FaridoonServer) ensureUsernameAvailable(ctx context.Context, username string) error {
	existing, findErr := s.store.FindUserByUsername(ctx, username)
	if findErr != nil {
		return connect.NewError(connect.CodeInternal, findErr)
	}
	if existing != nil {
		return connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("username taken"))
	}
	return nil
}

func (s *FaridoonServer) insertRegisteredUser(ctx context.Context, username, password string, groupID int) (int, error) {
	hash, hashErr := authpass.Hash(password)
	if hashErr != nil {
		return 0, connect.NewError(connect.CodeInternal, hashErr)
	}
	id, createErr := s.store.CreateUser(ctx, username, hash, groupID)
	if createErr != nil {
		return 0, connect.NewError(connect.CodeInternal, createErr)
	}
	return id, nil
}

func (s *FaridoonServer) createRegisteredUser(ctx context.Context, username, password string) (id, groupID int, err error) {
	if availErr := s.ensureUsernameAvailable(ctx, username); availErr != nil {
		return 0, 0, availErr
	}
	count, countErr := s.store.UserCount(ctx)
	if countErr != nil {
		return 0, 0, connect.NewError(connect.CodeInternal, countErr)
	}
	groupID = registerGroupID(count)
	id, insertErr := s.insertRegisteredUser(ctx, username, password, groupID)
	if insertErr != nil {
		return 0, 0, insertErr
	}
	return id, groupID, nil
}

func (s *FaridoonServer) Register(ctx context.Context, req *connect.Request[faridoonv1.RegisterRequest]) (*connect.Response[faridoonv1.RegisterResponse], error) {
	if !s.registrationEnabled(ctx) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("registration disabled"))
	}
	if valErr := validateRegisterRequest(req.Msg); valErr != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, valErr)
	}
	username := strings.TrimSpace(req.Msg.Username)
	id, groupID, createErr := s.createRegisteredUser(ctx, username, req.Msg.Password)
	if createErr != nil {
		return nil, createErr
	}
	su, sid, err := s.registerSession(ctx, id, username, groupID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	s.audit(ctx, su, "user.register", "user", id, detailf("group=%d", groupID))
	res := connect.NewResponse(&faridoonv1.RegisterResponse{User: s.toProtoUser(su)})
	s.attachSessionCookie(res.Header(), sid)
	return res, nil
}

func (s *FaridoonServer) attachSessionCookie(h http.Header, sid string) {
	name := "faridoon_session"
	if s.auth != nil && s.auth.Config != nil {
		name = s.auth.Config.GetLocalSessionCookieName()
	}
	cookie := &http.Cookie{Name: name, Value: sid, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 7 * 24 * 3600}
	h.Add("Set-Cookie", cookie.String())
}

func (s *FaridoonServer) clearSessionCookie(h http.Header) {
	name := "faridoon_session"
	if s.auth != nil && s.auth.Config != nil {
		name = s.auth.Config.GetLocalSessionCookieName()
	}
	cookie := &http.Cookie{Name: name, Value: "", Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: -1}
	h.Add("Set-Cookie", cookie.String())
}

func (s *FaridoonServer) ListQuotes(ctx context.Context, req *connect.Request[faridoonv1.ListQuotesRequest]) (*connect.Response[faridoonv1.ListQuotesResponse], error) {
	page, order, query, perPage := s.listQuotesParams(ctx, req.Msg)
	rows, total, err := s.store.ListApproved(ctx, order, page, perPage, query)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	totalPages := (total + perPage - 1) / perPage
	out := &faridoonv1.ListQuotesResponse{Page: int32(page), Total: int32(total), TotalPages: int32(totalPages)}
	out.Quotes = s.approvedQuoteProtos(ctx, rows)
	return connect.NewResponse(out), nil
}

func (s *FaridoonServer) listQuotesParams(ctx context.Context, msg *faridoonv1.ListQuotesRequest) (page int, order, query string, perPage int) {
	page = int(msg.Page)
	if page < 1 {
		page = 1
	}
	order = msg.Order
	if order == "" {
		order = "latest"
	}
	query = strings.TrimSpace(msg.Query)
	perPage = s.quotesPerPage(ctx)
	if query != "" {
		perPage = 15
	}
	return page, order, query, perPage
}

func (s *FaridoonServer) approvedQuoteProtos(ctx context.Context, rows []store.QuoteRow) []*faridoonv1.Quote {
	out := make([]*faridoonv1.Quote, 0, len(rows))
	for i := range rows {
		// Defense in depth: search/list must never expose pending quotes.
		if !rows[i].Approved {
			continue
		}
		out = append(out, s.formatQuote(ctx, &rows[i]))
	}
	return out
}

func (s *FaridoonServer) GetQuote(ctx context.Context, req *connect.Request[faridoonv1.GetQuoteRequest]) (*connect.Response[faridoonv1.GetQuoteResponse], error) {
	q, err := s.store.FindQuote(ctx, int(req.Msg.Id))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if q == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("quote not found"))
	}
	if !q.Approved {
		if gateErr := s.authorizePendingQuoteRead(ctx, q); gateErr != nil {
			return nil, gateErr
		}
	}
	return connect.NewResponse(&faridoonv1.GetQuoteResponse{Quote: s.formatQuote(ctx, q)}), nil
}

func (s *FaridoonServer) authorizePendingQuoteRead(ctx context.Context, q *store.QuoteRow) error {
	su, err := s.loadSessionUser(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if canReadPendingQuote(su, q) {
		return nil
	}
	return connect.NewError(connect.CodeNotFound, fmt.Errorf("quote not found"))
}

func canReadPendingQuote(su *sessionUser, q *store.QuoteRow) bool {
	if su == nil {
		return false
	}
	if su.hasPriv("APPROVE_QUOTES") {
		return true
	}
	return q.SubmittedByUserID > 0 && su.ID == q.SubmittedByUserID
}

func canGuestAdd(su *sessionUser, guestAddEnabled bool) error {
	if su == nil && !guestAddEnabled {
		return fmt.Errorf("guests cannot add quotes")
	}
	return nil
}

func approvalForSubmission(su *sessionUser, guestRequireApproval bool) int {
	if su == nil {
		if guestRequireApproval {
			return 0
		}
		return 1
	}
	if su.hasPriv("BYPASS_APPROVAL") {
		return 1
	}
	return 0
}

func normalizeQuoteContent(raw string) (string, error) {
	content := quote.NormalizeInput(raw)
	if content == "" {
		return "", fmt.Errorf("content required")
	}
	return content, nil
}

func (s *FaridoonServer) dispatchPendingApproval(ctx context.Context, id, approval int) {
	if approval != 0 {
		return
	}
	raw, _ := s.store.FindQuoteRaw(ctx, id)
	if raw != nil {
		s.webhooks.DispatchApprovalRequested(ctx, raw)
	}
}

func submitterFromSession(su *sessionUser) (userID int, username string) {
	if su == nil {
		return 0, "Guest"
	}
	return su.ID, su.Username
}

func (s *FaridoonServer) CreateQuote(ctx context.Context, req *connect.Request[faridoonv1.CreateQuoteRequest]) (*connect.Response[faridoonv1.CreateQuoteResponse], error) {
	su, _ := s.loadSessionUser(ctx)
	if guestErr := canGuestAdd(su, s.guestAddEnabled(ctx)); guestErr != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, guestErr)
	}
	content, contentErr := normalizeQuoteContent(req.Msg.Content)
	if contentErr != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, contentErr)
	}
	approval := approvalForSubmission(su, s.guestAddRequireApproval(ctx))
	userID, username := submitterFromSession(su)
	markdownEnabled := req.Msg.MarkdownEnabled && s.markdownEnabled(ctx)
	id, err := s.store.CreateQuote(ctx, content, approval, req.Msg.SyntaxHighlighting, markdownEnabled, userID, username)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	s.dispatchPendingApproval(ctx, id, approval)
	s.audit(ctx, su, "quote.create", "quote", id, detailf("approval=%d", approval))
	q, _ := s.store.FindQuote(ctx, id)
	return connect.NewResponse(&faridoonv1.CreateQuoteResponse{
		Quote: s.formatQuote(ctx, q), PendingApproval: approval == 0,
	}), nil
}

func (s *FaridoonServer) FormatQuote(ctx context.Context, req *connect.Request[faridoonv1.FormatQuoteRequest]) (*connect.Response[faridoonv1.FormatQuoteResponse], error) {
	// NormalizeInput applies Discord paste fixup and newline normalization.
	content := quote.NormalizeInput(req.Msg.Content)
	markdownEnabled := req.Msg.MarkdownEnabled && s.markdownEnabled(ctx)
	f := s.formatter.Format(0, content, "", 0, false, "")
	lines, signatureHTML := s.formatQuoteLines(ctx, f, markdownEnabled)
	return connect.NewResponse(&faridoonv1.FormatQuoteResponse{
		Lines:               lines,
		FormatStyle:         f.FormatStyle,
		SignatureAuthor:     f.SignatureAuthor,
		SignatureAuthorHtml: signatureHTML,
	}), nil
}

func (s *FaridoonServer) UpdateQuote(ctx context.Context, req *connect.Request[faridoonv1.UpdateQuoteRequest]) (*connect.Response[faridoonv1.Quote], error) {
	su, err := s.requirePriv(ctx, "APPROVE_QUOTES")
	if err != nil {
		return nil, err
	}
	q, err := s.saveQuoteEdit(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, su, "quote.update", "quote", int(req.Msg.Id), "")
	return connect.NewResponse(s.formatQuote(ctx, q)), nil
}

func (s *FaridoonServer) saveQuoteEdit(ctx context.Context, msg *faridoonv1.UpdateQuoteRequest) (*store.QuoteRow, error) {
	content := quote.NormalizeInput(strings.TrimSpace(msg.Content))
	markdownEnabled := msg.MarkdownEnabled && s.markdownEnabled(ctx)
	if err := s.store.UpdateQuote(ctx, int(msg.Id), content, msg.SyntaxHighlighting, markdownEnabled); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	q, err := s.store.FindQuote(ctx, int(msg.Id))
	if err != nil || q == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("quote not found"))
	}
	return q, nil
}

func (s *FaridoonServer) DeleteQuote(ctx context.Context, req *connect.Request[faridoonv1.DeleteQuoteRequest]) (*connect.Response[emptypb.Empty], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if delErr := s.store.DeleteQuote(ctx, int(req.Msg.Id)); delErr != nil {
		return nil, connect.NewError(connect.CodeInternal, delErr)
	}
	s.audit(ctx, su, "quote.delete", "quote", int(req.Msg.Id), "")
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func validateVoteDelta(delta int) error {
	if delta != 1 && delta != -1 {
		return fmt.Errorf("delta must be 1 or -1")
	}
	return nil
}

func (s *FaridoonServer) castVoteAndSum(ctx context.Context, quoteID, userID, delta int) (int, error) {
	if voteErr := s.store.CastVote(ctx, quoteID, userID, delta); voteErr != nil {
		return 0, voteErr
	}
	return s.store.VoteSum(ctx, quoteID)
}

func (s *FaridoonServer) VoteQuote(ctx context.Context, req *connect.Request[faridoonv1.VoteQuoteRequest]) (*connect.Response[faridoonv1.VoteQuoteResponse], error) {
	su, err := s.requireVotingUser(ctx)
	if err != nil {
		return nil, err
	}
	delta := int(req.Msg.Delta)
	if deltaErr := validateVoteDelta(delta); deltaErr != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, deltaErr)
	}
	sum, castErr := s.voteOnApprovedQuote(ctx, int(req.Msg.Id), su.ID, delta)
	if castErr != nil {
		return nil, castErr
	}
	s.audit(ctx, su, "quote.vote", "quote", int(req.Msg.Id), detailf("delta=%d", delta))
	return connect.NewResponse(&faridoonv1.VoteQuoteResponse{VoteCount: int32(sum)}), nil
}

func (s *FaridoonServer) requireVotingUser(ctx context.Context) (*sessionUser, error) {
	if !s.votingEnabled(ctx) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("voting disabled"))
	}
	return s.requireAuth(ctx)
}

func (s *FaridoonServer) voteOnApprovedQuote(ctx context.Context, quoteID, userID, delta int) (int, error) {
	if voteErr := s.requireApprovedQuote(ctx, quoteID); voteErr != nil {
		return 0, voteErr
	}
	sum, castErr := s.castVoteAndSum(ctx, quoteID, userID, delta)
	if castErr != nil {
		return 0, connect.NewError(connect.CodeInternal, castErr)
	}
	return sum, nil
}

func (s *FaridoonServer) requireApprovedQuote(ctx context.Context, id int) error {
	q, findErr := s.store.FindQuote(ctx, id)
	if findErr != nil {
		return connect.NewError(connect.CodeInternal, findErr)
	}
	if q == nil || !q.Approved {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("quote not found"))
	}
	return nil
}

func (s *FaridoonServer) ListApprovals(ctx context.Context, _ *connect.Request[faridoonv1.ListApprovalsRequest]) (*connect.Response[faridoonv1.ListApprovalsResponse], error) {
	if _, err := s.requirePriv(ctx, "APPROVE_QUOTES"); err != nil {
		return nil, err
	}
	rows, err := s.store.ListPending(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &faridoonv1.ListApprovalsResponse{}
	for i := range rows {
		out.Quotes = append(out.Quotes, s.formatQuote(ctx, &rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (s *FaridoonServer) ApproveQuote(ctx context.Context, req *connect.Request[faridoonv1.ApproveQuoteRequest]) (*connect.Response[emptypb.Empty], error) {
	su, err := s.requirePriv(ctx, "APPROVE_QUOTES")
	if err != nil {
		return nil, err
	}
	if _, pendErr := s.requirePendingQuote(ctx, int(req.Msg.Id)); pendErr != nil {
		return nil, pendErr
	}
	if approveErr := s.store.ApproveQuote(ctx, int(req.Msg.Id)); approveErr != nil {
		return nil, connect.NewError(connect.CodeInternal, approveErr)
	}
	s.audit(ctx, su, "quote.approve", "quote", int(req.Msg.Id), "")
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *FaridoonServer) RejectQuote(ctx context.Context, req *connect.Request[faridoonv1.RejectQuoteRequest]) (*connect.Response[emptypb.Empty], error) {
	su, err := s.requirePriv(ctx, "APPROVE_QUOTES")
	if err != nil {
		return nil, err
	}
	if _, pendErr := s.requirePendingQuote(ctx, int(req.Msg.Id)); pendErr != nil {
		return nil, pendErr
	}
	if delErr := s.store.DeleteQuote(ctx, int(req.Msg.Id)); delErr != nil {
		return nil, connect.NewError(connect.CodeInternal, delErr)
	}
	s.audit(ctx, su, "quote.reject", "quote", int(req.Msg.Id), "")
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *FaridoonServer) requirePendingQuote(ctx context.Context, id int) (*store.QuoteRow, error) {
	raw, findErr := s.store.FindQuoteRaw(ctx, id)
	if findErr != nil || raw == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("quote not found"))
	}
	if raw.Approved {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("quote is already approved"))
	}
	return raw, nil
}
