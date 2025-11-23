package pg_repo

import (
	"context"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserProfileRepository struct {
	db pg.Querier
}

func NewUserProfileRepository(postgres *pg.Postgres) *UserProfileRepository {
	return &UserProfileRepository{
		db: postgres.Pool,
	}
}

func (r *UserProfileRepository) WithTx(tx pgx.Tx) domain.UserProfileRepository {
	return &UserProfileRepository{
		db: tx,
	}
}

func (r *UserProfileRepository) Create(ctx context.Context, userID uuid.UUID, pv *domain.UserProfile) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO user_profiles (user_id, name, avatar_url) VALUES ($1, $2, $3)",
		userID, pv.Name, pv.AvatarURL,
	)
	return err
}
