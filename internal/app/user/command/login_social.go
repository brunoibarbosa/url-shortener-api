package command

import (
	"context"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
)

type LoginSocialCommand struct {
	Email        string
	Provider     string
	ProviderID   string
	AccessToken  string
	RefreshToken string
}

type LoginSocialHandler struct {
	repo             domain.UserRepository
	encryptSecretKey string
}

func NewLoginSocialHandler(repo domain.UserRepository, secretKey string) *LoginSocialHandler {
	return &LoginSocialHandler{
		repo:             repo,
		encryptSecretKey: secretKey,
	}
}

func (h *LoginSocialHandler) Handle(ctx context.Context, cmd LoginSocialCommand) (string, error) {
	u, err := h.repo.GetByProvider(ctx, cmd.Provider, cmd.ProviderID)
	if err != nil {
		_, err := h.repo.CreateUserWithProvider(ctx, cmd.Email, cmd.Provider, cmd.ProviderID, cmd.AccessToken, cmd.RefreshToken)
		if err != nil {
			return "", err
		}
		u, _ = h.repo.GetByEmail(ctx, cmd.Email)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": u.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(h.encryptSecretKey)
}
