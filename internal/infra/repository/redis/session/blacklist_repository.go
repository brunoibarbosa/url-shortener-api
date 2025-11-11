package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type BlacklistRepository struct {
	client *redis.Client
}

func NewBlacklistRepository(client *redis.Client) *BlacklistRepository {
	return &BlacklistRepository{
		client: client,
	}
}

func (r *BlacklistRepository) IsRevoked(ctx context.Context, token string) (bool, error) {
	exists, err := r.client.Exists(ctx, "revoked:"+token).Result()
	return exists > 0, err
}

func (r *BlacklistRepository) Revoke(ctx context.Context, token string, expiresIn time.Duration) error {
	return r.client.Set(ctx, "revoked:"+token, true, expiresIn).Err()
}
