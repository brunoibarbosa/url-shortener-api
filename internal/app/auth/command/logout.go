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

	const maxRetries = 3
	var blErr error

	if err := h.blacklistRepo.Revoke(ctx, hashed, remainder); err != nil {
		for i := 0; i < maxRetries; i++ {
			blErr = h.blacklistRepo.Revoke(ctx, hashed, remainder)
			if blErr == nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}

		if blErr != nil {
			return session_domain.ErrRevokeFailed
		}
	}

	return nil
}
