package main

import (
	"context"
	"runtime"

	"connectrpc.com/connect"

	"faridoon/service/buildinfo"
	faridoonv1 "faridoon/service/gen/faridoon/v1"
)

func (s *FaridoonServer) GetDiagnostics(ctx context.Context, _ *connect.Request[faridoonv1.GetDiagnosticsRequest]) (*connect.Response[faridoonv1.GetDiagnosticsResponse], error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	out := &faridoonv1.GetDiagnosticsResponse{
		Version: buildinfo.Version, SiteTitle: s.siteTitle(ctx), GoVersion: runtime.Version(),
	}
	if mig, err := s.store.LatestMigration(ctx); err == nil {
		out.DatabaseMigration = mig
	}
	if n, err := s.store.UserCount(ctx); err == nil {
		out.UserCount = int32(n)
	}
	if n, err := s.store.CountPending(ctx); err == nil {
		out.PendingApprovals = int32(n)
	}
	if n, err := s.store.CountApproved(ctx); err == nil {
		out.ApprovedQuotes = int32(n)
	}
	return connect.NewResponse(out), nil
}
