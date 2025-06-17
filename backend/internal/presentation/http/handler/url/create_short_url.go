package handler

import (
	"encoding/json"
	err "errors"
	"io"
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	handler "github.com/brunoibarbosa/url-shortener/internal/presentation/http"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
	"github.com/brunoibarbosa/url-shortener/pkg/validation"
)

type CreateShortURLPayload struct {
	URL string `json:"url"`
}

type CreateShortURL201Response struct {
	ShortCode string `json:"short_code"`
}

type CreateShortURLHTTPHandler struct {
	cmd *command.CreateShortURLHandler
}

func NewCreateShortURLHTTPHandler(cmd *command.CreateShortURLHandler) *CreateShortURLHTTPHandler {
	return &CreateShortURLHTTPHandler{
		cmd: cmd,
	}
}

func (h *CreateShortURLHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) {
	payload := parseAndValidatePayload(r)

	appCmd := command.CreateShortURLCommand{OriginalURL: payload.URL}
	shortCode, err := h.cmd.Handle(appCmd)
	if err != nil {
		panic(handler.NewHTTPError(http.StatusInternalServerError, errors.CodeInternalError, "Failed to create short URL"))
	}

	response := CreateShortURL201Response{
		ShortCode: shortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(handler.NewHTTPError(http.StatusInternalServerError, errors.CodeInternalError, "failed to encode response"))
	}
}

func parseAndValidatePayload(r *http.Request) CreateShortURLPayload {
	var payload CreateShortURLPayload
	decodeErr := json.NewDecoder(r.Body).Decode(&payload)

	if err.Is(decodeErr, io.EOF) {
		panic(handler.NewHTTPError(http.StatusBadRequest, errors.CodeBadRequest, "request body must not be empty"))
	}

	if payload.URL == "" {
		panic(handler.NewHTTPError(http.StatusBadRequest, errors.CodeValidationError, "'url' field is required in the request body"))
	}

	if err := validation.ValidateURL(payload.URL); err != nil {
		panic(handler.NewHTTPError(http.StatusBadRequest, errors.CodeValidationError, err.Error()))
	}

	return payload
}
