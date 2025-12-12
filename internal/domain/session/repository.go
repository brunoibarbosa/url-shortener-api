package session

import (
	"context"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain"
	"github.com/google/uuid"
)

type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	FindByRefreshToken(ctx context.Context, hash string) (*Session, error)
	Revoke(ctx context.Context, id uuid.UUID) error
}

type BlacklistRepository interface {
	IsRevoked(ctx context.Context, token string) (bool, error)
	Revoke(ctx context.Context, token string, expiresIn time.Duration) error
}

type SessionQueryRepository interface {
	List(ctx context.Context, params ListSessionsParams) ([]ListSessionsDTO, uint64, error)
}

type ListSessionsDTO struct {
	UserAgent string
	IPAddress string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type ListSessionsSortBy uint8

const (
	ListSessionsSortByNone ListSessionsSortBy = iota
	ListSessionsSortByUserAgent
	ListSessionsSortByIPAddress
	ListSessionsSortByCreatedAt
	ListSessionsSortByExpiresAt
)

type ListSessionsParams struct {
	SortBy     ListSessionsSortBy
	SortKind   domain.SortKind
	Pagination domain.Pagination
}
