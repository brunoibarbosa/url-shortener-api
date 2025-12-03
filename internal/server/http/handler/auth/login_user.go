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
)

type LoginUserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUser200Response struct {
	AccessToken string `json:"accessToken"`
}

type LoginUserHTTPHandler struct {
	cmd                  *command.LoginUserHandler
	refreshTokenDuration time.Duration
}

func NewLoginUserHTTPHandler(cmd *command.LoginUserHandler, refreshTokenDuration time.Duration) *LoginUserHTTPHandler {
	return &LoginUserHTTPHandler{
		cmd,
		refreshTokenDuration,
	}
}

func (h *LoginUserHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) *http_handler.HTTPError {
	ctx := r.Context()

	payload, validationErr := validateLoginUserPayload(r, ctx)
	if validationErr != nil {
		return validationErr
	}

	appCmd := command.LoginUserCommand{
		Email:     payload.Email,
		Password:  payload.Password,
		UserAgent: r.UserAgent(),
		IPAddress: r.RemoteAddr,
	}
	accessToken, refreshToken, handleErr := h.cmd.Handle(r.Context(), appCmd)
	if handleErr != nil {
		switch {
		case err.Is(handleErr, domain.ErrInvalidCredentials):
			return http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeValidationError, "error.login.invalid_credentials", nil)
		case err.Is(handleErr, domain.ErrSocialLoginOnly):
			return http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeValidationError, "error.login.invalid_credentials", nil)
		default:
			return http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.login.failed", nil)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(h.refreshTokenDuration.Seconds()),
	})

	response := LoginUser200Response{
		AccessToken: accessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		return http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.common.encode_failed", nil)
	}

	return nil
}

func validateLoginUserPayload(r *http.Request, ctx context.Context) (LoginUserPayload, *http_handler.HTTPError) {
	var payload LoginUserPayload
	decodeErr := json.NewDecoder(r.Body).Decode(&payload)

	if err.Is(decodeErr, io.EOF) {
		return LoginUserPayload{}, http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.common.empty_body", nil)
	}

	ec := http_handler.NewErrorCollector(ctx)

	if payload.Email == "" {
		ec.AddFieldError("email", "error.details.field_required")
	} else if validationErr := validation.ValidateEmail(payload.Email); validationErr != nil {
		ec.AddFieldError("email", "error.details.email.invalid_format")
	}

	if len(payload.Password) == 0 {
		ec.AddFieldError("password", "error.details.field_required")
	}

	if ec.HasErrors() {
		return LoginUserPayload{}, ec.ToHTTPError(http.StatusBadRequest, errors.CodeValidationError, "error.validation.failed")
	}

	return payload, nil
}
