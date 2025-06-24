package redis_repo

import (
	"context"
	"fmt"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain/url"
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
	exists, err := r.client.Exists(ctx, r.getKey(shortCode)).Result()
	return exists > 0, err
}

func (r *URLCacheRepository) Save(ctx context.Context, url *url.URL, expires time.Duration) error {
	return r.client.Set(ctx, url.ShortCode, url.EncryptedURL, expires).Err()
}

func (r *URLCacheRepository) FindByShortCode(ctx context.Context, shortCode string) (*url.URL, error) {
	encryptedUrl, err := r.client.Get(ctx, shortCode).Result()

	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	return &url.URL{
		ShortCode:    shortCode,
		EncryptedURL: encryptedUrl,
	}, nil
}

func (r *URLCacheRepository) Delete(ctx context.Context, shortCode string) error {
	return r.client.Del(ctx, r.getKey(shortCode)).Err()
}

func (r *URLCacheRepository) getKey(shortCode string) string {
	return fmt.Sprintf("url:short_code:%s", shortCode)
}
