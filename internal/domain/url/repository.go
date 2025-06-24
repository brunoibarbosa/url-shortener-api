package url

import (
	"context"
	"time"
)

type URLRepository interface {
	Save(ctx context.Context, url *URL) error
	Exists(ctx context.Context, shortCode string) (bool, error)
	FindByShortCode(ctx context.Context, shortCode string) (*URL, error)
}

type URLCacheRepository interface {
	Exists(ctx context.Context, shortCode string) (bool, error)
	Save(ctx context.Context, url *URL, expires time.Duration) error
	Delete(ctx context.Context, shortCode string) error
	FindByShortCode(ctx context.Context, shortCode string) (*URL, error)
}
