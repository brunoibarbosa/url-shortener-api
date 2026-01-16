package pg_repo

import (
	"context"
	"database/sql"
	"errors"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	base "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/base"
	"github.com/google/uuid"
)

type SessionRepository struct {
	base.BaseRepository
}

func NewSessionRepository(q pg.Querier) *SessionRepository {
	return &SessionRepository{
		BaseRepository: base.NewBaseRepository(q),
	}
}

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	return r.Q(ctx).QueryRow(
		ctx,
		"INSERT INTO sessions (user_id, refresh_token_hash, user_agent, ip_address, expires_at) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at",
		s.UserID, s.RefreshTokenHash, s.UserAgent, s.IPAddress, s.ExpiresAt,
	).Scan(&s.ID, &s.CreatedAt)
}

func (r *SessionRepository) FindByRefreshToken(ctx context.Context, hash string) (*domain.Session, error) {
	row := r.Q(ctx).QueryRow(
		ctx,
		`SELECT id, user_id, refresh_token_hash, user_agent, ip_address, created_at, expires_at, revoked_at 
		 FROM sessions 
		 WHERE refresh_token_hash=$1`,
		hash,
	)

	s := &domain.Session{}
	if err := row.Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.UserAgent, &s.IPAddress, &s.CreatedAt, &s.ExpiresAt, &s.RevokedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return s, nil
}

func (r *SessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.Q(ctx).Exec(ctx, "UPDATE sessions SET revoked_at = NOW() WHERE id = $1", id)
	return err
}
