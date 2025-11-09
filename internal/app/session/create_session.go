package command

import (
	"context"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/google/uuid"
)

type CreateSessionCommand struct {
	UserID           uuid.UUID
	RefreshTokenHash string
	UserAgent        string
	IPAddress        string
	ExpiresAt        *time.Time
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
	s := &domain.Session{
		UserID:           cmd.UserID,
		UserAgent:        cmd.UserAgent,
		IPAddress:        cmd.IPAddress,
		RefreshTokenHash: cmd.RefreshTokenHash,
		ExpiresAt:        cmd.ExpiresAt,
	}

	return h.repo.Create(ctx, s)
}
