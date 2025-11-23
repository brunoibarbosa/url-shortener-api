package command

import (
	"context"
	"time"

	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
)

type LogoutCommand struct {
	RefreshToken string
}

type LogoutHandler struct {
	sessionRepo   session_domain.SessionRepository
	blacklistRepo session_domain.BlacklistRepository
}

func NewLogoutHandler(
	sessionRepo session_domain.SessionRepository,
	blacklistRepo session_domain.BlacklistRepository,
) *LogoutHandler {
	return &LogoutHandler{
		sessionRepo,
		blacklistRepo,
	}
}

func (h *LogoutHandler) Handle(ctx context.Context, cmd LogoutCommand) error {
	hashed := crypto.HashRefreshToken(cmd.RefreshToken)

	s, err := h.sessionRepo.FindByRefreshToken(ctx, hashed)
	if err != nil || s == nil {
		return session_domain.ErrNotFound
	}

	if s.IsExpired() {
		return session_domain.ErrInvalidRefreshToken
	}

	if err := h.sessionRepo.Revoke(ctx, s.ID); err != nil {
		return session_domain.ErrRevokeFailed
	}

	remainder := time.Until(*s.ExpiresAt)
	_ = h.blacklistRepo.Revoke(ctx, hashed, remainder)

	return nil
}
