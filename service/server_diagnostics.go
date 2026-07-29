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
	s.fillDiagnosticsCounts(ctx, out)
	return connect.NewResponse(out), nil
}

func (s *FaridoonServer) fillDiagnosticsCounts(ctx context.Context, out *faridoonv1.GetDiagnosticsResponse) {
	if mig, err := s.store.LatestMigration(ctx); err == nil {
		out.DatabaseMigration = mig
	}
	out.UserCount = int32OrZero(s.store.UserCount(ctx))
	out.PendingApprovals = int32OrZero(s.store.CountPending(ctx))
	out.ApprovedQuotes = int32OrZero(s.store.CountApproved(ctx))
}

func int32OrZero(n int, err error) int32 {
	if err != nil {
		return 0
	}
	return int32(n)
}
