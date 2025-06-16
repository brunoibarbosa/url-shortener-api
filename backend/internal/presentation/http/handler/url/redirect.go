package handler

import (
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
	"github.com/go-chi/chi/v5"
)

type RedirectHTTPHandler struct {
	cmd *command.GetOriginalURLHandler
}

func NewRedirectHTTPHandler(cmd *command.GetOriginalURLHandler) *RedirectHTTPHandler {
	return &RedirectHTTPHandler{cmd: cmd}
}

func (h *RedirectHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")

	appQuery := command.GetOriginalURLQuery{ShortCode: shortCode}
	originalURL, err := h.cmd.Handle(appQuery)
	if err != nil || originalURL == "" {
		panic(handler.NewHTTPError(http.StatusNotFound, errors.CodeNotFound, "The requested URL was not found"))
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
