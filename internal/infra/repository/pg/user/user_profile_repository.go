package pg_repo

import (
	"context"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	base "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/base"
	"github.com/google/uuid"
)

type UserProfileRepository struct {
	base.BaseRepository
}

func NewUserProfileRepository(q pg.Querier) *UserProfileRepository {
	return &UserProfileRepository{
		BaseRepository: base.NewBaseRepository(q),
	}
}

func (r *UserProfileRepository) Create(ctx context.Context, userID uuid.UUID, pv *domain.UserProfile) error {
	_, err := r.Q(ctx).Exec(ctx,
		"INSERT INTO user_profiles (user_id, name, avatar_url) VALUES ($1, $2, $3)",
		userID, pv.Name, pv.AvatarURL,
	)
	return err
}
