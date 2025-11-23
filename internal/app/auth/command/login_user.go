package command

import (
	"context"
	"time"

	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	user_domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
	"github.com/jackc/pgx/v5"
)

type LoginUserCommand struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

type LoginUserHandler struct {
	db                   *pg.Postgres
	providerRepo         user_domain.UserProviderRepository
	sessionRepo          session_domain.SessionRepository
	tokenService         session_domain.TokenService
	refreshTokenDuration time.Duration
	accessTokenDuration  time.Duration
}

func NewLoginUserHandler(
	db *pg.Postgres,
	providerRepo user_domain.UserProviderRepository,
	sessionRepo session_domain.SessionRepository,
	tokenService session_domain.TokenService,
	refreshTokenDuration time.Duration,
	accessTokenDuration time.Duration,
) *LoginUserHandler {
	return &LoginUserHandler{
		db,
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
		return "", "", user_domain.ErrInvalidCredentials
	}

	if !crypto.CheckPassword(cmd.Password, *u.PasswordHash) {
		return "", "", user_domain.ErrInvalidCredentials
	}

	var sess *session_domain.Session
	var refreshToken string

	err = h.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		sessionRepo := h.sessionRepo.WithTx(tx)

		refreshTokenObj := h.tokenService.GenerateRefreshToken()
		refreshToken = refreshTokenObj.String()
		refreshHash := crypto.HashRefreshToken(refreshToken)

		expiresAt := time.Now().Add(h.refreshTokenDuration)

		sess = &session_domain.Session{
			UserID:           u.UserID,
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
		UserID:    u.UserID,
		SessionID: sess.ID,
		Duration:  h.accessTokenDuration,
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
