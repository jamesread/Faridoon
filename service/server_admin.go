package main

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	faridoonv1 "faridoon/service/gen/faridoon/v1"
	"faridoon/service/internal/authpass"
	"faridoon/service/internal/store"
	"faridoon/service/internal/webhook"
)

func (s *FaridoonServer) fillUsers(ctx context.Context, out *faridoonv1.ListUsersResponse, users []store.UserRow) {
	for i := range users {
		u := users[i]
		u.Password = ""
		out.Users = append(out.Users, s.toProtoUserRow(ctx, &u))
	}
}

func (s *FaridoonServer) fillGroups(out *faridoonv1.ListUsersResponse, groups []store.GroupRow, gperms map[int][]store.GroupPermissionRow) {
	for _, g := range groups {
		pg := &faridoonv1.Group{Id: int32(g.ID), Title: g.Title}
		for _, p := range gperms[g.ID] {
			pg.Permissions = append(pg.Permissions, &faridoonv1.GroupPermission{
				PermissionId: int32(p.PermissionID), Key: p.Key, Description: p.Description,
			})
		}
		out.Groups = append(out.Groups, pg)
	}
}

func fillPermissions(out *faridoonv1.ListUsersResponse, perms []store.PermissionRow) {
	for _, p := range perms {
		out.Permissions = append(out.Permissions, &faridoonv1.Permission{
			Id: int32(p.ID), Key: p.Key, Description: p.Description,
		})
	}
}

func (s *FaridoonServer) ListUsers(ctx context.Context, _ *connect.Request[faridoonv1.ListUsersRequest]) (*connect.Response[faridoonv1.ListUsersResponse], error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	users, err := s.store.ListUsers(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	groups, gperms, err := s.store.GroupsWithPermissions(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	perms, err := s.store.ListPermissions(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &faridoonv1.ListUsersResponse{}
	s.fillUsers(ctx, out, users)
	s.fillGroups(out, groups, gperms)
	fillPermissions(out, perms)
	return connect.NewResponse(out), nil
}

func protoGroups(groups []store.GroupRow) []*faridoonv1.Group {
	out := make([]*faridoonv1.Group, 0, len(groups))
	for _, g := range groups {
		out = append(out, &faridoonv1.Group{Id: int32(g.ID), Title: g.Title})
	}
	return out
}

func (s *FaridoonServer) GetUser(ctx context.Context, req *connect.Request[faridoonv1.GetUserRequest]) (*connect.Response[faridoonv1.GetUserResponse], error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	row, err := s.store.FindUser(ctx, int(req.Msg.Id))
	if err != nil || row == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
	}
	groups, _, err := s.store.GroupsWithPermissions(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&faridoonv1.GetUserResponse{
		User: s.toProtoUserRow(ctx, row), Groups: protoGroups(groups),
	}), nil
}

func toProtoGroup(g *store.GroupRow, perms []store.GroupPermissionRow) *faridoonv1.Group {
	pg := &faridoonv1.Group{Id: int32(g.ID), Title: g.Title}
	for _, p := range perms {
		pg.Permissions = append(pg.Permissions, &faridoonv1.GroupPermission{
			PermissionId: int32(p.PermissionID), Key: p.Key, Description: p.Description,
		})
	}
	return pg
}

func (s *FaridoonServer) membersForGroup(ctx context.Context, groupID int) ([]*faridoonv1.User, error) {
	users, err := s.store.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	var members []*faridoonv1.User
	for i := range users {
		if users[i].GroupID != groupID {
			continue
		}
		members = append(members, s.toProtoUserRow(ctx, &users[i]))
	}
	return members, nil
}

func appendProtoPermissions(dst []*faridoonv1.Permission, perms []store.PermissionRow) []*faridoonv1.Permission {
	for _, p := range perms {
		dst = append(dst, &faridoonv1.Permission{
			Id: int32(p.ID), Key: p.Key, Description: p.Description,
		})
	}
	return dst
}

func (s *FaridoonServer) buildGetGroupResponse(ctx context.Context, row *store.GroupRow) (*faridoonv1.GetGroupResponse, error) {
	perms, err := s.store.GroupPermissions(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	members, err := s.membersForGroup(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	allPerms, err := s.store.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	return &faridoonv1.GetGroupResponse{
		Group: toProtoGroup(row, perms), Members: members,
		Permissions: appendProtoPermissions(nil, allPerms),
	}, nil
}

func (s *FaridoonServer) GetGroup(ctx context.Context, req *connect.Request[faridoonv1.GetGroupRequest]) (*connect.Response[faridoonv1.GetGroupResponse], error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	row, err := s.store.FindGroup(ctx, int(req.Msg.Id))
	if err != nil || row == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("group not found"))
	}
	out, buildErr := s.buildGetGroupResponse(ctx, row)
	if buildErr != nil {
		return nil, connect.NewError(connect.CodeInternal, buildErr)
	}
	return connect.NewResponse(out), nil
}

func (s *FaridoonServer) UpdateUser(ctx context.Context, req *connect.Request[faridoonv1.UpdateUserRequest]) (*connect.Response[faridoonv1.User], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if applyErr := s.applyUserGroupUpdate(ctx, int(req.Msg.Id), int(req.Msg.GroupId)); applyErr != nil {
		return nil, applyErr
	}
	s.audit(ctx, su, "user.update_group", "user", int(req.Msg.Id), detailf("group_id=%d", req.Msg.GroupId))
	row, loadErr := s.requireUser(ctx, int(req.Msg.Id))
	if loadErr != nil {
		return nil, loadErr
	}
	return connect.NewResponse(s.toProtoUserRow(ctx, row)), nil
}

func (s *FaridoonServer) applyUserGroupUpdate(ctx context.Context, userID, groupID int) error {
	target, findErr := s.requireUser(ctx, userID)
	if findErr != nil {
		return findErr
	}
	if lockErr := s.denyIfLastAdminDemotion(ctx, target, groupID); lockErr != nil {
		return lockErr
	}
	if updErr := s.store.UpdateUserGroup(ctx, userID, groupID); updErr != nil {
		return connect.NewError(connect.CodeInternal, updErr)
	}
	return nil
}

func (s *FaridoonServer) requireUser(ctx context.Context, id int) (*store.UserRow, error) {
	row, err := s.store.FindUser(ctx, id)
	if err != nil || row == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
	}
	return row, nil
}

func (s *FaridoonServer) denyIfLastAdminDemotion(ctx context.Context, target *store.UserRow, newGroupID int) error {
	oldPrivs, err := s.store.UserPrivileges(ctx, target.ID, target.GroupID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if !slices.Contains(oldPrivs, "SUPERUSER") {
		return nil
	}
	newPrivs, err := s.store.UserPrivileges(ctx, target.ID, newGroupID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if slices.Contains(newPrivs, "SUPERUSER") {
		return nil
	}
	return s.denyIfSoleAdmin(ctx)
}

func (s *FaridoonServer) denyIfSoleAdmin(ctx context.Context) error {
	n, err := s.store.CountUsersWithPrivilege(ctx, "SUPERUSER")
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if n <= 1 {
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("cannot remove the last admin"))
	}
	return nil
}

func (s *FaridoonServer) DeleteUser(ctx context.Context, req *connect.Request[faridoonv1.DeleteUserRequest]) (*connect.Response[emptypb.Empty], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if delErr := s.deleteUserByID(ctx, su, int(req.Msg.Id)); delErr != nil {
		return nil, delErr
	}
	s.audit(ctx, su, "user.delete", "user", int(req.Msg.Id), "")
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *FaridoonServer) deleteUserByID(ctx context.Context, su *sessionUser, userID int) error {
	if su.ID == userID {
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("cannot delete your own account"))
	}
	target, findErr := s.requireUser(ctx, userID)
	if findErr != nil {
		return findErr
	}
	if lockErr := s.denyIfDeletingAdmin(ctx, target); lockErr != nil {
		return lockErr
	}
	if delErr := s.store.DeleteUser(ctx, userID); delErr != nil {
		return connect.NewError(connect.CodeInternal, delErr)
	}
	return nil
}

func (s *FaridoonServer) denyIfDeletingAdmin(ctx context.Context, target *store.UserRow) error {
	privs, privErr := s.store.UserPrivileges(ctx, target.ID, target.GroupID)
	if privErr != nil {
		return connect.NewError(connect.CodeInternal, privErr)
	}
	if !slices.Contains(privs, "SUPERUSER") {
		return nil
	}
	return s.denyIfSoleAdmin(ctx)
}

func validateResetPassword(password, confirmation string) error {
	if len(password) < 4 {
		return fmt.Errorf("password must be at least 4 characters")
	}
	if password != confirmation {
		return fmt.Errorf("password confirmation mismatch")
	}
	return nil
}

func (s *FaridoonServer) ResetUserPassword(ctx context.Context, req *connect.Request[faridoonv1.ResetUserPasswordRequest]) (*connect.Response[emptypb.Empty], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if valErr := validateResetPassword(req.Msg.Password, req.Msg.PasswordConfirmation); valErr != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, valErr)
	}
	row, findErr := s.requireUser(ctx, int(req.Msg.Id))
	if findErr != nil {
		return nil, findErr
	}
	if setErr := s.setUserPassword(ctx, row.ID, req.Msg.Password); setErr != nil {
		return nil, setErr
	}
	s.audit(ctx, su, "user.reset_password", "user", row.ID, row.Username)
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *FaridoonServer) setUserPassword(ctx context.Context, userID int, password string) error {
	hash, hashErr := authpass.Hash(password)
	if hashErr != nil {
		return connect.NewError(connect.CodeInternal, hashErr)
	}
	if updErr := s.store.UpdatePassword(ctx, userID, hash); updErr != nil {
		return connect.NewError(connect.CodeInternal, updErr)
	}
	return nil
}

func (s *FaridoonServer) CreateGroup(ctx context.Context, req *connect.Request[faridoonv1.CreateGroupRequest]) (*connect.Response[faridoonv1.CreateGroupResponse], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(req.Msg.Title)
	if title == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("title required"))
	}
	id, err := s.store.CreateGroup(ctx, title)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	s.audit(ctx, su, "group.create", "group", id, title)
	return connect.NewResponse(&faridoonv1.CreateGroupResponse{
		Group: &faridoonv1.Group{Id: int32(id), Title: title},
	}), nil
}

func (s *FaridoonServer) DeleteGroup(ctx context.Context, req *connect.Request[faridoonv1.DeleteGroupRequest]) (*connect.Response[emptypb.Empty], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if req.Msg.Id <= 2 {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("cannot delete system group"))
	}
	if delErr := s.store.DeleteGroup(ctx, int(req.Msg.Id)); delErr != nil {
		return nil, connect.NewError(connect.CodeInternal, delErr)
	}
	s.audit(ctx, su, "group.delete", "group", int(req.Msg.Id), "")
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *FaridoonServer) GrantPermission(ctx context.Context, req *connect.Request[faridoonv1.GrantPermissionRequest]) (*connect.Response[emptypb.Empty], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if grantErr := s.store.GrantPermission(ctx, int(req.Msg.GroupId), int(req.Msg.PermissionId)); grantErr != nil {
		return nil, connect.NewError(connect.CodeInternal, grantErr)
	}
	s.audit(ctx, su, "group.grant_permission", "group", int(req.Msg.GroupId), detailf("permission_id=%d", req.Msg.PermissionId))
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *FaridoonServer) RevokePermission(ctx context.Context, req *connect.Request[faridoonv1.RevokePermissionRequest]) (*connect.Response[emptypb.Empty], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if lockErr := s.denyIfRevokeRemovesLastAdmin(ctx, int(req.Msg.GroupId), int(req.Msg.PermissionId)); lockErr != nil {
		return nil, lockErr
	}
	if revokeErr := s.store.RevokePermission(ctx, int(req.Msg.GroupId), int(req.Msg.PermissionId)); revokeErr != nil {
		return nil, connect.NewError(connect.CodeInternal, revokeErr)
	}
	s.audit(ctx, su, "group.revoke_permission", "group", int(req.Msg.GroupId), detailf("permission_id=%d", req.Msg.PermissionId))
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *FaridoonServer) denyIfRevokeRemovesLastAdmin(ctx context.Context, groupID, permissionID int) error {
	isSuper, err := s.isSuperuserPermission(ctx, permissionID)
	if err != nil {
		return err
	}
	if !isSuper {
		return nil
	}
	remaining, countErr := s.store.CountUsersWithPrivilegeExcludingGroup(ctx, "SUPERUSER", groupID)
	if countErr != nil {
		return connect.NewError(connect.CodeInternal, countErr)
	}
	if remaining < 1 {
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("cannot remove the last admin"))
	}
	return nil
}

func (s *FaridoonServer) isSuperuserPermission(ctx context.Context, permissionID int) (bool, error) {
	perm, err := s.store.FindPermission(ctx, permissionID)
	if err != nil {
		return false, connect.NewError(connect.CodeInternal, err)
	}
	return perm != nil && perm.Key == "SUPERUSER", nil
}

func (s *FaridoonServer) ListWebhooks(ctx context.Context, _ *connect.Request[faridoonv1.ListWebhooksRequest]) (*connect.Response[faridoonv1.ListWebhooksResponse], error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	rows, err := s.store.ListWebhooks(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &faridoonv1.ListWebhooksResponse{Events: webhook.SupportedEvents}
	for _, w := range rows {
		out.Webhooks = append(out.Webhooks, &faridoonv1.Webhook{
			Id: int32(w.ID), Url: w.URL, Event: w.Event, Enabled: w.Enabled,
			Created: w.Created, Updated: w.Updated,
		})
	}
	return connect.NewResponse(out), nil
}

func validateCreateWebhookInput(req *faridoonv1.CreateWebhookRequest) (url, event string, err error) {
	url, urlErr := webhook.NormalizeURL(req.Url)
	if urlErr != nil {
		return "", "", urlErr
	}
	event = req.Event
	if event == "" {
		event = "approval.requested"
	}
	event, eventErr := webhook.NormalizeEvent(event)
	if eventErr != nil {
		return "", "", eventErr
	}
	if strings.TrimSpace(req.Secret) == "" {
		return "", "", fmt.Errorf("secret required")
	}
	return url, event, nil
}

func (s *FaridoonServer) CreateWebhook(ctx context.Context, req *connect.Request[faridoonv1.CreateWebhookRequest]) (*connect.Response[faridoonv1.CreateWebhookResponse], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	url, event, valErr := validateCreateWebhookInput(req.Msg)
	if valErr != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, valErr)
	}
	id, err := s.store.CreateWebhook(ctx, url, req.Msg.Secret, event, req.Msg.Enabled)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	s.audit(ctx, su, "webhook.create", "webhook", id, event)
	w, _ := s.store.FindWebhook(ctx, id)
	return connect.NewResponse(&faridoonv1.CreateWebhookResponse{
		Webhook: &faridoonv1.Webhook{Id: int32(w.ID), Url: w.URL, Event: w.Event, Enabled: w.Enabled, Created: w.Created, Updated: w.Updated},
	}), nil
}

func applyWebhookURLField(fields map[string]any, rawURL string) error {
	if rawURL == "" {
		return nil
	}
	url, urlErr := webhook.NormalizeURL(rawURL)
	if urlErr != nil {
		return urlErr
	}
	fields["url"] = url
	return nil
}

func applyWebhookEventField(fields map[string]any, rawEvent string) error {
	if rawEvent == "" {
		return nil
	}
	event, eventErr := webhook.NormalizeEvent(rawEvent)
	if eventErr != nil {
		return eventErr
	}
	fields["event"] = event
	return nil
}

func webhookEnabledValue(enabled bool) int {
	if enabled {
		return 1
	}
	return 0
}

func buildWebhookFields(req *faridoonv1.UpdateWebhookRequest) (map[string]any, error) {
	fields := map[string]any{}
	if urlErr := applyWebhookURLField(fields, req.Url); urlErr != nil {
		return nil, urlErr
	}
	if req.Secret != "" {
		fields["secret"] = req.Secret
	}
	if eventErr := applyWebhookEventField(fields, req.Event); eventErr != nil {
		return nil, eventErr
	}
	fields["enabled"] = webhookEnabledValue(req.Enabled)
	return fields, nil
}

func toProtoWebhook(w *store.WebhookRow) *faridoonv1.Webhook {
	return &faridoonv1.Webhook{
		Id: int32(w.ID), Url: w.URL, Event: w.Event, Enabled: w.Enabled,
		Created: w.Created, Updated: w.Updated,
	}
}

func (s *FaridoonServer) findWebhookProto(ctx context.Context, id int) (*faridoonv1.Webhook, error) {
	w, findErr := s.store.FindWebhook(ctx, id)
	if findErr != nil || w == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("webhook not found"))
	}
	return toProtoWebhook(w), nil
}

func (s *FaridoonServer) UpdateWebhook(ctx context.Context, req *connect.Request[faridoonv1.UpdateWebhookRequest]) (*connect.Response[faridoonv1.Webhook], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	fields, buildErr := buildWebhookFields(req.Msg)
	if buildErr != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, buildErr)
	}
	if updErr := s.store.UpdateWebhook(ctx, int(req.Msg.Id), fields); updErr != nil {
		return nil, connect.NewError(connect.CodeInternal, updErr)
	}
	s.audit(ctx, su, "webhook.update", "webhook", int(req.Msg.Id), "")
	proto, findErr := s.findWebhookProto(ctx, int(req.Msg.Id))
	if findErr != nil {
		return nil, findErr
	}
	return connect.NewResponse(proto), nil
}

func (s *FaridoonServer) DeleteWebhook(ctx context.Context, req *connect.Request[faridoonv1.DeleteWebhookRequest]) (*connect.Response[emptypb.Empty], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if delErr := s.store.DeleteWebhook(ctx, int(req.Msg.Id)); delErr != nil {
		return nil, connect.NewError(connect.CodeInternal, delErr)
	}
	s.audit(ctx, su, "webhook.delete", "webhook", int(req.Msg.Id), "")
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func clampLogPageSize(pageSize int) int {
	if pageSize == 0 {
		return 25
	}
	return pageSize
}

func clampLogPage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func appendAuditLogs(out *faridoonv1.ListLogsResponse, rows []store.LogEntry) {
	for _, e := range rows {
		out.Logs = append(out.Logs, &faridoonv1.AuditLog{
			Id: int32(e.ID), Created: e.Created, ActorUserId: int32(e.ActorUserID),
			ActorUsername: e.ActorUsername, Action: e.Action, EntityType: e.EntityType,
			EntityId: int32(e.EntityID), Detail: e.Detail, Ip: e.IP,
		})
	}
}

func (s *FaridoonServer) ListLogs(ctx context.Context, req *connect.Request[faridoonv1.ListLogsRequest]) (*connect.Response[faridoonv1.ListLogsResponse], error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	page := int(req.Msg.Page)
	pageSize := clampLogPageSize(int(req.Msg.PageSize))
	rows, total, err := s.store.ListLogs(ctx, page, pageSize)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	page = clampLogPage(page)
	totalPages := (total + pageSize - 1) / pageSize
	out := &faridoonv1.ListLogsResponse{Page: int32(page), Total: int32(total), TotalPages: int32(totalPages)}
	appendAuditLogs(out, rows)
	return connect.NewResponse(out), nil
}
