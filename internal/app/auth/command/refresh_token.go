package command

import (
	"context"
	"time"

	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
	"github.com/jackc/pgx/v5"
)

type RefreshTokenCommand struct {
	RefreshToken string
	UserAgent    string
	IPAddress    string
}

type RefreshTokenHandler struct {
	db                   *pg.Postgres
	sessionRepo          session_domain.SessionRepository
	blacklistRepo        session_domain.BlacklistRepository
	tokenService         session_domain.TokenService
	refreshTokenDuration time.Duration
	accessTokenDuration  time.Duration
}

func NewRefreshTokenHandler(
	db *pg.Postgres,
	sessionRepo session_domain.SessionRepository,
	blacklistRepo session_domain.BlacklistRepository,
	tokenService session_domain.TokenService,
	refreshTokenDuration time.Duration,
	accessTokenDuration time.Duration,
) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		db,
		sessionRepo,
		blacklistRepo,
		tokenService,
		refreshTokenDuration,
		accessTokenDuration,
	}
}

func (h *RefreshTokenHandler) Handle(ctx context.Context, cmd RefreshTokenCommand) (string, string, error) {
	hashed := crypto.HashRefreshToken(cmd.RefreshToken)

	s, err := h.sessionRepo.FindByRefreshToken(ctx, hashed)
	if err != nil || s == nil || s.IsExpired() {
		return "", "", session_domain.ErrInvalidRefreshToken
	}

	revoked, err := h.blacklistRepo.IsRevoked(ctx, hashed)
	if err != nil {
		return "", "", err
	}
	if revoked {
		return "", "", session_domain.ErrInvalidRefreshToken
	}

	var sess *session_domain.Session
	var refreshToken string

	err = h.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		sessionRepo := h.sessionRepo.WithTx(tx)

		if err := sessionRepo.Revoke(ctx, s.ID); err != nil {
			return err
		}

		remainder := time.Until(*s.ExpiresAt)
		_ = h.blacklistRepo.Revoke(ctx, hashed, remainder)

		refreshTokenObj := h.tokenService.GenerateRefreshToken()
		refreshToken = refreshTokenObj.String()
		refreshHash := crypto.HashRefreshToken(refreshToken)

		expiresAt := time.Now().Add(h.refreshTokenDuration)

		sess = &session_domain.Session{
			UserID:           s.UserID,
			RefreshTokenHash: refreshHash,
			UserAgent:        cmd.UserAgent,
			IPAddress:        cmd.IPAddress,
			ExpiresAt:        &expiresAt,
		}
		return sessionRepo.Create(ctx, sess)
	})
	if err != nil {
		return "", "", err
	}

	accessToken, err := h.tokenService.GenerateAccessToken(&session_domain.TokenParams{
		UserID:    s.UserID,
		SessionID: sess.ID,
		Duration:  h.accessTokenDuration,
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
