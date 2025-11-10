package command

import (
	"context"
	"time"

	session "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	user "github.com/brunoibarbosa/url-shortener/internal/domain/user"
)

type LoginGoogleCommand struct {
	Provider      string
	ProviderID    string
	Email         string
	EmailVerified bool
	Name          string
	AvatarURL     *string
	AccessToken   string
	RefreshToken  string
}

type LoginGoogleHandler struct {
	provider     session.OAuthProvider
	userRepo     user.UserRepository
	providerRepo user.UserProviderRepository
	profileRepo  user.UserProfileRepository
	tokenService session.TokenService
}

func NewLoginGoogleHandler(provider session.OAuthProvider, userRepo user.UserRepository, providerRepo user.UserProviderRepository, profileRepo user.UserProfileRepository, tokenService session.TokenService) *LoginGoogleHandler {
	return &LoginGoogleHandler{
		provider,
		userRepo,
		providerRepo,
		profileRepo,
		tokenService,
	}
}

func (h *LoginGoogleHandler) Handle(ctx context.Context, code string) (string, error) {
	provider := "google"
	oauthUser, err := h.provider.ExchangeCode(ctx, code)
	if err != nil {
		return "", err
	}

	existingProvider, err := h.providerRepo.Find(ctx, provider, oauthUser.ID)
	if err != nil {
		return "", err
	}

	if existingProvider != nil {
		tp := &session.TokenParams{
			UserID: existingProvider.UserID,
		}
		return h.tokenService.GenerateAccessToken(tp)
	}

	u, err := h.userRepo.GetByEmail(ctx, oauthUser.Email)
	if err != nil || u == nil {
		newUser := &user.User{
			Email:     oauthUser.Email,
			CreatedAt: time.Now(),
		}
		if err := h.userRepo.Create(ctx, newUser); err != nil {
			return "", err
		}
		u = newUser
	}

	pv := &user.UserProvider{
		UserID:     u.ID,
		Provider:   provider,
		ProviderID: oauthUser.ID,
	}
	if err := h.providerRepo.Create(ctx, u.ID, pv); err != nil {
		return "", err
	}

	if u.Profile == nil && (oauthUser.Name != "") {
		pf := &user.UserProfile{
			Name:      oauthUser.Name,
			AvatarURL: oauthUser.AvatarURL,
		}
		if err := h.profileRepo.Create(ctx, u.ID, pf); err != nil {
			return "", err
		}
	}

	token, err := h.tokenService.GenerateAccessToken(&session.TokenParams{
		UserID: u.ID,
	})
	if err != nil {
		return "", err
	}

	return token, nil
}
