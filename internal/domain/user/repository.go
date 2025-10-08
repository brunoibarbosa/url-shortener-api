package user

import (
	"context"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*User, error)
	CreateUserWithProvider(ctx context.Context, email, provider, providerID, accessToken, refreshToken string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByProvider(ctx context.Context, provider, providerID string) (*User, error)
}
