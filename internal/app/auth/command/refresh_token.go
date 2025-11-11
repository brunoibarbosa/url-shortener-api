package command

import (
	"context"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
)

type RefreshTokenCommand struct {
	RefreshToken string
	UserAgent    string
	IPAddress    string
}

type RefreshTokenHandler struct {
	sessionRepo   session.SessionRepository
	blacklistRepo session.BlacklistRepository
	tokenService  session.TokenService
}

func NewRefreshTokenHandler(
	sessionRepo session.SessionRepository,
	blacklistRepo session.BlacklistRepository,
	tokenService session.TokenService,
) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		sessionRepo,
		blacklistRepo,
		tokenService,
	}
}

func (h *RefreshTokenHandler) Handle(ctx context.Context, cmd RefreshTokenCommand) (string, string, error) {
	hashed := crypto.HashRefreshToken(cmd.RefreshToken)

	s, err := h.sessionRepo.FindByRefreshToken(ctx, hashed)
	if err != nil || s == nil || s.IsExpired() {
		return "", "", session.ErrInvalidRefreshToken
	}

	revoked, err := h.blacklistRepo.IsRevoked(ctx, hashed)
	if err != nil {
		return "", "", err
	}
	if revoked {
		return "", "", session.ErrInvalidRefreshToken
	}

	_ = h.sessionRepo.Revoke(ctx, s.ID)

	remainder := time.Until(*s.ExpiresAt)
	_ = h.blacklistRepo.Revoke(ctx, hashed, remainder)

	refreshToken := h.tokenService.GenerateRefreshToken()
	refreshHash := crypto.HashRefreshToken(refreshToken.String())

	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	sess := &session.Session{
		UserID:           s.UserID,
		RefreshTokenHash: refreshHash,
		UserAgent:        cmd.UserAgent,
		IPAddress:        cmd.IPAddress,
		ExpiresAt:        &expiresAt,
	}
	if err := h.sessionRepo.Create(ctx, sess); err != nil {
		return "", "", err
	}

	accessToken, err := h.tokenService.GenerateAccessToken(&session.TokenParams{
		UserID:    s.UserID,
		SessionID: sess.ID,
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken.String(), nil
}
