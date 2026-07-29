package main

import (
	"context"
	"net/http"
	"slices"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	auth "github.com/jamesread/httpauthshim"
	"github.com/jamesread/httpauthshim/authpublic"

	faridoonv1 "faridoon/service/gen/faridoon/v1"
	"faridoon/service/gen/faridoon/v1/faridoonv1connect"
	"faridoon/service/internal/config"
	"faridoon/service/internal/quote"
	"faridoon/service/internal/store"
	"faridoon/service/internal/webhook"
)

type FaridoonServer struct {
	faridoonv1connect.UnimplementedFaridoonServiceHandler
	cfg       *config.Config
	store     store.Store
	auth      *auth.AuthShimContext
	formatter *quote.Formatter
	webhooks  *webhook.Dispatcher
}

type sessionUser struct {
	Username   string
	GroupTitle string
	Privileges []string
	ID         int
	GroupID    int
}

func (s *FaridoonServer) httpReq(ctx context.Context) *http.Request {
	r, _ := ctx.Value(httpRequestKey).(*http.Request)
	return r
}

func (s *FaridoonServer) authFromRequest(r *http.Request) *authpublic.AuthenticatedUser {
	u := s.auth.AuthFromHttpReq(r)
	if u == nil || u.IsGuest() || u.Username == "" {
		return nil
	}
	return u
}

func (s *FaridoonServer) authUser(ctx context.Context) *authpublic.AuthenticatedUser {
	if s.auth == nil {
		return nil
	}
	r := s.httpReq(ctx)
	if r == nil {
		return nil
	}
	return s.authFromRequest(r)
}

func (s *FaridoonServer) loadSessionUser(ctx context.Context) (*sessionUser, error) {
	au := s.authUser(ctx)
	if au == nil {
		return nil, nil
	}
	row, err := s.store.FindUserByUsername(ctx, au.Username)
	if err != nil || row == nil {
		return nil, err
	}
	privs, err := s.store.UserPrivileges(ctx, row.ID, row.GroupID)
	if err != nil {
		return nil, err
	}
	return &sessionUser{
		Username: row.Username, GroupTitle: row.GroupTitle, Privileges: privs,
		ID: row.ID, GroupID: row.GroupID,
	}, nil
}

func (su *sessionUser) hasPriv(key string) bool {
	if su == nil {
		return false
	}
	if slices.Contains(su.Privileges, "SUPERUSER") {
		return true
	}
	return slices.Contains(su.Privileges, key)
}

func (su *sessionUser) isAdmin() bool {
	return su.hasPriv("SUPERUSER")
}

func (s *FaridoonServer) requireAuth(ctx context.Context) (*sessionUser, error) {
	u, err := s.loadSessionUser(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if u == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errLoginRequired)
	}
	return u, nil
}

func (s *FaridoonServer) requirePriv(ctx context.Context, priv string) (*sessionUser, error) {
	u, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !u.hasPriv(priv) {
		return nil, connect.NewError(connect.CodePermissionDenied, errForbidden)
	}
	return u, nil
}

func (s *FaridoonServer) requireAdmin(ctx context.Context) (*sessionUser, error) {
	u, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !u.isAdmin() {
		return nil, connect.NewError(connect.CodePermissionDenied, errForbidden)
	}
	return u, nil
}

func (s *FaridoonServer) toProtoUser(su *sessionUser) *faridoonv1.User {
	if su == nil {
		return nil
	}
	return &faridoonv1.User{
		Id: int32(su.ID), Username: su.Username, GroupId: int32(su.GroupID),
		GroupTitle: su.GroupTitle, IsAdmin: su.isAdmin(),
		CanApproveQuotes: su.hasPriv("APPROVE_QUOTES"), CanBypassApproval: su.hasPriv("BYPASS_APPROVAL"),
		Privileges: su.Privileges,
	}
}

func (s *FaridoonServer) toProtoUserRow(ctx context.Context, row *store.UserRow) *faridoonv1.User {
	if row == nil {
		return nil
	}
	privs, _ := s.store.UserPrivileges(ctx, row.ID, row.GroupID)
	su := &sessionUser{Username: row.Username, GroupTitle: row.GroupTitle, Privileges: privs, ID: row.ID, GroupID: row.GroupID}
	return s.toProtoUser(su)
}

func (s *FaridoonServer) formatQuote(q *store.QuoteRow) *faridoonv1.Quote {
	if q == nil {
		return nil
	}
	f := s.formatter.Format(q.ID, q.Content, q.Created, q.VoteCount, q.Approved, q.SyntaxHighlighting)
	out := &faridoonv1.Quote{
		Id: int32(f.ID), Content: f.Content, Created: f.Created, VoteCount: int32(f.VoteCount),
		Approved: f.Approved, SyntaxHighlighting: f.SyntaxHighlighting,
		SubmittedByUserId: int32(q.SubmittedByUserID), SubmittedByUsername: q.SubmittedByUsername,
	}
	for _, line := range f.Lines {
		out.Lines = append(out.Lines, &faridoonv1.QuoteLine{
			Content: line.Content, Username: line.Username, UsernameColor: int32(line.UsernameColor),
		})
	}
	return out
}

func newSessionID() string {
	return uuid.New().String()
}

var (
	errLoginRequired = errString("login required")
	errForbidden     = errString("forbidden")
)

type errString string

func (e errString) Error() string { return string(e) }
