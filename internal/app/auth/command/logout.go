package command

import (
	"context"
	"errors"
	"time"

	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
)

type LogoutCommand struct {
	RefreshToken string
}

type LogoutHandler struct {
	sessionRepo      session_domain.SessionRepository
	blacklistRepo    session_domain.BlacklistRepository
	sessionEncrypter session_domain.SessionEncrypter
}

func NewLogoutHandler(
	sessionRepo session_domain.SessionRepository,
	blacklistRepo session_domain.BlacklistRepository,
	sessionEncrypter session_domain.SessionEncrypter,
) *LogoutHandler {
	return &LogoutHandler{
		sessionRepo,
		blacklistRepo,
		sessionEncrypter,
	}
}

func (h *LogoutHandler) Handle(ctx context.Context, cmd LogoutCommand) error {
	hashed := h.sessionEncrypter.HashRefreshToken(cmd.RefreshToken)

	s, err := h.sessionRepo.FindByRefreshToken(ctx, hashed)
	if err != nil {
		switch {
		case errors.Is(err, session_domain.ErrNotFound):
			return session_domain.ErrInvalidRefreshToken
		default:
			return err
		}
	}

	if s == nil {
		return session_domain.ErrInvalidRefreshToken
	}

	if s.IsExpired() {
		return session_domain.ErrInvalidRefreshToken
	}

	if err := h.sessionRepo.Revoke(ctx, s.ID); err != nil {
		return session_domain.ErrRevokeFailed
	}

	remainder := time.Until(*s.ExpiresAt)
	if err := h.blacklistRepo.Revoke(ctx, hashed, remainder); err != nil {
		return session_domain.ErrRevokeFailed
	}

	return nil
}
