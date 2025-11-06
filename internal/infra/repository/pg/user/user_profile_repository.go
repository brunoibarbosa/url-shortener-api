package pg_repo

import (
	"context"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserProfileRepository struct {
	db *pg.Postgres
}

func NewUserProfileRepository(pg *pg.Postgres) *UserProfileRepository {
	return &UserProfileRepository{
		db: pg,
	}
}

func (r *UserProfileRepository) Create(ctx context.Context, userID uuid.UUID, pv *domain.UserProfile) error {
	tx, err := r.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO user_profiles (user_id, name, avatar_url) VALUES ($1, $2, $3) RETURNING id",
		userID, pv.Name, pv.AvatarURL,
	)

	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
