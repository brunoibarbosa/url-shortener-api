package http_handler

import (
	"context"
	"net/http"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	user_domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	http_handler "github.com/brunoibarbosa/url-shortener/internal/server/http/handler"
	http_middleware "github.com/brunoibarbosa/url-shortener/internal/server/http/middleware"
	"github.com/brunoibarbosa/url-shortener/pkg/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type DeleteURLHTTPHandler struct {
	cmd *command.DeleteURLHandler
}

func NewDeleteURLHTTPHandler(cmd *command.DeleteURLHandler) *DeleteURLHTTPHandler {
	return &DeleteURLHTTPHandler{
		cmd: cmd,
	}
}

func (h *DeleteURLHTTPHandler) Handle(w http.ResponseWriter, r *http.Request) *http_handler.HTTPError {
	ctx := r.Context()

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.url.missing_id", nil)
	}

	id, parseErr := uuid.Parse(idStr)
	if parseErr != nil {
		return http_handler.NewI18nHTTPError(ctx, http.StatusBadRequest, errors.CodeBadRequest, "error.url.invalid_id", nil)
	}

	userID, err := extractUserID(ctx)
	if err != nil {
		return http_handler.NewI18nHTTPError(ctx, http.StatusUnauthorized, errors.CodeUnauthorized, "error.auth.unauthorized", nil)
	}

	appCmd := command.DeleteURLCommand{
		ID:     id,
		UserID: userID,
	}

	if handleErr := h.cmd.Handle(ctx, appCmd); handleErr != nil {
		return http_handler.NewI18nHTTPError(ctx, http.StatusInternalServerError, errors.CodeInternalError, "error.url.delete_failed", nil)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func extractUserID(ctx context.Context) (uuid.UUID, error) {
	userIDValue := ctx.Value(http_middleware.UserIDKey)
	if userIDValue == nil {
		return uuid.Nil, user_domain.ErrUserNotAuthenticated
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return uuid.Nil, user_domain.ErrInvalidUserIDContext
	}

	return userID, nil
}
