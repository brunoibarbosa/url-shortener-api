package query

import (
	"context"

	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
)

type ListSessionsHandler struct {
	repo session_domain.SessionQueryRepository
}

func NewListSessionsHandler(repo session_domain.SessionQueryRepository) *ListSessionsHandler {
	return &ListSessionsHandler{
		repo: repo,
	}
}

func (h *ListSessionsHandler) Handle(ctx context.Context, params session_domain.ListSessionsParams) ([]session_domain.ListSessionsDTO, uint64, error) {
	return h.repo.List(ctx, params)
}
