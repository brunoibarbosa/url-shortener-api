package handler

import (
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler"
)

type RedirectGoogleHTTPHandler struct {
	cmd *command.RedirectGoogleHandler
}

func NewRedirectGoogleHTTPHandler(cmd *command.RedirectGoogleHandler) *RedirectGoogleHTTPHandler {
	return &RedirectGoogleHTTPHandler{
		cmd: cmd,
	}
}

func (h *RedirectGoogleHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) (handler.HandlerResponse, *handler.HTTPError) {
	ctx := r.Context()
	url := h.cmd.Handle(ctx)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	return nil, nil
}
