package handler

import (
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
	"github.com/go-chi/chi/v5"
)

type RedirectHTTPHandler struct {
	cmd *command.GetOriginalURLHandler
}

func NewRedirectHTTPHandler(cmd *command.GetOriginalURLHandler) *RedirectHTTPHandler {
	return &RedirectHTTPHandler{cmd: cmd}
}

func (h *RedirectHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) (handler.HandlerResponse, *handler.HTTPError) {
	shortCode := chi.URLParam(r, "shortCode")

	appQuery := command.GetOriginalURLQuery{ShortCode: shortCode}
	originalURL, err := h.cmd.Handle(r.Context(), appQuery)
	ctx := r.Context()
	if err != nil || originalURL == "" {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusNotFound, errors.CodeNotFound, "error.common.not_found", nil)
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
	return nil, nil
}
