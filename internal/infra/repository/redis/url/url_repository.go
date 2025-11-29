package redis_repo

import (
	"context"
	"fmt"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/redis/go-redis/v9"
)

type URLCacheRepository struct {
	client *redis.Client
}

func NewURLCacheRepository(client *redis.Client) *URLCacheRepository {
	return &URLCacheRepository{
		client: client,
	}
}

func (r *URLCacheRepository) Exists(ctx context.Context, shortCode string) (bool, error) {
	key := r.getKey(shortCode)
	exists, err := r.client.Exists(ctx, key).Result()
	return exists > 0, err
}

func (r *URLCacheRepository) Save(ctx context.Context, url *domain.URL, expires time.Duration) error {
	key := r.getKey(url.ShortCode)
	return r.client.Set(ctx, key, url.EncryptedURL, expires).Err()
}

func (r *URLCacheRepository) FindByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	key := r.getKey(shortCode)
	encryptedUrl, err := r.client.Get(ctx, key).Result()

	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	return &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: encryptedUrl,
		ExpiresAt:    nil,
	}, nil
}

func (r *URLCacheRepository) Delete(ctx context.Context, shortCode string) error {
	key := r.getKey(shortCode)
	return r.client.Del(ctx, key).Err()
}

func (r *URLCacheRepository) getKey(shortCode string) string {
	key := fmt.Sprintf("url:short_code:%s", shortCode)
	return key
}
