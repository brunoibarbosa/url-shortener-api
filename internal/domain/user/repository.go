package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Exists(ctx context.Context, email string) (bool, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, u *User) error
}

type UserProfileRepository interface {
	Create(ctx context.Context, userID uuid.UUID, p *UserProfile) error
}

type UserProviderRepository interface {
	Find(ctx context.Context, provider, providerID string) (*UserProvider, error)
	Create(ctx context.Context, userID uuid.UUID, pv *UserProvider) error
}
