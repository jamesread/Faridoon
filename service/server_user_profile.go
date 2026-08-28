package main

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	faridoonv1 "faridoon/service/gen/faridoon/v1"
	"faridoon/service/internal/authpass"
	"faridoon/service/internal/i18n"
)

func (s *FaridoonServer) GetUserPreferences(ctx context.Context, _ *connect.Request[faridoonv1.GetUserPreferencesRequest]) (*connect.Response[faridoonv1.GetUserPreferencesResponse], error) {
	su, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	prefs, err := s.store.GetUserPreferences(ctx, su.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&faridoonv1.GetUserPreferencesResponse{
		Language:           prefs.Language,
		AvailableLanguages: i18n.AvailableLanguageCodes(),
		SidebarEnabled:     prefs.SidebarEnabled,
	}), nil
}

func (s *FaridoonServer) SaveUserPreferences(ctx context.Context, req *connect.Request[faridoonv1.SaveUserPreferencesRequest]) (*connect.Response[faridoonv1.SaveUserPreferencesResponse], error) {
	su, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	language := strings.TrimSpace(req.Msg.Language)
	if !i18n.IsSupportedLanguage(language) {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unsupported language"))
	}
	if saveErr := s.store.SaveUserPreferences(ctx, su.ID, language, req.Msg.SidebarEnabled); saveErr != nil {
		return nil, connect.NewError(connect.CodeInternal, saveErr)
	}
	s.audit(ctx, su, "user.preferences.save", "user", su.ID, "")
	return connect.NewResponse(&faridoonv1.SaveUserPreferencesResponse{}), nil
}

func validateChangePasswordRequest(current, newPassword string) error {
	if current == "" || newPassword == "" {
		return fmt.Errorf("current and new password required")
	}
	if len(newPassword) < 8 {
		return fmt.Errorf("new password must be at least 8 characters")
	}
	return nil
}

func (s *FaridoonServer) verifyCurrentPassword(ctx context.Context, userID int, current string) error {
	row, err := s.store.FindUser(ctx, userID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if row == nil || !authpass.Verify(current, row.Password) {
		return connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("current password is incorrect"))
	}
	return nil
}

func (s *FaridoonServer) updateUserPassword(ctx context.Context, userID int, newPassword string) error {
	hash, err := authpass.Hash(newPassword)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if updErr := s.store.UpdatePassword(ctx, userID, hash); updErr != nil {
		return connect.NewError(connect.CodeInternal, updErr)
	}
	return nil
}

func (s *FaridoonServer) ChangePassword(ctx context.Context, req *connect.Request[faridoonv1.ChangePasswordRequest]) (*connect.Response[faridoonv1.ChangePasswordResponse], error) {
	su, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	current := req.Msg.CurrentPassword
	newPassword := req.Msg.NewPassword
	if validateErr := validateChangePasswordRequest(current, newPassword); validateErr != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, validateErr)
	}
	if verifyErr := s.verifyCurrentPassword(ctx, su.ID, current); verifyErr != nil {
		return nil, verifyErr
	}
	if updErr := s.updateUserPassword(ctx, su.ID, newPassword); updErr != nil {
		return nil, updErr
	}
	s.audit(ctx, su, "user.password.change", "user", su.ID, "")
	return connect.NewResponse(&faridoonv1.ChangePasswordResponse{}), nil
}
