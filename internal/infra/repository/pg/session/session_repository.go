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
	db *pg.Postgres
}

func NewSessionRepository(pg *pg.Postgres) *SessionRepository {
	return &SessionRepository{
		db: pg,
	}
}

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	tx, err := r.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	err = tx.QueryRow(
		ctx,
		"INSERT INTO sessions (user_id, refresh_token_hash, user_agent, ip_address, expires_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		s.UserID, s.RefreshTokenHash, s.UserAgent, s.IPAddress, s.ExpiresAt,
	).Scan(&s.ID)

	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (r *SessionRepository) FindByRefreshToken(ctx context.Context, hash string) (*domain.Session, error) {
	row := r.db.Pool.QueryRow(ctx,
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
	tx, err := r.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, "UPDATE sessions SET revoked_at = NOW() WHERE id = $1", id)

	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
