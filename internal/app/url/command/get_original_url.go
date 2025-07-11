package command

import (
	"context"
	"errors"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
	"github.com/brunoibarbosa/url-shortener/pkg/util"
)

type GetOriginalURLQuery struct {
	ShortCode string
}

type GetOriginalURLHandler struct {
	persistRepo             url.URLRepository
	cacheRepo               url.URLCacheRepository
	decryptSecretKey        string
	cacheExpirationDuration time.Duration
}

func NewGetOriginalURLHandler(repo url.URLRepository, cache url.URLCacheRepository, secretKey string, cacheExpirationDuration time.Duration) *GetOriginalURLHandler {
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
		decryptedUrl := crypto.Decrypt(cachedUrl.EncryptedURL, h.decryptSecretKey)
		return decryptedUrl, nil
	}

	url, err := h.persistRepo.FindByShortCode(ctx, query.ShortCode)

	if err != nil {
		return "", err
	}

	if url == nil {
		return "", errors.New("URL not found")
	}

	cacheDuration := h.cacheExpirationDuration
	if url.ExpiresAt != nil {
		timeRemainingToExpire := url.ExpiresAt.UTC().Sub(time.Now().UTC())
		cacheDuration = util.MinTimeDuration(timeRemainingToExpire, h.cacheExpirationDuration)
	}
	h.cacheRepo.Save(ctx, url, cacheDuration)

	decryptedUrl := crypto.Decrypt(url.EncryptedURL, h.decryptSecretKey)
	return decryptedUrl, nil
}
