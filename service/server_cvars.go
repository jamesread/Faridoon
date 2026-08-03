package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	faridoonv1 "faridoon/service/gen/faridoon/v1"
	"faridoon/service/internal/cvar"
	"faridoon/service/internal/store"
)

func toProtoCvar(row *store.CvarRow) *faridoonv1.Cvar {
	return &faridoonv1.Cvar{
		Key: row.Key, MainType: row.MainType,
		ValueInt: int32(row.ValueInt), ValueString: row.ValueString,
		Title: row.Title, Description: row.Description,
		Category: row.Category, Ordinal: int32(row.Ordinal),
	}
}

func appendProtoCvars(dst []*faridoonv1.Cvar, rows []store.CvarRow) []*faridoonv1.Cvar {
	for i := range rows {
		dst = append(dst, toProtoCvar(&rows[i]))
	}
	return dst
}

func ensureDefaultCvars(ctx context.Context, st store.Store, siteTitle string) error {
	for _, def := range cvar.Defaults(siteTitle) {
		if err := st.InsertCvarIfMissing(ctx, store.CvarRow{
			Key: def.Key, MainType: def.MainType,
			Title: def.Title, Description: def.Description,
			Category: def.Category, Ordinal: def.Ordinal,
			ValueInt: def.ValueInt, ValueString: def.ValueString,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *FaridoonServer) siteTitle(ctx context.Context) string {
	row, err := s.store.FindCvar(ctx, cvar.KeySiteTitle)
	if err != nil || row == nil || row.ValueString == "" {
		return s.cfg.SiteTitle
	}
	return row.ValueString
}

func (s *FaridoonServer) boolCvar(ctx context.Context, key string, fallback bool) bool {
	row, err := s.store.FindCvar(ctx, key)
	if err != nil || row == nil {
		return fallback
	}
	return row.ValueInt != 0
}

func (s *FaridoonServer) votingEnabled(ctx context.Context) bool {
	return s.boolCvar(ctx, cvar.KeyEnableVoting, false)
}

func (s *FaridoonServer) registrationEnabled(ctx context.Context) bool {
	return s.boolCvar(ctx, cvar.KeyEnableRegistration, true)
}

func (s *FaridoonServer) guestAddEnabled(ctx context.Context) bool {
	return s.boolCvar(ctx, cvar.KeyEnableGuestAdd, true)
}

func (s *FaridoonServer) guestAddRequireApproval(ctx context.Context) bool {
	return s.boolCvar(ctx, cvar.KeyGuestAddRequireApproval, true)
}

func (s *FaridoonServer) syntaxHighlightingEnabled(ctx context.Context) bool {
	return s.boolCvar(ctx, cvar.KeyEnableSyntaxHighlighting, false)
}

func (s *FaridoonServer) quotesPerPage(ctx context.Context) int {
	row, err := s.store.FindCvar(ctx, cvar.KeyQuotesPerPage)
	if err != nil || row == nil || row.ValueInt < 1 {
		return cvar.DefaultQuotesPerPage
	}
	if row.ValueInt > cvar.MaxQuotesPerPage {
		return cvar.MaxQuotesPerPage
	}
	return row.ValueInt
}

func (s *FaridoonServer) ListCvars(ctx context.Context, _ *connect.Request[faridoonv1.ListCvarsRequest]) (*connect.Response[faridoonv1.ListCvarsResponse], error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	rows, err := s.store.ListCvars(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&faridoonv1.ListCvarsResponse{
		Cvars: appendProtoCvars(nil, rows),
	}), nil
}

func validateCvarUpdate(row *store.CvarRow, valueInt int32, valueString string) (int, string, error) {
	switch row.MainType {
	case cvar.TypeString:
		return validateStringCvar(valueString)
	case cvar.TypeInt:
		return validateIntCvar(row.Key, valueInt)
	case cvar.TypeBool:
		return validateBoolCvar(valueInt), "", nil
	default:
		return 0, "", fmt.Errorf("unsupported cvar type")
	}
}

func validateStringCvar(valueString string) (int, string, error) {
	if valueString == "" {
		return 0, "", fmt.Errorf("value required")
	}
	if len(valueString) > 255 {
		return 0, "", fmt.Errorf("value too long")
	}
	return 0, valueString, nil
}

func validateIntCvar(key string, valueInt int32) (int, string, error) {
	v := int(valueInt)
	if key == cvar.KeyQuotesPerPage && (v < 1 || v > cvar.MaxQuotesPerPage) {
		return 0, "", fmt.Errorf("quotes_per_page must be between 1 and %d", cvar.MaxQuotesPerPage)
	}
	return v, "", nil
}

func validateBoolCvar(valueInt int32) int {
	if valueInt != 0 {
		return 1
	}
	return 0
}

func (s *FaridoonServer) UpdateCvar(ctx context.Context, req *connect.Request[faridoonv1.UpdateCvarRequest]) (*connect.Response[faridoonv1.Cvar], error) {
	su, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	row, findErr := s.requireCvar(ctx, req.Msg.Key)
	if findErr != nil {
		return nil, findErr
	}
	valueInt, valueString, valErr := validateCvarUpdate(row, req.Msg.ValueInt, req.Msg.ValueString)
	if valErr != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, valErr)
	}
	if updErr := s.store.UpdateCvar(ctx, row.Key, valueInt, valueString); updErr != nil {
		return nil, connect.NewError(connect.CodeInternal, updErr)
	}
	s.audit(ctx, su, "cvar.update", "cvar", 0, row.Key)
	updated, _ := s.store.FindCvar(ctx, row.Key)
	return connect.NewResponse(toProtoCvar(updated)), nil
}

func (s *FaridoonServer) requireCvar(ctx context.Context, key string) (*store.CvarRow, error) {
	row, err := s.store.FindCvar(ctx, key)
	if err != nil || row == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("cvar not found"))
	}
	return row, nil
}
