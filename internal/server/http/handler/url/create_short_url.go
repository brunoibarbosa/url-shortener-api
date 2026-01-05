package http_handler

import (
	"context"
	"encoding/json"
	err "errors"
	"io"
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler"
	http_middleware "github.com/brunoibarbosa/url-shortener/internal/server/http/middleware"
	"github.com/brunoibarbosa/url-shortener/internal/validation"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
	"github.com/google/uuid"
)

type CreateShortURLPayload struct {
	URL string `json:"url"`
}

type CreateShortURL201Response struct {
	ShortCode string `json:"shortCode"`
}

type CreateShortURLHTTPHandler struct {
	cmd *command.CreateShortURLHandler
}

func NewCreateShortURLHTTPHandler(cmd *command.CreateShortURLHandler) *CreateShortURLHTTPHandler {
	return &CreateShortURLHTTPHandler{
		cmd: cmd,
	}
}

func (h *CreateShortURLHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) *http_handler.HTTPError {
	ctx := r.Context()

	payload, err := validateShortenPayload(r, ctx)
	if err != nil {
		return err
	}

	userID := extractUserIDFromContext(r)

	appCmd := command.CreateShortURLCommand{
		OriginalURL: payload.URL,
		UserID:      userID,
		Length:      6,
		MaxRetries:  10,
	}
	url, handleErr := h.cmd.Handle(r.Context(), appCmd)
	if handleErr != nil {
		return http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.url.create_failed", nil)
	}

	response := CreateShortURL201Response{
		ShortCode: url.ShortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		return http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.common.encode_failed", nil)
	}

	return nil
}

func validateShortenPayload(r *http.Request, ctx context.Context) (CreateShortURLPayload, *http_handler.HTTPError) {
	var payload CreateShortURLPayload
	decodeErr := json.NewDecoder(r.Body).Decode(&payload)

	if err.Is(decodeErr, io.EOF) {
		return CreateShortURLPayload{}, http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.common.empty_body", nil)
	}

	ec := http_handler.NewErrorCollector(ctx)

	if payload.URL == "" {
		ec.AddFieldError("url", "error.details.field_required")
	}

	if validationErr := validation.ValidateURL(payload.URL); validationErr != nil {
		var detailKey string

		switch {
		case err.Is(validationErr, domain.ErrMissingURLSchema):
			detailKey = "error.details.url.missing_scheme"
		case err.Is(validationErr, domain.ErrUnsupportedURLSchema):
			detailKey = "error.details.url.invalid_format"
		case err.Is(validationErr, domain.ErrMissingURLHost):
			detailKey = "error.details.url.invalid_format"
		default:
			detailKey = "error.details.url.invalid_format"
		}

		ec.AddFieldError("url", detailKey)
	}

	if ec.HasErrors() {
		return CreateShortURLPayload{}, ec.ToHTTPError(http.StatusBadRequest, errors.CodeValidationError, "error.validation.failed")
	}

	return payload, nil
}

func extractUserIDFromContext(r *http.Request) *uuid.UUID {
	if userID, ok := r.Context().Value(http_middleware.UserIDKey).(uuid.UUID); ok {
		return &userID
	}
	return nil
}
