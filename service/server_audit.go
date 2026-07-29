package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"faridoon/service/internal/store"
)

func clientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// audit writes a best-effort audit log row; failures are logged and ignored.
func (s *FaridoonServer) audit(ctx context.Context, actor *sessionUser, action, entityType string, entityID int, detail string) {
	entry := store.LogEntry{
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Detail:     detail,
		IP:         clientIP(s.httpReq(ctx)),
	}
	if actor != nil {
		entry.ActorUserID = actor.ID
		entry.ActorUsername = actor.Username
	}
	if err := s.store.InsertLog(ctx, entry); err != nil {
		logrus.WithError(err).Warn("audit log insert failed")
	}
}

func detailf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
