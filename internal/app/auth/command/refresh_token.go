package command

import (
	"context"
	"errors"
	"time"

	bd_domain "github.com/brunoibarbosa/url-shortener/internal/domain/bd"
	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
)

type RefreshTokenCommand struct {
	RefreshToken string
	UserAgent    string
	IPAddress    string
}

type RefreshTokenHandler struct {
	tx                   bd_domain.TransactionManager
	sessionRepo          session_domain.SessionRepository
	blacklistRepo        session_domain.BlacklistRepository
	tokenService         session_domain.TokenService
	sessionEncrypter     session_domain.SessionEncrypter
	refreshTokenDuration time.Duration
	accessTokenDuration  time.Duration
}

type RefreshTokenResponse struct {
	AccessToken  string
	RefreshToken string
}

func NewRefreshTokenHandler(
	tx bd_domain.TransactionManager,
	sessionRepo session_domain.SessionRepository,
	blacklistRepo session_domain.BlacklistRepository,
	tokenService session_domain.TokenService,
	sessionEncrypter session_domain.SessionEncrypter,
	refreshTokenDuration time.Duration,
	accessTokenDuration time.Duration,
) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		tx,
		sessionRepo,
		blacklistRepo,
		tokenService,
		sessionEncrypter,
		refreshTokenDuration,
		accessTokenDuration,
	}
}

func (h *RefreshTokenHandler) Handle(ctx context.Context, cmd RefreshTokenCommand) (RefreshTokenResponse, error) {
	if cmd.RefreshToken == "" {
		return RefreshTokenResponse{}, session_domain.ErrInvalidRefreshToken
	}

	hashed := h.sessionEncrypter.HashRefreshToken(cmd.RefreshToken)

	s, err := h.sessionRepo.FindByRefreshToken(ctx, hashed)
	if err != nil {
		switch {
		case errors.Is(err, session_domain.ErrNotFound):
			return RefreshTokenResponse{}, session_domain.ErrInvalidRefreshToken
		default:
			return RefreshTokenResponse{}, err
		}
	}
	if s == nil || s.IsExpired() {
		return RefreshTokenResponse{}, session_domain.ErrInvalidRefreshToken
	}

	revoked, err := h.blacklistRepo.IsRevoked(ctx, hashed)
	if err != nil {
		return RefreshTokenResponse{}, err
	}
	if revoked {
		return RefreshTokenResponse{}, session_domain.ErrInvalidRefreshToken
	}

	var sess *session_domain.Session
	var refreshToken string

	err = h.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := h.sessionRepo.Revoke(txCtx, s.ID); err != nil {
			return err
		}

		remainder := time.Until(*s.ExpiresAt)
		if err := h.blacklistRepo.Revoke(txCtx, hashed, remainder); err != nil {
			return err
		}

		refreshTokenObj := h.tokenService.GenerateRefreshToken()
		refreshToken = refreshTokenObj.String()
		refreshHash := h.sessionEncrypter.HashRefreshToken(refreshToken)

		expiresAt := time.Now().Add(h.refreshTokenDuration)

		sess = &session_domain.Session{
			UserID:           s.UserID,
			RefreshTokenHash: refreshHash,
			UserAgent:        cmd.UserAgent,
			IPAddress:        cmd.IPAddress,
			ExpiresAt:        &expiresAt,
		}
		return h.sessionRepo.Create(txCtx, sess)
	})
	if err != nil {
		return RefreshTokenResponse{}, err
	}

	accessToken, err := h.tokenService.GenerateAccessToken(&session_domain.TokenParams{
		UserID:    s.UserID,
		SessionID: sess.ID,
		Duration:  h.accessTokenDuration,
	})
	if err != nil {
		return RefreshTokenResponse{}, err
	}

	return RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
