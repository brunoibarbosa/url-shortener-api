package handler

import (
	"errors"
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler"
	app_errors "github.com/brunoibarbosa/url-shortener/pkg/errors"
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

	if err != nil && errors.Is(err, domain.ErrExpiredURL) {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusGone, app_errors.CodeNotFound, "error.url.expired_url", handler.Detail(ctx, "shortCode", "error.details.shortcode.expired"))
	}

	if (err != nil && errors.Is(err, domain.ErrURLNotFound)) || originalURL == "" {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusNotFound, app_errors.CodeNotFound, "error.common.not_found", handler.Detail(ctx, "shortCode", "error.details.shortcode.not_found"))
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
	return nil, nil
}
