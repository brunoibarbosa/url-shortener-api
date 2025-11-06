package command

import (
	"context"
	"fmt"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
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
	provider     domain.OAuthProvider
	userRepo     domain.UserRepository
	providerRepo domain.UserProviderRepository
	profileRepo  domain.UserProfileRepository
	tokenService domain.TokenService
}

func NewLoginGoogleHandler(provider domain.OAuthProvider, userRepo domain.UserRepository, providerRepo domain.UserProviderRepository, profileRepo domain.UserProfileRepository, tokenService domain.TokenService) *LoginGoogleHandler {
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
	user, err := h.provider.ExchangeCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("error exchanging code: %w", err)
	}

	existingProvider, err := h.providerRepo.Find(ctx, provider, user.ID)
	if err != nil {
		return "", err
	}

	if existingProvider != nil {
		tp := &domain.TokenParams{
			UserID: existingProvider.UserID,
		}
		return h.tokenService.GenerateAccessToken(tp)
	}

	u, err := h.userRepo.GetByEmail(ctx, user.Email)
	if err != nil || u == nil {
		newUser := &domain.User{
			Email:     user.Email,
			CreatedAt: time.Now(),
		}
		if err := h.userRepo.Create(ctx, newUser); err != nil {
			return "", fmt.Errorf("error creating user: %w", err)
		}
		u = newUser
	}

	pv := &domain.UserProvider{
		UserID:     u.ID,
		Provider:   provider,
		ProviderID: user.ID,
	}
	if err := h.providerRepo.Create(ctx, u.ID, pv); err != nil {
		return "", fmt.Errorf("error linking provider: %w", err)
	}

	if u.Profile == nil && (user.Name != "") {
		pf := &domain.UserProfile{
			Name:      user.Name,
			AvatarURL: user.AvatarURL,
		}
		if err := h.profileRepo.Create(ctx, u.ID, pf); err != nil {
			return "", fmt.Errorf("error creating user: %w", err)
		}
	}

	token, err := h.tokenService.GenerateAccessToken(&domain.TokenParams{
		UserID: u.ID,
	})
	if err != nil {
		return "", fmt.Errorf("token generation failed: %w", err)
	}

	return token, nil
}
