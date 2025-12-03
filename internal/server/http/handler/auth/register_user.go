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
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler"
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
		cmd,
	}
}

func (h *RegisterUserHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) *http_handler.HTTPError {
	ctx := r.Context()

	payload, validationErr := validateRegisterPayload(r, ctx)
	if validationErr != nil {
		return validationErr
	}

	appCmd := command.RegisterUserCommand{Email: payload.Email, Password: payload.Password, Name: payload.Name}
	user, handleErr := h.cmd.Handle(r.Context(), appCmd)
	if handleErr != nil {
		switch {
		case err.Is(handleErr, domain.ErrEmailAlreadyExists):
			return http_handler.NewI18nHTTPError(ctx, http.StatusConflict, errors.CodeValidationError, "error.validation.failed", http_handler.Detail(ctx, "email", "error.details.email.already_exists"))
		default:
			return http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.user.create_failed", nil)
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
		return http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.common.encode_failed", nil)
	}

	return nil
}

func validateRegisterPayload(r *http.Request, ctx context.Context) (RegisterUserPayload, *http_handler.HTTPError) {
	var payload RegisterUserPayload
	decodeErr := json.NewDecoder(r.Body).Decode(&payload)

	if err.Is(decodeErr, io.EOF) {
		return RegisterUserPayload{}, http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.common.empty_body", nil)
	}

	ec := http_handler.NewErrorCollector(ctx)

	if payload.Email == "" {
		ec.AddFieldError("email", "error.details.email.invalid_format")
	} else if validationErr := validation.ValidateEmail(payload.Email); validationErr != nil {
		ec.AddFieldError("email", "error.details.email.invalid_format")
	}

	if payload.Password == "" {
		ec.AddFieldError("password", "error.details.field_required")
	} else if validationErr := validation.ValidatePassword(payload.Password); validationErr != nil {
		var errorCode string

		switch {
		case err.Is(validationErr, domain.ErrPasswordMissingDigit):
			errorCode = "error.details.password.missing_digit"
		case err.Is(validationErr, domain.ErrPasswordMissingLower):
			errorCode = "error.details.password.missing_lower"
		case err.Is(validationErr, domain.ErrPasswordMissingUpper):
			errorCode = "error.details.password.missing_upper"
		case err.Is(validationErr, domain.ErrPasswordMissingSymbol):
			errorCode = "error.details.password.missing_symbol"
		default:
			errorCode = "error.details.password.too_short"
		}

		ec.AddFieldError("password", errorCode)
	}

	if payload.Name == "" {
		ec.AddFieldError("name", "error.details.field_required")
	}

	if ec.HasErrors() {
		return RegisterUserPayload{}, ec.ToHTTPError(http.StatusBadRequest, errors.CodeValidationError, "error.validation.failed")
	}

	return payload, nil
}
