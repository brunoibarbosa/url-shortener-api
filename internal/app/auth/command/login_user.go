package command

import (
	"context"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
)

type LoginUserCommand struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

type LoginUserHandler struct {
	providerRepo         user.UserProviderRepository
	sessionRepo          session.SessionRepository
	tokenService         session.TokenService
	refreshTokenDuration time.Duration
	accessTokenDuration  time.Duration
}

func NewLoginUserHandler(
	providerRepo user.UserProviderRepository,
	sessionRepo session.SessionRepository,
	tokenService session.TokenService,
	refreshTokenDuration time.Duration,
	accessTokenDuration time.Duration,
) *LoginUserHandler {
	return &LoginUserHandler{
		providerRepo,
		sessionRepo,
		tokenService,
		refreshTokenDuration,
		accessTokenDuration,
	}
}

func (h *LoginUserHandler) Handle(ctx context.Context, cmd LoginUserCommand) (string, string, error) {
	u, err := h.providerRepo.Find(ctx, "password", cmd.Email)
	if err != nil {
		return "", "", err
	}

	if u == nil {
		return "", "", user.ErrInvalidCredentials
	}

	if !crypto.CheckPassword(cmd.Password, *u.PasswordHash) {
		return "", "", user.ErrInvalidCredentials
	}

	refreshToken := h.tokenService.GenerateRefreshToken()
	refreshHash := crypto.HashRefreshToken(refreshToken.String())

	expiresAt := time.Now().Add(h.refreshTokenDuration)

	sess := &session.Session{
		UserID:           u.UserID,
		RefreshTokenHash: refreshHash,
		UserAgent:        cmd.UserAgent,
		IPAddress:        cmd.IPAddress,
		ExpiresAt:        &expiresAt,
	}
	if err := h.sessionRepo.Create(ctx, sess); err != nil {
		return "", "", err
	}

	accessToken, err := h.tokenService.GenerateAccessToken(&session.TokenParams{
		UserID:    u.UserID,
		SessionID: sess.ID,
		Duration:  h.accessTokenDuration,
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken.String(), nil
}
