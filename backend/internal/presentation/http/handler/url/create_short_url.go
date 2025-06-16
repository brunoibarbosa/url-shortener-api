package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http"
)

type CreateShortURLPayload struct {
	URL string `json:"url"`
}

type CreateShortURL201Response struct {
	ShortCode string `json:"short_code"`
}

type CreateShortURLHTTPHandler struct {
	useCase *command.CreateShortURLHandler
}

func NewCreateShortURLHTTPHandler(useCase *command.CreateShortURLHandler) *CreateShortURLHTTPHandler {
	return &CreateShortURLHTTPHandler{
		useCase: useCase,
	}
}

func (h *CreateShortURLHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var payload CreateShortURLPayload
	decodeErr := json.NewDecoder(r.Body).Decode(&payload)

	if errors.Is(decodeErr, io.EOF) {
		handler.WriteJSONError(w, http.StatusBadRequest, handler.ErrorCode.InvalidRequest, "Request body must not be empty")
		return
	}

	if payload.URL == "" {
		handler.WriteJSONError(w, http.StatusBadRequest, handler.ErrorCode.InvalidRequest, "'url' field is required in the request body")
		return
	}

	if !(strings.HasPrefix(payload.URL, "https://") || strings.HasPrefix(payload.URL, "http://")) {
		handler.WriteJSONError(w, http.StatusBadRequest, handler.ErrorCode.InvalidRequest, "The 'url' field must start with https:// or http://")
		return
	}

	appCmd := command.CreateShortURLCommand{OriginalURL: payload.URL}
	shortCode, err := h.useCase.Handle(appCmd)
	if err != nil {
		handler.WriteJSONError(w, http.StatusBadRequest, handler.ErrorCode.InvalidRequest, "Failed to create short URL")
		return
	}

	response := CreateShortURL201Response{
		ShortCode: shortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(response)
}
