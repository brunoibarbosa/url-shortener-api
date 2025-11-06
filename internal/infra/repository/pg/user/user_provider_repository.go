package pg_repo

import (
	"context"
	"errors"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserProviderRepository struct {
	db *pg.Postgres
}

func NewUserProviderRepository(pg *pg.Postgres) *UserProviderRepository {
	return &UserProviderRepository{
		db: pg,
	}
}

func (r *UserProviderRepository) Find(ctx context.Context, provider, providerID string) (*domain.UserProvider, error) {
	row := r.db.Pool.QueryRow(ctx,
		`SELECT id, user_id, password_hash
		 FROM user_providers p
		 WHERE p.provider=$1 AND p.provider_id=$2`,
		provider, providerID,
	)

	p := &domain.UserProvider{
		Provider:   provider,
		ProviderID: providerID,
	}
	err := row.Scan(&p.ID, &p.UserID, &p.PasswordHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return p, nil
}

func (r *UserProviderRepository) Create(ctx context.Context, userID uuid.UUID, pv *domain.UserProvider) error {
	tx, err := r.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	err = tx.QueryRow(ctx,
		"INSERT INTO user_providers (user_id, provider, provider_id, password_hash) VALUES ($1, $2, $3, $4) RETURNING id",
		userID, pv.Provider, pv.ProviderID, pv.PasswordHash,
	).Scan(&pv.ID)

	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
