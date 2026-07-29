package main

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	faridoonv1 "faridoon/service/gen/faridoon/v1"
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

func (s *FaridoonServer) UpdateUser(ctx context.Context, req *connect.Request[faridoonv1.UpdateUserRequest]) (*connect.Response[faridoonv1.User], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if updErr := s.store.UpdateUserGroup(ctx, int(req.Msg.Id), int(req.Msg.GroupId)); updErr != nil {
		return nil, connect.NewError(connect.CodeInternal, updErr)
	}
	s.audit(ctx, su, "user.update_group", "user", int(req.Msg.Id), detailf("group_id=%d", req.Msg.GroupId))
	row, err := s.store.FindUser(ctx, int(req.Msg.Id))
	if err != nil || row == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
	}
	return connect.NewResponse(s.toProtoUserRow(ctx, row)), nil
}

func (s *FaridoonServer) DeleteUser(ctx context.Context, req *connect.Request[faridoonv1.DeleteUserRequest]) (*connect.Response[emptypb.Empty], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if delErr := s.store.DeleteUser(ctx, int(req.Msg.Id)); delErr != nil {
		return nil, connect.NewError(connect.CodeInternal, delErr)
	}
	s.audit(ctx, su, "user.delete", "user", int(req.Msg.Id), "")
	return connect.NewResponse(&emptypb.Empty{}), nil
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
	if revokeErr := s.store.RevokePermission(ctx, int(req.Msg.GroupId), int(req.Msg.PermissionId)); revokeErr != nil {
		return nil, connect.NewError(connect.CodeInternal, revokeErr)
	}
	s.audit(ctx, su, "group.revoke_permission", "group", int(req.Msg.GroupId), detailf("permission_id=%d", req.Msg.PermissionId))
	return connect.NewResponse(&emptypb.Empty{}), nil
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
