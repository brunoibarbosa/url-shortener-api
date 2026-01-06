package command

import (
	"context"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/google/uuid"
)

type DeleteURLCommand struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

type DeleteURLHandler struct {
	repo domain.URLRepository
}

func NewDeleteURLHandler(repo domain.URLRepository) *DeleteURLHandler {
	return &DeleteURLHandler{
		repo: repo,
	}
}

func (h *DeleteURLHandler) Handle(ctx context.Context, cmd DeleteURLCommand) error {
	return h.repo.SoftDelete(ctx, cmd.ID, cmd.UserID)
}
