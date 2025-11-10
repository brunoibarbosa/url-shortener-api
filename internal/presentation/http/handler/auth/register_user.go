package handler

import (
	"context"
	"encoding/json"
	err "errors"
	"io"
	"net/http"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler"
	"github.com/brunoibarbosa/url-shortener/internal/validation"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
	"github.com/google/uuid"
)

type RegisterUserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type RegisterUserProfile201Response struct {
	Name string `json:"name"`
}

type RegisterUser201Response struct {
	ID        uuid.UUID                      `json:"id"`
	Email     string                         `json:"email"`
	CreatedAt time.Time                      `json:"createdAt"`
	Profile   RegisterUserProfile201Response `json:"profile"`
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

	payload, validationErr := validateRegisterPayload(r, ctx)
	if validationErr != nil {
		return nil, validationErr
	}

	appCmd := command.RegisterUserCommand{Email: payload.Email, Password: payload.Password, Name: payload.Name}
	user, handleErr := h.cmd.Handle(r.Context(), appCmd)
	if handleErr != nil {
		switch {
		case err.Is(handleErr, domain.ErrEmailAlreadyExists):
			return nil, handler.NewI18nHTTPError(ctx, http.StatusConflict, errors.CodeValidationError, "error.email.email_already_exists", nil)
		default:
			return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.user.create_failed", nil)
		}
	}

	response := RegisterUser201Response{
		ID:    user.ID,
		Email: user.Email,
		Profile: RegisterUserProfile201Response{
			Name: user.Profile.Name,
		},
		CreatedAt: user.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.common.encode_failed", nil)
	}

	return nil, nil
}

func validateRegisterPayload(r *http.Request, ctx context.Context) (RegisterUserPayload, *handler.HTTPError) {
	var payload RegisterUserPayload
	decodeErr := json.NewDecoder(r.Body).Decode(&payload)

	if err.Is(decodeErr, io.EOF) {
		return RegisterUserPayload{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.common.empty_body", nil)
	}

	if validationErr := validation.ValidateEmail(payload.Email); validationErr != nil {
		return RegisterUserPayload{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeValidationError, "error.email.invalid_email_format", nil)
	}

	if validationErr := validation.ValidatePassword(payload.Password); validationErr != nil {
		var errorCode string

		switch {
		case err.Is(validationErr, domain.ErrPasswordMissingDigit):
			errorCode = "error.password.missing_digit"
		case err.Is(validationErr, domain.ErrPasswordMissingLower):
			errorCode = "error.password.missing_lower"
		case err.Is(validationErr, domain.ErrPasswordMissingUpper):
			errorCode = "error.password.missing_uper"
		case err.Is(validationErr, domain.ErrPasswordMissingSymbol):
			errorCode = "error.password.missing_symbol"
		default:
			errorCode = "error.password.too_short"
		}

		return RegisterUserPayload{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeValidationError, errorCode, nil)
	}

	return payload, nil
}
