package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
)

type LoginGoogle200Response struct {
	AccessToken string `json:"accessToken"`
}

type LoginGoogleHTTPHandler struct {
	cmd                  *command.LoginGoogleHandler
	refreshTokenDuration time.Duration
}

func NewLoginGoogleHTTPHandler(cmd *command.LoginGoogleHandler, refreshTokenDuration time.Duration) *LoginGoogleHTTPHandler {
	return &LoginGoogleHTTPHandler{
		cmd,
		refreshTokenDuration,
	}
}

func (h *LoginGoogleHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) (handler.HandlerResponse, *handler.HTTPError) {
	ctx := r.Context()

	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.login.failed", handler.Detail(ctx, "code", "error.details.field_required"))
	}

	appCmd := command.LoginGoogleCommand{
		Code:      code,
		UserAgent: r.UserAgent(),
		IPAddress: r.RemoteAddr,
	}

	accessToken, refreshToken, err := h.cmd.Handle(ctx, appCmd)
	if err != nil {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.login.failed", nil)
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

	response := LoginGoogle200Response{
		AccessToken: accessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.common.encode_failed", nil)
	}

	return nil, nil
}
