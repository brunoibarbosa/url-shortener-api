package command

import (
	"context"

	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
)

type RedirectGoogleHandler struct {
	provider     session_domain.OAuthProvider
	stateService session_domain.StateService
}

func NewRedirectGoogleHandler(
	provider session_domain.OAuthProvider,
	stateService session_domain.StateService,
) *RedirectGoogleHandler {
	return &RedirectGoogleHandler{
		provider:     provider,
		stateService: stateService,
	}
}

func (h *RedirectGoogleHandler) Handle(ctx context.Context) (string, error) {
	state, err := h.stateService.GenerateState(ctx)
	if err != nil {
		return "", err
	}

	url := h.provider.GetAuthURL(state)
	return url, nil
}
