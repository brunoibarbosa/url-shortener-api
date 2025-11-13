package session

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrNotFound            = errors.New("session not found")
	ErrRevokeFailed        = errors.New("failed to revoke token")
)

type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	UserAgent        string
	IPAddress        string
	ExpiresAt        *time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
}

func (s *Session) IsExpired() bool {
	return time.Now().After(*s.ExpiresAt) || s.RevokedAt != nil
}
