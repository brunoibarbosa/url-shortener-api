package handler

import (
	"context"
	"encoding/json"
	err "errors"
	"io"
	"net/http"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/user/command"
	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler"
	"github.com/brunoibarbosa/url-shortener/internal/validation"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
	"github.com/google/uuid"
)

type LoginUserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserPayload struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

type LoginUser200Response struct {
	AccessToken string `json:"accessToken"`
}

type LoginUserHTTPHandler struct {
	cmd *command.LoginUserHandler
}

func NewLoginUserHTTPHandler(cmd *command.LoginUserHandler) *LoginUserHTTPHandler {
	return &LoginUserHTTPHandler{
		cmd: cmd,
	}
}

func (h *LoginUserHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) (handler.HandlerResponse, *handler.HTTPError) {
	ctx := r.Context()

	payload, validationErr := validateLoginUserPayload(r, ctx)
	if validationErr != nil {
		return nil, validationErr
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
			return nil, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeValidationError, "error.login.invalid_credentials", nil)
		case err.Is(handleErr, domain.ErrSocialLoginOnly):
			return nil, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeValidationError, "error.login.invalid_credentials", nil)
		default:
			return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.login.failed", nil)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/auth/refresh",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(30 * 24 * time.Hour / time.Second),
	})

	response := LoginUser200Response{
		AccessToken: accessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.common.encode_failed", nil)
	}

	return nil, nil
}

func validateLoginUserPayload(r *http.Request, ctx context.Context) (LoginUserPayload, *handler.HTTPError) {
	var payload LoginUserPayload
	decodeErr := json.NewDecoder(r.Body).Decode(&payload)

	if err.Is(decodeErr, io.EOF) {
		return LoginUserPayload{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.common.empty_body", nil)
	}

	if validationErr := validation.ValidateEmail(payload.Email); validationErr != nil {
		return LoginUserPayload{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeValidationError, "error.email.invalid_email_format", nil)
	}

	if len(payload.Password) == 0 {
		return LoginUserPayload{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeValidationError, "error.password.required", nil)
	}

	return payload, nil
}
