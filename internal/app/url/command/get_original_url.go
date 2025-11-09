package command

import (
	"context"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
	"github.com/brunoibarbosa/url-shortener/pkg/util"
)

type GetOriginalURLQuery struct {
	ShortCode string
}

type GetOriginalURLHandler struct {
	persistRepo             domain.URLRepository
	cacheRepo               domain.URLCacheRepository
	decryptSecretKey        string
	cacheExpirationDuration time.Duration
}

func NewGetOriginalURLHandler(repo domain.URLRepository, cache domain.URLCacheRepository, secretKey string, cacheExpirationDuration time.Duration) *GetOriginalURLHandler {
	return &GetOriginalURLHandler{
		persistRepo:             repo,
		cacheRepo:               cache,
		decryptSecretKey:        secretKey,
		cacheExpirationDuration: cacheExpirationDuration,
	}
}

func (h *GetOriginalURLHandler) Handle(ctx context.Context, query GetOriginalURLQuery) (string, error) {
	cachedUrl, err := h.cacheRepo.FindByShortCode(ctx, query.ShortCode)
	if err != nil {
		return "", err
	}
	if cachedUrl != nil {
		decryptedUrl := crypto.DecryptURL(cachedUrl.EncryptedURL, h.decryptSecretKey)
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
		timeRemainingToExpire := url.ExpiresAt.UTC().Sub(time.Now().UTC())
		cacheDuration = util.MinTimeDuration(timeRemainingToExpire, h.cacheExpirationDuration)
	}
	h.cacheRepo.Save(ctx, url, cacheDuration)

	decryptedUrl := crypto.DecryptURL(url.EncryptedURL, h.decryptSecretKey)
	return decryptedUrl, nil
}
