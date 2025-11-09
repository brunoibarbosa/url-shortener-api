package session

import "context"

type SessionRepository interface {
	Create(ctx context.Context, u *Session) error
}
