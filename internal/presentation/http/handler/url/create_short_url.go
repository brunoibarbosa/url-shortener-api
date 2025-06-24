package handler

import (
	"context"
	"encoding/json"
	err "errors"
	"io"
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	"github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler"
	"github.com/brunoibarbosa/url-shortener/internal/validation"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
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

func (h *CreateShortURLHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) (handler.HandlerResponse, *handler.HTTPError) {
	ctx := r.Context()

	payload, err := parseAndValidatePayload(r, ctx)
	if err != nil {
		return nil, err
	}

	appCmd := command.CreateShortURLCommand{OriginalURL: payload.URL, Length: 6, MaxRetries: 10}
	url, handleErr := h.cmd.Handle(r.Context(), appCmd)
	if handleErr != nil {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.url.create_failed", nil)
	}

	response := CreateShortURL201Response{
		ShortCode: url.ShortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.common.encode_failed", nil)
	}

	return nil, nil
}

func parseAndValidatePayload(r *http.Request, ctx context.Context) (CreateShortURLPayload, *handler.HTTPError) {
	var payload CreateShortURLPayload
	decodeErr := json.NewDecoder(r.Body).Decode(&payload)

	if err.Is(decodeErr, io.EOF) {
		return CreateShortURLPayload{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.common.empty_body", nil)
	}

	if payload.URL == "" {
		return CreateShortURLPayload{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeValidationError, "error.url.url_required", nil)
	}

	if validationErr := validation.ValidateURL(payload.URL); validationErr != nil {
		var errorCode string

		switch {
		case err.Is(validationErr, url.ErrMissingURLSchema):
			errorCode = "error.url.missing_schema"
		case err.Is(validationErr, url.ErrUnsupportedURLSchema):
			errorCode = "error.url.unsupported_schema"
		case err.Is(validationErr, url.ErrMissingURLHost):
			errorCode = "error.url.missing_host"
		default:
			errorCode = "error.url.invalid_url"
		}

		return CreateShortURLPayload{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeValidationError, errorCode, nil)
	}

	return payload, nil
}
