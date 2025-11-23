package session

import (
	"context"
	"errors"
	"time"

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
	Duration  time.Duration
}

type TokenClaims struct {
	Sub string
	Sid string
	Exp int64
	Iat int64
}

type OAuthProvider interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*OAuthUser, error)
}

type TokenService interface {
	GenerateAccessToken(params *TokenParams) (string, error)
	GenerateRefreshToken() uuid.UUID
}
