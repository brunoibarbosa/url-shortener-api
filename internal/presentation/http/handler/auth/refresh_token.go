package handler

import (
	"context"
	"encoding/json"
	err "errors"
	"net/http"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	sd "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
)

type RefreshTokenPayload struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshToken200Response struct {
	AccessToken string `json:"accessToken"`
}

type RefreshTokenHTTPHandler struct {
	cmd *command.RefreshTokenHandler
}

func NewRefreshTokenHTTPHandler(cmd *command.RefreshTokenHandler) *RefreshTokenHTTPHandler {
	return &RefreshTokenHTTPHandler{
		cmd: cmd,
	}
}

func (h *RefreshTokenHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) (handler.HandlerResponse, *handler.HTTPError) {
	ctx := r.Context()

	payload, validationErr := validateRefreshTokenPayload(r, ctx)
	if validationErr != nil {
		return nil, validationErr
	}

	appCmd := command.RefreshTokenCommand{
		RefreshToken: payload.RefreshToken,
		UserAgent:    r.UserAgent(),
		IPAddress:    r.RemoteAddr,
	}
	accessToken, refreshToken, handleErr := h.cmd.Handle(r.Context(), appCmd)
	if handleErr != nil {
		switch {
		case err.Is(handleErr, sd.ErrInvalidRefreshToken):
			return nil, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.login.invalid_refresh_token", nil)
		case err.Is(handleErr, sd.ErrTokenGenerate):
			return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.session.generate_refresh_token", nil)
		default:
			return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.server.internal", nil)
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

	response := RefreshToken200Response{
		AccessToken: accessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.common.encode_failed", nil)
	}

	return nil, nil
}

func validateRefreshTokenPayload(r *http.Request, ctx context.Context) (RefreshTokenPayload, *handler.HTTPError) {
	var payload = RefreshTokenPayload{
		RefreshToken: "",
	}

	if cookie, handleErr := r.Cookie("refresh_token"); handleErr == nil {
		payload.RefreshToken = cookie.Value
	}

	if payload.RefreshToken == "" {
		return RefreshTokenPayload{}, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeValidationError, "error.session.missing_refresh_token", nil)
	}

	return payload, nil
}
