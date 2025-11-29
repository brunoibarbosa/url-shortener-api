package command

import (
	"context"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/pkg/util"
)

type GetOriginalURLQuery struct {
	ShortCode string
}

type GetOriginalURLHandler struct {
	persistRepo             domain.URLRepository
	cacheRepo               domain.URLCacheRepository
	encrypter               domain.URLEncrypter
	cacheExpirationDuration time.Duration
}

func NewGetOriginalURLHandler(
	repo domain.URLRepository,
	cache domain.URLCacheRepository,
	encrypter domain.URLEncrypter,
	cacheExpirationDuration time.Duration,
) *GetOriginalURLHandler {
	return &GetOriginalURLHandler{
		persistRepo:             repo,
		cacheRepo:               cache,
		encrypter:               encrypter,
		cacheExpirationDuration: cacheExpirationDuration,
	}
}

func (h *GetOriginalURLHandler) Handle(ctx context.Context, query GetOriginalURLQuery) (string, error) {
	cachedUrl, err := h.cacheRepo.FindByShortCode(ctx, query.ShortCode)
	if err != nil {
		return "", err
	}
	if cachedUrl != nil {
		decryptedUrl, err := h.encrypter.Decrypt(cachedUrl.EncryptedURL)
		if err != nil {
			return "", err
		}
		return decryptedUrl, nil
	}

	url, err := h.persistRepo.FindByShortCode(ctx, query.ShortCode)

	if err != nil {
		return "", err
	}

	if url == nil {
		return "", domain.ErrURLNotFound
	}

	cacheDuration := h.cacheExpirationDuration
	if url.ExpiresAt != nil {
		timeRemainingToExpire := url.RemainingTTL(time.Now().UTC())
		cacheDuration = util.MinTimeDuration(timeRemainingToExpire, h.cacheExpirationDuration)
	}
	_ = h.cacheRepo.Save(ctx, url, cacheDuration)

	decryptedUrl, err := h.encrypter.Decrypt(url.EncryptedURL)
	if err != nil {
		return "", err
	}
	return decryptedUrl, nil
}
