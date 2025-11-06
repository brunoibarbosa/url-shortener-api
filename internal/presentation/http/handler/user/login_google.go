package handler

import (
	"encoding/json"
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/app/user/command"
	"github.com/brunoibarbosa/url-shortener/internal/presentation/http/handler"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
)

type LoginGoogle200Response struct {
	AccessToken string `json:"accessToken"`
}

type LoginGoogleHTTPHandler struct {
	cmd *command.LoginGoogleHandler
}

func NewLoginGoogleHTTPHandler(cmd *command.LoginGoogleHandler) *LoginGoogleHTTPHandler {
	return &LoginGoogleHTTPHandler{
		cmd: cmd,
	}
}

func (h *LoginGoogleHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) (handler.HandlerResponse, *handler.HTTPError) {
	ctx := r.Context()

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return nil, handler.NewHTTPError(http.StatusBadRequest, "missing code", "erro")
	}

	token, err := h.cmd.Handle(ctx, code)
	if err != nil {
		http.Error(w, "failed to login", http.StatusUnauthorized)
		return nil, handler.NewHTTPError(http.StatusUnauthorized, "failed to login", "error")
	}

	response := LoginGoogle200Response{
		AccessToken: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		return nil, handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.common.encode_failed", nil)
	}

	return nil, nil
}
