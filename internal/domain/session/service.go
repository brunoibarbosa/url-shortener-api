package session

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrTokenGenerate = errors.New("failed to generate token")
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
	UserID    uuid.UUID
	SessionID uuid.UUID
}

type OAuthProvider interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*OAuthUser, error)
}

type TokenService interface {
	GenerateAccessToken(params *TokenParams) (string, error)
	GenerateRefreshToken() uuid.UUID
}
