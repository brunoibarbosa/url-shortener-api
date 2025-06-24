package command

import (
	"context"
	"errors"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
)

type GetOriginalURLQuery struct {
	ShortCode string
}

type GetOriginalURLHandler struct {
	repo           url.URLRepository
	cache          url.URLCacheRepository
	secretKey      string
	expireDuration time.Duration
}

func NewGetOriginalURLHandler(repo url.URLRepository, cache url.URLCacheRepository, secretKey string, expireDuration time.Duration) *GetOriginalURLHandler {
	return &GetOriginalURLHandler{
		repo:           repo,
		cache:          cache,
		secretKey:      secretKey,
		expireDuration: expireDuration,
	}
}

func (h *GetOriginalURLHandler) Handle(ctx context.Context, query GetOriginalURLQuery) (string, error) {
	cachedUrl, err := h.cache.FindByShortCode(ctx, query.ShortCode)
	if err != nil {
		return "", err
	}
	if cachedUrl != nil {
		decryptedUrl := crypto.Decrypt(cachedUrl.EncryptedURL, h.secretKey)
		return decryptedUrl, nil
	}

	url, err := h.repo.FindByShortCode(ctx, query.ShortCode)

	if err != nil {
		return "", err
	}

	if url == nil {
		return "", errors.New("URL not found")
	}

	h.cache.Save(ctx, url, h.expireDuration)

	decryptedUrl := crypto.Decrypt(url.EncryptedURL, h.secretKey)
	return decryptedUrl, nil
}
