package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler"
	pkg_errors "github.com/brunoibarbosa/url-shortener/pkg/errors"
)

type LoginGooglePayload struct {
	Code  string
	State string
}

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

func (h *LoginGoogleHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) (http_handler.HandlerResponse, *http_handler.HTTPError) {
	ctx := r.Context()

	payload, validationErr := validateLoginPayload(r, ctx)
	if validationErr != nil {
		return nil, validationErr
	}

	appCmd := command.LoginGoogleCommand{
		Code:      payload.Code,
		State:     payload.State,
		UserAgent: r.UserAgent(),
		IPAddress: r.RemoteAddr,
	}

	accessToken, refreshToken, err := h.cmd.Handle(ctx, appCmd)
	if err != nil {
		switch {
		case errors.Is(err, session_domain.ErrInvalidState):
			return nil, http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, pkg_errors.CodeBadRequest, "error.session.invalid_state", nil)
		default:
			return nil, http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, pkg_errors.CodeInternalError, "error.login.failed", nil)
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

	response := LoginGoogle200Response{
		AccessToken: accessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		return nil, http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, pkg_errors.CodeInternalError, "error.common.encode_failed", nil)
	}

	return nil, nil
}

func validateLoginPayload(r *http.Request, ctx context.Context) (LoginGooglePayload, *http_handler.HTTPError) {
	var payload = LoginGooglePayload{
		Code:  "",
		State: "",
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		return LoginGooglePayload{}, http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, pkg_errors.CodeValidationError, "error.login.failed", http_handler.Detail(ctx, "code", "error.details.field_required"))
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		return LoginGooglePayload{}, http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, pkg_errors.CodeValidationError, "error.login.failed", http_handler.Detail(ctx, "state", "error.details.field_required"))
	}

	return payload, nil
}
