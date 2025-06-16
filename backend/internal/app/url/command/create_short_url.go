package command

import (
	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
)

type CreateShortURLCommand struct {
	OriginalURL string
}

type CreateShortURLHandler struct {
	repo      domain.URLRepository
	secretKey string
}

func NewCreateShortURLHandler(repo domain.URLRepository, secretKey string) *CreateShortURLHandler {
	return &CreateShortURLHandler{
		repo:      repo,
		secretKey: secretKey,
	}
}

func (h *CreateShortURLHandler) Handle(cmd CreateShortURLCommand) (string, error) {
	shortCode := domain.GenerateShortCode()

	encryptedUrl := crypto.Encrypt(cmd.OriginalURL, h.secretKey)
	url := &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: encryptedUrl,
	}

	if err := h.repo.Save(url); err != nil {
		return "", err
	}

	return shortCode, nil
}
