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
	repo      domain.URLRepository
	cacheRepo domain.URLCacheRepository
}

func NewDeleteURLHandler(repo domain.URLRepository, cacheRepo domain.URLCacheRepository) *DeleteURLHandler {
	return &DeleteURLHandler{
		repo:      repo,
		cacheRepo: cacheRepo,
	}
}

func (h *DeleteURLHandler) Handle(ctx context.Context, cmd DeleteURLCommand) error {
	shortCode, err := h.repo.SoftDelete(ctx, cmd.ID, cmd.UserID)
	if err != nil {
		return err
	}

	_ = h.cacheRepo.Delete(ctx, shortCode)

	return nil
}
