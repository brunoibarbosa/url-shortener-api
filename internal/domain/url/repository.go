package url

import (
	"context"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain"
	"github.com/google/uuid"
)

type URLRepository interface {
	Save(ctx context.Context, url *URL) error
	Exists(ctx context.Context, shortCode string) (bool, error)
	FindByShortCode(ctx context.Context, shortCode string) (*URL, error)
	SoftDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type URLCacheRepository interface {
	Exists(ctx context.Context, shortCode string) (bool, error)
	Save(ctx context.Context, url *URL, expires time.Duration) error
	Delete(ctx context.Context, shortCode string) error
	FindByShortCode(ctx context.Context, shortCode string) (*URL, error)
}

type URLQueryRepository interface {
	ListByUserID(ctx context.Context, userID uuid.UUID, params ListURLsParams) ([]ListURLsDTO, uint64, error)
}

type ListURLsDTO struct {
	ID        uuid.UUID
	ShortCode string
	ExpiresAt *time.Time
	CreatedAt time.Time
	DeletedAt *time.Time
}

type ListURLsSortBy uint8

const (
	ListURLsSortByNone ListURLsSortBy = iota
	ListURLsSortByCreatedAt
	ListURLsSortByExpiresAt
)

type ListURLsParams struct {
	SortBy     ListURLsSortBy
	SortKind   domain.SortKind
	Pagination domain.Pagination
}
