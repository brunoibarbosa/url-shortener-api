package command

import (
	"github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
)

type CreateShortURLCommand struct {
	OriginalURL string
}

type CreateShortURLHandler struct {
	repo      url.URLRepository
	secretKey string
}

func NewCreateShortURLHandler(repo url.URLRepository, secretKey string) *CreateShortURLHandler {
	return &CreateShortURLHandler{
		repo:      repo,
		secretKey: secretKey,
	}
}

func (h *CreateShortURLHandler) Handle(cmd CreateShortURLCommand) (string, error) {
	shortCode := url.GenerateShortCode()

	encryptedUrl := crypto.Encrypt(cmd.OriginalURL, h.secretKey)
	url := &url.URL{
		ShortCode:    shortCode,
		EncryptedURL: encryptedUrl,
	}

	if err := h.repo.Save(url); err != nil {
		return "", err
	}

	return shortCode, nil
}
