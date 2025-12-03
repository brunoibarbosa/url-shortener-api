package http_handler

import (
	"errors"
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler"
	app_errors "github.com/brunoibarbosa/url-shortener/pkg/errors"
	"github.com/go-chi/chi/v5"
)

type RedirectHTTPHandler struct {
	cmd *command.GetOriginalURLHandler
}

func NewRedirectHTTPHandler(cmd *command.GetOriginalURLHandler) *RedirectHTTPHandler {
	return &RedirectHTTPHandler{cmd: cmd}
}

func (h *RedirectHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) *http_handler.HTTPError {
	shortCode := chi.URLParam(r, "shortCode")
	ctx := r.Context()

	if shortCode == "" {
		return http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, app_errors.CodeBadRequest, "error.url.expired_url", http_handler.Detail(ctx, "shortCode", "error.details.shortcode.expired"))
	}

	appQuery := command.GetOriginalURLQuery{ShortCode: shortCode}
	originalURL, err := h.cmd.Handle(r.Context(), appQuery)

	if err != nil {
		if errors.Is(err, domain.ErrExpiredURL) {
			return http_handler.NewI18nHTTPError(ctx, http.StatusGone, app_errors.CodeNotFound, "error.url.expired_url", http_handler.Detail(ctx, "shortCode", "error.details.shortcode.expired"))
		}

		if errors.Is(err, domain.ErrInvalidShortCode) {
			return http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, app_errors.CodeBadRequest, "error.url.required_short_code", nil)
		}

		if (errors.Is(err, domain.ErrURLNotFound)) || originalURL == "" {
			return http_handler.NewI18nHTTPError(ctx, http.StatusNotFound, app_errors.CodeNotFound, "error.common.not_found", http_handler.Detail(ctx, "shortCode", "error.details.shortcode.not_found"))
		}
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
	return nil
}
