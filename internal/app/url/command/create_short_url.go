package command

import (
	"context"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/google/uuid"

	"github.com/brunoibarbosa/url-shortener/pkg/util"
)

type CreateShortURLCommand struct {
	OriginalURL string
	UserID      *uuid.UUID
	Length      int
	MaxRetries  int
}

type CreateShortURLHandler struct {
	persistRepo               domain.URLRepository
	cacheRepo                 domain.URLCacheRepository
	encrypter                 domain.URLEncrypter
	shortCodeGenerator        domain.ShortCodeGenerator
	persistExpirationDuration time.Duration
	cacheExpirationDuration   time.Duration
}

type CreateShortURLResult struct {
	ShortCode string
}

func NewCreateShortURLHandler(
	repo domain.URLRepository,
	cache domain.URLCacheRepository,
	encrypter domain.URLEncrypter,
	shortCodeGenerator domain.ShortCodeGenerator,
	persistExpirationDuration time.Duration,
	cacheExpirationDuration time.Duration,
) *CreateShortURLHandler {
	return &CreateShortURLHandler{
		persistRepo:               repo,
		cacheRepo:                 cache,
		encrypter:                 encrypter,
		shortCodeGenerator:        shortCodeGenerator,
		persistExpirationDuration: persistExpirationDuration,
		cacheExpirationDuration:   cacheExpirationDuration,
	}
}

func (h *CreateShortURLHandler) Handle(ctx context.Context, cmd CreateShortURLCommand) (CreateShortURLResult, error) {
	for i := 0; i < cmd.MaxRetries; i++ {
		shortCode, err := h.shortCodeGenerator.Generate(cmd.Length)
		if err != nil {
			return CreateShortURLResult{}, err
		}

		existsRedis, err := h.cacheRepo.Exists(ctx, shortCode)
		if err != nil {
			return CreateShortURLResult{}, err
		}
		if existsRedis {
			continue
		}

		existsDB, err := h.persistRepo.Exists(ctx, shortCode)
		if err != nil {
			return CreateShortURLResult{}, err
		}
		if existsDB {
			continue
		}

		encryptedUrl, err := h.encrypter.Encrypt(cmd.OriginalURL)
		if err != nil {
			return CreateShortURLResult{}, err
		}

		expiresAt := time.Now().Add(h.persistExpirationDuration)
		u := &domain.URL{
			ShortCode:    shortCode,
			EncryptedURL: encryptedUrl,
			UserID:       cmd.UserID,
			ExpiresAt:    &expiresAt,
		}

		cacheDuration := util.MinTimeDuration(h.cacheExpirationDuration, h.persistExpirationDuration)
		err = h.cacheRepo.Save(ctx, u, cacheDuration)
		if err != nil {
			return CreateShortURLResult{}, err
		}

		shortURL := CreateShortURLResult{ShortCode: shortCode}
		if err := h.persistRepo.Save(ctx, u); err != nil {
			_ = h.cacheRepo.Delete(ctx, shortCode)
			return CreateShortURLResult{}, err
		}

		return shortURL, nil
	}

	return CreateShortURLResult{}, domain.ErrMaxRetries
}
