package handler

import (
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
)

type RedirectGoogleHTTPHandler struct {
	cmd *command.RedirectGoogleHandler
}

func NewRedirectGoogleHTTPHandler(cmd *command.RedirectGoogleHandler) *RedirectGoogleHTTPHandler {
	return &RedirectGoogleHTTPHandler{
		cmd,
	}
}

func (h *RedirectGoogleHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) (http_handler.HandlerResponse, *http_handler.HTTPError) {
	ctx := r.Context()
	url, err := h.cmd.Handle(ctx)
	if err != nil {
		return nil, http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.redirect.failed", nil)
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	return nil, nil
}
