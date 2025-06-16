package command

import (
	"errors"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
)

type GetOriginalURLQuery struct {
	ShortCode string
}

type GetOriginalURLHandler struct {
	repo      domain.URLRepository
	secretKey string
}

func NewGetOriginalURLHandler(repo domain.URLRepository, secretKey string) *GetOriginalURLHandler {
	return &GetOriginalURLHandler{
		repo:      repo,
		secretKey: secretKey,
	}
}

func (h *GetOriginalURLHandler) Handle(query GetOriginalURLQuery) (string, error) {
	url, err := h.repo.FindByShortCode(query.ShortCode)

	if err != nil {
		return "", err
	}

	if url == nil {
		return "", errors.New("URL not found")
	}

	decryptedUrl := crypto.Decrypt(url.EncryptedURL, h.secretKey)
	return decryptedUrl, nil
}
