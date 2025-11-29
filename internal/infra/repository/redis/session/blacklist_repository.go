package cache

import (
	"context"
	"fmt"
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
	key := r.getKey(token)
	exists, err := r.client.Exists(ctx, key).Result()
	return exists > 0, err
}

func (r *BlacklistRepository) Revoke(ctx context.Context, token string, expiresIn time.Duration) error {
	key := r.getKey(token)
	return r.client.Set(ctx, key, true, expiresIn).Err()
}

func (r *BlacklistRepository) getKey(token string) string {
	key := fmt.Sprintf("session:revoked:%s", token)
	return key
}
