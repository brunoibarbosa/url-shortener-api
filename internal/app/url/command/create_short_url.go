package command

import (
	"context"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
	"github.com/brunoibarbosa/url-shortener/pkg/util"
)

type CreateShortURLCommand struct {
	OriginalURL string
	Length      int
	MaxRetries  int
}

type CreateShortURLHandler struct {
	persistRepo               url.URLRepository
	cacheRepo                 url.URLCacheRepository
	encryptSecretKey          string
	persistExpirationDuration time.Duration
	cacheExpirationDuration   time.Duration
}

func NewCreateShortURLHandler(repo url.URLRepository, cache url.URLCacheRepository, secretKey string, persistExpirationDuration time.Duration, cacheExpirationDuration time.Duration) *CreateShortURLHandler {
	return &CreateShortURLHandler{
		persistRepo:               repo,
		cacheRepo:                 cache,
		encryptSecretKey:          secretKey,
		persistExpirationDuration: persistExpirationDuration,
		cacheExpirationDuration:   cacheExpirationDuration,
	}
}

func (h *CreateShortURLHandler) Handle(ctx context.Context, cmd CreateShortURLCommand) (url.URL, error) {
	for range cmd.MaxRetries {
		shortCode, err := url.GenerateShortCode(cmd.Length)
		if err != nil {
			return url.URL{}, err
		}

		existsRedis, err := h.cacheRepo.Exists(ctx, shortCode)
		if err != nil {
			return url.URL{}, err
		}
		if existsRedis {
			continue
		}

		existsDB, err := h.persistRepo.Exists(ctx, shortCode)
		if err != nil {
			return url.URL{}, err
		}
		if existsDB {
			continue
		}

		encryptedUrl := crypto.Encrypt(cmd.OriginalURL, h.encryptSecretKey)
		expiresAt := time.Now().Add(h.persistExpirationDuration)
		u := &url.URL{
			ShortCode:    shortCode,
			EncryptedURL: encryptedUrl,
			ExpiresAt:    &expiresAt,
		}

		cacheDuration := util.MinTimeDuration(h.cacheExpirationDuration, h.persistExpirationDuration)
		err = h.cacheRepo.Save(ctx, u, cacheDuration)
		if err != nil {
			return url.URL{}, err
		}

		shortURL := url.URL{ShortCode: shortCode}
		if err := h.persistRepo.Save(ctx, u); err != nil {
			_ = h.cacheRepo.Delete(ctx, shortCode)
			return url.URL{}, err
		}

		return shortURL, nil
	}

	return url.URL{}, url.ErrMaxRetries
}
