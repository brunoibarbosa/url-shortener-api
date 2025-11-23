package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	Exists(ctx context.Context, email string) (bool, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, u *User) error
	WithTx(tx pgx.Tx) UserRepository
}

type UserProfileRepository interface {
	Create(ctx context.Context, userID uuid.UUID, p *UserProfile) error
	WithTx(tx pgx.Tx) UserProfileRepository
}

type UserProviderRepository interface {
	Find(ctx context.Context, provider, providerID string) (*UserProvider, error)
	Create(ctx context.Context, userID uuid.UUID, pv *UserProvider) error
	WithTx(tx pgx.Tx) UserProviderRepository
}
