package command

import (
	"context"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
)

type RedirectGoogleHandler struct {
	provider domain.OAuthProvider
}

func NewRedirectGoogleHandler(provider domain.OAuthProvider) *RedirectGoogleHandler {
	return &RedirectGoogleHandler{
		provider: provider,
	}
}

func (h *RedirectGoogleHandler) Handle(ctx context.Context) string {
	url := h.provider.GetAuthURL("state-token")
	return url
}
