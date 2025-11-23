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

type LoginGoogleCommand struct {
	Code      string
	UserAgent string
	IPAddress string
}

type LoginGoogleHandler struct {
	db                   *pg.Postgres
	provider             session_domain.OAuthProvider
	userRepo             user_domain.UserRepository
	providerRepo         user_domain.UserProviderRepository
	profileRepo          user_domain.UserProfileRepository
	sessionRepo          session_domain.SessionRepository
	tokenService         session_domain.TokenService
	refreshTokenDuration time.Duration
	accessTokenDuration  time.Duration
}

func NewLoginGoogleHandler(
	db *pg.Postgres,
	provider session_domain.OAuthProvider,
	userRepo user_domain.UserRepository,
	providerRepo user_domain.UserProviderRepository,
	profileRepo user_domain.UserProfileRepository,
	sessionRepo session_domain.SessionRepository,
	tokenService session_domain.TokenService,
	refreshTokenDuration time.Duration,
	accessTokenDuration time.Duration,
) *LoginGoogleHandler {
	return &LoginGoogleHandler{
		db,
		provider,
		userRepo,
		providerRepo,
		profileRepo,
		sessionRepo,
		tokenService,
		refreshTokenDuration,
		accessTokenDuration,
	}
}

func (h *LoginGoogleHandler) Handle(ctx context.Context, cmd LoginGoogleCommand) (string, string, error) {
	provider := "google"
	oauthUser, err := h.provider.ExchangeCode(ctx, cmd.Code)
	if err != nil {
		return "", "", err
	}

	existingProvider, err := h.providerRepo.Find(ctx, provider, oauthUser.ID)
	if err != nil {
		return "", "", err
	}

	var s *session_domain.Session
	var refreshToken string

	if existingProvider == nil {
		err = h.db.WithTransaction(ctx, func(tx pgx.Tx) error {
			userRepo := h.userRepo.WithTx(tx)
			providerRepo := h.providerRepo.WithTx(tx)
			profileRepo := h.profileRepo.WithTx(tx)
			sessionRepo := h.sessionRepo.WithTx(tx)

			u, err := userRepo.GetByEmail(ctx, oauthUser.Email)
			if err != nil || u == nil {
				newUser := &user_domain.User{
					Email:     oauthUser.Email,
					CreatedAt: time.Now(),
				}
				if err := userRepo.Create(ctx, newUser); err != nil {
					return err
				}
				u = newUser
			}

			pv := &user_domain.UserProvider{
				UserID:     u.ID,
				Provider:   provider,
				ProviderID: oauthUser.ID,
			}
			if err := providerRepo.Create(ctx, u.ID, pv); err != nil {
				return err
			}

			if u.Profile == nil && (oauthUser.Name != "") {
				pf := &user_domain.UserProfile{
					Name:      oauthUser.Name,
					AvatarURL: oauthUser.AvatarURL,
				}
				if err := profileRepo.Create(ctx, u.ID, pf); err != nil {
					return err
				}
			}

			refreshTokenObj := h.tokenService.GenerateRefreshToken()
			refreshToken = refreshTokenObj.String()
			expiresAt := time.Now().Add(h.refreshTokenDuration)

			s = &session_domain.Session{
				UserID:           u.ID,
				UserAgent:        cmd.UserAgent,
				IPAddress:        cmd.IPAddress,
				RefreshTokenHash: crypto.HashRefreshToken(refreshToken),
				ExpiresAt:        &expiresAt,
			}
			return sessionRepo.Create(ctx, s)
		})
		if err != nil {
			return "", "", err
		}
	} else {
		refreshTokenObj := h.tokenService.GenerateRefreshToken()
		refreshToken = refreshTokenObj.String()
		expiresAt := time.Now().Add(h.refreshTokenDuration)

		s = &session_domain.Session{
			UserID:           existingProvider.UserID,
			UserAgent:        cmd.UserAgent,
			IPAddress:        cmd.IPAddress,
			RefreshTokenHash: crypto.HashRefreshToken(refreshToken),
			ExpiresAt:        &expiresAt,
		}
		if err := h.sessionRepo.Create(ctx, s); err != nil {
			return "", "", err
		}
	}

	accessToken, err := h.tokenService.GenerateAccessToken(&session_domain.TokenParams{
		UserID:    s.UserID,
		SessionID: s.ID,
		Duration:  h.accessTokenDuration,
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
