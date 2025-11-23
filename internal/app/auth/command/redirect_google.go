package command

import (
	"context"

	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
)

type RedirectGoogleHandler struct {
	provider session_domain.OAuthProvider
}

func NewRedirectGoogleHandler(provider session_domain.OAuthProvider) *RedirectGoogleHandler {
	return &RedirectGoogleHandler{
		provider: provider,
	}
}

func (h *RedirectGoogleHandler) Handle(ctx context.Context) string {
	url := h.provider.GetAuthURL("state-token")
	return url
}
