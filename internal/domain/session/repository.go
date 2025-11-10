package session

import (
	"context"

	"github.com/google/uuid"
)

type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	FindByRefreshToken(ctx context.Context, hash string) (*Session, error)
	Revoke(ctx context.Context, id uuid.UUID) error
}
