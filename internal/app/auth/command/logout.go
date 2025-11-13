package command

import (
	"context"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
)

type LogoutCommand struct {
	RefreshToken string
}

type LogoutHandler struct {
	sessionRepo   session.SessionRepository
	blacklistRepo session.BlacklistRepository
}

func NewLogoutHandler(
	sessionRepo session.SessionRepository,
	blacklistRepo session.BlacklistRepository,
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
		return session.ErrNotFound
	}

	if s.IsExpired() {
		return session.ErrInvalidRefreshToken
	}

	if err := h.sessionRepo.Revoke(ctx, s.ID); err != nil {
		return session.ErrRevokeFailed
	}

	remainder := time.Until(*s.ExpiresAt)
	_ = h.blacklistRepo.Revoke(ctx, hashed, remainder)

	return nil
}
