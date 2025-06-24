package command

import (
	"context"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
)

type CreateShortURLCommand struct {
	OriginalURL string
	Length      int
	MaxRetries  int
}

type CreateShortURLHandler struct {
	repo           url.URLRepository
	cache          url.URLCacheRepository
	secretKey      string
	expireDuration time.Duration
}

func NewCreateShortURLHandler(repo url.URLRepository, cache url.URLCacheRepository, secretKey string, expireDuration time.Duration) *CreateShortURLHandler {
	return &CreateShortURLHandler{
		repo:           repo,
		cache:          cache,
		secretKey:      secretKey,
		expireDuration: expireDuration,
	}
}

func (h *CreateShortURLHandler) Handle(ctx context.Context, cmd CreateShortURLCommand) (url.URL, error) {
	for range cmd.MaxRetries {
		shortCode, err := url.GenerateShortCode(cmd.Length)
		if err != nil {
			return url.URL{}, err
		}

		existsRedis, err := h.cache.Exists(ctx, shortCode)
		if err != nil {
			return url.URL{}, err
		}
		if existsRedis {
			continue
		}

		existsDB, err := h.repo.Exists(ctx, shortCode)
		if err != nil {
			return url.URL{}, err
		}
		if existsDB {
			continue
		}

		encryptedUrl := crypto.Encrypt(cmd.OriginalURL, h.secretKey)
		u := &url.URL{
			ShortCode:    shortCode,
			EncryptedURL: encryptedUrl,
		}

		err = h.cache.Save(ctx, u, h.expireDuration)
		if err != nil {
			return url.URL{}, err
		}

		shortURL := url.URL{ShortCode: shortCode}
		if err := h.repo.Save(ctx, u); err != nil {
			_ = h.cache.Delete(ctx, shortCode)
			return url.URL{}, err
		}

		return shortURL, nil
	}

	return url.URL{}, url.ErrMaxRetries
}
