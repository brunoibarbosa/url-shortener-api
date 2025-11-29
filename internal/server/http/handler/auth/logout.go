package handler

import (
	"context"
	err "errors"
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	sd "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
)

type LogoutPayload struct {
	RefreshToken string
}

type LogoutHTTPHandler struct {
	cmd *command.LogoutHandler
}

func NewLogoutHTTPHandler(cmd *command.LogoutHandler) *LogoutHTTPHandler {
	return &LogoutHTTPHandler{
		cmd,
	}
}

func (h *LogoutHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) (http_handler.HandlerResponse, *http_handler.HTTPError) {
	ctx := r.Context()

	payload, validationErr := validateLogoutPayload(r, ctx)
	if validationErr != nil {
		return nil, validationErr
	}

	appCmd := command.LogoutCommand{
		RefreshToken: payload.RefreshToken,
	}
	handleErr := h.cmd.Handle(r.Context(), appCmd)
	if handleErr != nil {
		switch {
		case err.Is(handleErr, sd.ErrInvalidRefreshToken):
			return nil, http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.session.invalid_refresh_token", nil)
		case err.Is(handleErr, sd.ErrTokenGenerate):
			return nil, http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.session.generate_refresh_token", nil)
		default:
			return nil, http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.server.internal", nil)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
	return nil, nil
}

func validateLogoutPayload(r *http.Request, ctx context.Context) (LogoutPayload, *http_handler.HTTPError) {
	var payload = LogoutPayload{
		RefreshToken: "",
	}

	if cookie, handleErr := r.Cookie("refresh_token"); handleErr == nil {
		payload.RefreshToken = cookie.Value
	}

	if payload.RefreshToken == "" {
		return LogoutPayload{}, http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeValidationError, "error.session.missing_refresh_token", nil)
	}

	return payload, nil
}
