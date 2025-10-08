package handler

import (
	"context"
	"encoding/json"
	err "errors"
	"io"
	"net/http"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/user/command"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler"
	"github.com/brunoibarbosa/url-shortener/internal/validation"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterUser201Response struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type RegisterUserHTTPHandler struct {
	cmd *command.RegisterUserHandler
}

func NewRegisterUserHTTPHandler(cmd *command.RegisterUserHandler) *RegisterUserHTTPHandler {
	return &RegisterUserHTTPHandler{
		cmd: cmd,
	}
}

func (h *RegisterUserHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) (handler.HandlerResponse, *handler.HTTPError) {
	ctx := r.Context()

	payload, err := parseAndValidatePayload(r, ctx)
	if err != nil {
		return nil, err
	}

	appCmd := command.RegisterUserCommand{Email: payload.Email, Password: payload.Password}
	user, handleErr := h.cmd.Handle(r.Context(), appCmd)
	if handleErr != nil {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.user.create_failed", nil)
	}

	response := RegisterUser201Response{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.common.encode_failed", nil)
	}

	return nil, nil
}

func parseAndValidatePayload(r *http.Request, ctx context.Context) (RegisterUserPayload, *handler.HTTPError) {
	var payload RegisterUserPayload
	decodeErr := json.NewDecoder(r.Body).Decode(&payload)

	if err.Is(decodeErr, io.EOF) {
		return RegisterUserPayload{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.common.empty_body", nil)
	}

	if validationErr := validation.ValidateEmail(payload.Email); validationErr != nil {
		var errorCode string

		switch {
		default:
			errorCode = "error.email.invalid_email_format"
		}

		return RegisterUserPayload{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeValidationError, errorCode, nil)
	}

	return payload, nil
}
