package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	faridoonv1 "faridoon/service/gen/faridoon/v1"
	"faridoon/service/internal/headerlink"
	"faridoon/service/internal/store"
)

func toProtoHeaderLink(row *store.HeaderLinkRow) *faridoonv1.HeaderLink {
	return &faridoonv1.HeaderLink{
		Id: int32(row.ID), Title: row.Title, Url: row.URL, SortOrder: int32(row.SortOrder),
		Enabled: row.Enabled, OpenInNewTab: row.OpenInNewTab, Created: row.Created, Updated: row.Updated,
	}
}

func appendProtoHeaderLinks(dst []*faridoonv1.HeaderLink, rows []store.HeaderLinkRow) []*faridoonv1.HeaderLink {
	for i := range rows {
		dst = append(dst, toProtoHeaderLink(&rows[i]))
	}
	return dst
}

func (s *FaridoonServer) loadEnabledHeaderLinks(ctx context.Context) []*faridoonv1.HeaderLink {
	rows, err := s.store.ListEnabledHeaderLinks(ctx)
	if err != nil {
		return nil
	}
	return appendProtoHeaderLinks(nil, rows)
}

func validateHeaderLinkFields(title, rawURL string) (string, string, error) {
	title, titleErr := headerlink.NormalizeTitle(title)
	if titleErr != nil {
		return "", "", titleErr
	}
	url, urlErr := headerlink.NormalizeURL(rawURL)
	if urlErr != nil {
		return "", "", urlErr
	}
	return title, url, nil
}

func (s *FaridoonServer) findHeaderLinkProto(ctx context.Context, id int) (*faridoonv1.HeaderLink, error) {
	row, err := s.store.FindHeaderLink(ctx, id)
	if err != nil || row == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("header link not found"))
	}
	return toProtoHeaderLink(row), nil
}

func (s *FaridoonServer) ListHeaderLinks(ctx context.Context, _ *connect.Request[faridoonv1.ListHeaderLinksRequest]) (*connect.Response[faridoonv1.ListHeaderLinksResponse], error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	rows, err := s.store.ListHeaderLinks(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&faridoonv1.ListHeaderLinksResponse{
		Links: appendProtoHeaderLinks(nil, rows),
	}), nil
}

func (s *FaridoonServer) CreateHeaderLink(ctx context.Context, req *connect.Request[faridoonv1.CreateHeaderLinkRequest]) (*connect.Response[faridoonv1.CreateHeaderLinkResponse], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	title, url, valErr := validateHeaderLinkFields(req.Msg.Title, req.Msg.Url)
	if valErr != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, valErr)
	}
	id, err := s.store.CreateHeaderLink(ctx, title, url, int(req.Msg.SortOrder), req.Msg.Enabled, req.Msg.OpenInNewTab)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	s.audit(ctx, su, "header_link.create", "header_link", id, title)
	proto, findErr := s.findHeaderLinkProto(ctx, id)
	if findErr != nil {
		return nil, findErr
	}
	return connect.NewResponse(&faridoonv1.CreateHeaderLinkResponse{Link: proto}), nil
}

func (s *FaridoonServer) UpdateHeaderLink(ctx context.Context, req *connect.Request[faridoonv1.UpdateHeaderLinkRequest]) (*connect.Response[faridoonv1.HeaderLink], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	title, url, valErr := validateHeaderLinkFields(req.Msg.Title, req.Msg.Url)
	if valErr != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, valErr)
	}
	if updErr := s.store.UpdateHeaderLink(ctx, int(req.Msg.Id), title, url, int(req.Msg.SortOrder), req.Msg.Enabled, req.Msg.OpenInNewTab); updErr != nil {
		return nil, connect.NewError(connect.CodeInternal, updErr)
	}
	s.audit(ctx, su, "header_link.update", "header_link", int(req.Msg.Id), title)
	proto, findErr := s.findHeaderLinkProto(ctx, int(req.Msg.Id))
	if findErr != nil {
		return nil, findErr
	}
	return connect.NewResponse(proto), nil
}

func (s *FaridoonServer) DeleteHeaderLink(ctx context.Context, req *connect.Request[faridoonv1.DeleteHeaderLinkRequest]) (*connect.Response[emptypb.Empty], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if delErr := s.store.DeleteHeaderLink(ctx, int(req.Msg.Id)); delErr != nil {
		return nil, connect.NewError(connect.CodeInternal, delErr)
	}
	s.audit(ctx, su, "header_link.delete", "header_link", int(req.Msg.Id), "")
	return connect.NewResponse(&emptypb.Empty{}), nil
}
