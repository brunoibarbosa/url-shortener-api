package pg_repo

import (
	"context"
	"database/sql"
	"errors"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SessionRepository struct {
	db pg.Querier
}

func NewSessionRepository(postgres *pg.Postgres) *SessionRepository {
	return &SessionRepository{
		db: postgres.Pool,
	}
}

func (r *SessionRepository) WithTx(tx pgx.Tx) domain.SessionRepository {
	return &SessionRepository{
		db: tx,
	}
}

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	return r.db.QueryRow(
		ctx,
		"INSERT INTO sessions (user_id, refresh_token_hash, user_agent, ip_address, expires_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		s.UserID, s.RefreshTokenHash, s.UserAgent, s.IPAddress, s.ExpiresAt,
	).Scan(&s.ID)
}

func (r *SessionRepository) FindByRefreshToken(ctx context.Context, hash string) (*domain.Session, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, user_id, refresh_token_hash, user_agent, ip_address, expires_at, revoked_at 
		 FROM sessions 
		 WHERE refresh_token_hash=$1`,
		hash,
	)

	s := &domain.Session{}
	if err := row.Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.UserAgent, &s.IPAddress, &s.ExpiresAt, &s.RevokedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("not found")
		}
		return nil, err
	}

	return s, nil
}

func (r *SessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "UPDATE sessions SET revoked_at = NOW() WHERE id = $1", id)
	return err
}
