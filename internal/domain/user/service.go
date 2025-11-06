package user

import (
	"context"

	"github.com/google/uuid"
)

type OAuthUser struct {
	ID           string
	Name         string
	Email        string
	AccessToken  string
	RefreshToken string
	AvatarURL    *string
}

type TokenParams struct {
	UserID uuid.UUID
}

type OAuthProvider interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*OAuthUser, error)
}

type TokenService interface {
	GenerateAccessToken(params *TokenParams) (string, error)
}
