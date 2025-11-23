package pg_repo

import (
	"context"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db pg.Querier
}

func NewUserRepository(postgres *pg.Postgres) *UserRepository {
	return &UserRepository{
		db: postgres.Pool,
	}
}

func (r *UserRepository) WithTx(tx pgx.Tx) domain.UserRepository {
	return &UserRepository{
		db: tx,
	}
}

func (r *UserRepository) Exists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	return exists, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u = domain.User{}
	var pf = domain.UserProfile{}

	err := r.db.QueryRow(ctx, `
		SELECT 
			u.id, 
			u.email, 
			u.created_at, 
			u.updated_at,
			p.id,
			p.name, 
			p.avatar_url
		FROM users u
        LEFT JOIN user_profiles p ON p.user_id = u.id
		WHERE email=$1
	`, email).Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt, &pf.ID, &pf.Name, &pf.AvatarURL)

	if pf.ID != 0 {
		u.Profile = &pf
	}

	return &u, err
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	return r.db.QueryRow(ctx, "INSERT INTO users (email) VALUES ($1) RETURNING id, created_at", u.Email).Scan(&u.ID, &u.CreatedAt)
}
