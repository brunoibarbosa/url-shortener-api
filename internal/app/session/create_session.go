package command

import (
	"context"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/google/uuid"
)

type CreateSessionCommand struct {
	UserID       uuid.UUID
	RefreshToken uuid.UUID
	UserAgent    *string
	IPAddress    *string
	ExpiresAt    *time.Time
}

type CreateSessionHandler struct {
	repo domain.SessionRepository
}

func NewCreateSession(repo domain.SessionRepository) *CreateSessionHandler {
	return &CreateSessionHandler{
		repo,
	}
}

func (h *CreateSessionHandler) Handle(ctx context.Context, cmd CreateSessionCommand) error {
	s := &domain.UserSession{
		UserID:       cmd.UserID,
		UserAgent:    cmd.UserAgent,
		IPAddress:    cmd.IPAddress,
		RefreshToken: cmd.RefreshToken,
		ExpiresAt:    cmd.ExpiresAt,
	}

	return h.repo.Create(ctx, s)
}
