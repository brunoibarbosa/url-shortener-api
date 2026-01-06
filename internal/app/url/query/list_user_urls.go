package query

import (
	"context"

	url_domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/google/uuid"
)

type ListUserURLsHandler struct {
	repo url_domain.URLQueryRepository
}

func NewListUserURLsHandler(repo url_domain.URLQueryRepository) *ListUserURLsHandler {
	return &ListUserURLsHandler{
		repo: repo,
	}
}

func (h *ListUserURLsHandler) Handle(ctx context.Context, userID uuid.UUID, params url_domain.ListURLsParams) ([]url_domain.ListURLsDTO, uint64, error) {
	return h.repo.ListByUserID(ctx, userID, params)
}
