package redis_repo

import (
	"context"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/redis/go-redis/v9"
)

type URLRepository struct {
	client *redis.Client
}

func NewURLRepository(client *redis.Client) *URLRepository {
	return &URLRepository{
		client: client,
	}
}

func (r *URLRepository) Save(url *url.URL) error {
	return r.client.Set(context.Background(), url.ShortCode, url.EncryptedURL, time.Hour).Err()
}

func (r *URLRepository) FindByShortCode(shortCode string) (*url.URL, error) {
	encryptedUrl, err := r.client.Get(context.Background(), shortCode).Result()

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
