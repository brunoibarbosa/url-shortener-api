package command

import (
	"context"
	"errors"
	"time"

	bd_domain "github.com/brunoibarbosa/url-shortener/internal/domain/bd"
	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	user_domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
)

type LoginUserCommand struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

type LoginUserHandler struct {
	tx                   bd_domain.TransactionManager
	providerRepo         user_domain.UserProviderRepository
	sessionRepo          session_domain.SessionRepository
	tokenService         session_domain.TokenService
	passwordEncrypter    user_domain.UserPasswordEncrypter
	sessionEncrypter     session_domain.SessionEncrypter
	refreshTokenDuration time.Duration
	accessTokenDuration  time.Duration
}

func NewLoginUserHandler(
	tx bd_domain.TransactionManager,
	providerRepo user_domain.UserProviderRepository,
	sessionRepo session_domain.SessionRepository,
	tokenService session_domain.TokenService,
	passwordEncrypter user_domain.UserPasswordEncrypter,
	sessionEncrypter session_domain.SessionEncrypter,
	refreshTokenDuration time.Duration,
	accessTokenDuration time.Duration,
) *LoginUserHandler {
	return &LoginUserHandler{
		tx,
		providerRepo,
		sessionRepo,
		tokenService,
		passwordEncrypter,
		sessionEncrypter,
		refreshTokenDuration,
		accessTokenDuration,
	}
}

func (h *LoginUserHandler) Handle(ctx context.Context, cmd LoginUserCommand) (string, string, error) {
	if cmd.Email == "" || cmd.Password == "" {
		return "", "", user_domain.ErrInvalidCredentials
	}

	u, err := h.providerRepo.Find(ctx, user_domain.ProviderPassword, cmd.Email)
	if err != nil {
		switch {
		case errors.Is(err, user_domain.ErrNotFound):
			return "", "", user_domain.ErrInvalidCredentials
		default:
			return "", "", err
		}
	}

	if u == nil {
		return "", "", user_domain.ErrInvalidCredentials
	}

	if !h.passwordEncrypter.CheckPassword(*u.PasswordHash, cmd.Password) {
		return "", "", user_domain.ErrInvalidCredentials
	}

	var sess *session_domain.Session
	var refreshToken string

	err = h.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		refreshTokenObj := h.tokenService.GenerateRefreshToken()
		refreshToken = refreshTokenObj.String()
		refreshHash := h.sessionEncrypter.HashRefreshToken(refreshToken)

		expiresAt := time.Now().Add(h.refreshTokenDuration)

		sess = &session_domain.Session{
			UserID:           u.UserID,
			RefreshTokenHash: refreshHash,
			UserAgent:        cmd.UserAgent,
			IPAddress:        cmd.IPAddress,
			ExpiresAt:        &expiresAt,
		}
		return h.sessionRepo.Create(txCtx, sess)
	})
	if err != nil {
		return "", "", err
	}

	accessToken, err := h.tokenService.GenerateAccessToken(&session_domain.TokenParams{
		UserID:    u.UserID,
		SessionID: sess.ID,
		Duration:  h.accessTokenDuration,
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
