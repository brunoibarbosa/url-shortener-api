package command

import (
	"context"
	"time"

	session "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	user "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
)

type LoginGoogleCommand struct {
	Code      string
	UserAgent string
	IPAddress string
}

type LoginGoogleHandler struct {
	provider             session.OAuthProvider
	userRepo             user.UserRepository
	providerRepo         user.UserProviderRepository
	profileRepo          user.UserProfileRepository
	sessionRepo          session.SessionRepository
	tokenService         session.TokenService
	refreshTokenDuration time.Duration
	accessTokenDuration  time.Duration
}

func NewLoginGoogleHandler(
	provider session.OAuthProvider,
	userRepo user.UserRepository,
	providerRepo user.UserProviderRepository,
	profileRepo user.UserProfileRepository,
	sessionRepo session.SessionRepository,
	tokenService session.TokenService,
	refreshTokenDuration time.Duration,
	accessTokenDuration time.Duration,
) *LoginGoogleHandler {
	return &LoginGoogleHandler{
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

	s := &session.Session{
		UserAgent: cmd.UserAgent,
		IPAddress: cmd.IPAddress,
	}

	existingProvider, err := h.providerRepo.Find(ctx, provider, oauthUser.ID)
	if err != nil {
		return "", "", err
	}

	if existingProvider == nil {
		u, err := h.userRepo.GetByEmail(ctx, oauthUser.Email)
		if err != nil || u == nil {
			newUser := &user.User{
				Email:     oauthUser.Email,
				CreatedAt: time.Now(),
			}
			if err := h.userRepo.Create(ctx, newUser); err != nil {
				return "", "", err
			}
			u = newUser
		}
		s.UserID = u.ID

		pv := &user.UserProvider{
			UserID:     u.ID,
			Provider:   provider,
			ProviderID: oauthUser.ID,
		}
		if err := h.providerRepo.Create(ctx, u.ID, pv); err != nil {
			return "", "", err
		}

		if u.Profile == nil && (oauthUser.Name != "") {
			pf := &user.UserProfile{
				Name:      oauthUser.Name,
				AvatarURL: oauthUser.AvatarURL,
			}
			if err := h.profileRepo.Create(ctx, u.ID, pf); err != nil {
				return "", "", err
			}
		}
	} else {
		s.UserID = existingProvider.UserID
	}

	refreshToken := h.tokenService.GenerateRefreshToken()
	s.RefreshTokenHash = crypto.HashRefreshToken(refreshToken.String())

	expiresAt := time.Now().Add(h.refreshTokenDuration)
	s.ExpiresAt = &expiresAt

	if err := h.sessionRepo.Create(ctx, s); err != nil {
		return "", "", err
	}

	accessToken, err := h.tokenService.GenerateAccessToken(&session.TokenParams{
		UserID:    s.UserID,
		SessionID: s.ID,
		Duration:  h.accessTokenDuration,
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken.String(), nil
}
