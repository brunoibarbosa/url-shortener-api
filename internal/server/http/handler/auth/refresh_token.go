package handler

import (
	"context"
	"encoding/json"
	err "errors"
	"net/http"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	sd "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
)

type RefreshTokenPayload struct {
	RefreshToken string
}

type RefreshToken200Response struct {
	AccessToken string `json:"accessToken"`
}

type RefreshTokenHTTPHandler struct {
	cmd                  *command.RefreshTokenHandler
	refreshTokenDuration time.Duration
}

func NewRefreshTokenHTTPHandler(cmd *command.RefreshTokenHandler, refreshTokenDuration time.Duration) *RefreshTokenHTTPHandler {
	return &RefreshTokenHTTPHandler{
		cmd,
		refreshTokenDuration,
	}
}

func (h *RefreshTokenHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) *http_handler.HTTPError {
	ctx := r.Context()

	payload, validationErr := validateRefreshTokenPayload(r, ctx)
	if validationErr != nil {
		return validationErr
	}

	appCmd := command.RefreshTokenCommand{
		RefreshToken: payload.RefreshToken,
		UserAgent:    r.UserAgent(),
		IPAddress:    r.RemoteAddr,
	}
	response, handleErr := h.cmd.Handle(r.Context(), appCmd)
	if handleErr != nil {
		switch {
		case err.Is(handleErr, sd.ErrInvalidRefreshToken):
			return http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.session.invalid_refresh_token", nil)
		case err.Is(handleErr, sd.ErrTokenGenerate):
			return http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.session.generate_refresh_token", nil)
		default:
			return http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.server.internal", nil)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    response.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(h.refreshTokenDuration.Seconds()),
	})

	responseBody := RefreshToken200Response{
		AccessToken: response.AccessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encodeErr := json.NewEncoder(w).Encode(responseBody); encodeErr != nil {
		return http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.common.encode_failed", nil)
	}

	return nil
}

func validateRefreshTokenPayload(r *http.Request, ctx context.Context) (RefreshTokenPayload, *http_handler.HTTPError) {
	var payload = RefreshTokenPayload{
		RefreshToken: "",
	}

	if cookie, handleErr := r.Cookie("refresh_token"); handleErr == nil {
		payload.RefreshToken = cookie.Value
	}

	if payload.RefreshToken == "" {
		return RefreshTokenPayload{}, http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeValidationError, "error.session.missing_refresh_token", nil)
	}

	return payload, nil
}
