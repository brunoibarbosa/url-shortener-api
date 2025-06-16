package handler

import (
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http"
	"github.com/go-chi/chi/v5"
)

type RedirectHTTPHandler struct {
	useCase *command.GetOriginalURLHandler
}

func NewRedirectHTTPHandler(useCase *command.GetOriginalURLHandler) *RedirectHTTPHandler {
	return &RedirectHTTPHandler{useCase: useCase}
}

func (h *RedirectHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")

	appQuery := command.GetOriginalURLQuery{ShortCode: shortCode}
	originalURL, err := h.useCase.Handle(appQuery)
	if err != nil || originalURL == "" {
		handler.WriteJSONError(w, http.StatusBadRequest, handler.ErrorCode.NotFound, "The requested URL was not found")
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
