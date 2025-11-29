package pg_repo

import (
	"context"
	"errors"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	base "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/base"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	base.BaseRepository
}

func NewUserRepository(q pg.Querier) *UserRepository {
	return &UserRepository{
		BaseRepository: base.NewBaseRepository(q),
	}
}

func (r *UserRepository) Exists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.Q(ctx).QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	return exists, err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u = domain.User{}
	var pf = domain.UserProfile{}

	err := r.Q(ctx).QueryRow(ctx, `
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
		WHERE u.id=$1
	`, id).Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt, &pf.ID, &pf.Name, &pf.AvatarURL)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if pf.ID != 0 {
		u.Profile = &pf
	}

	return &u, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u = domain.User{}
	var pf = domain.UserProfile{}

	err := r.Q(ctx).QueryRow(ctx, `
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

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if pf.ID != 0 {
		u.Profile = &pf
	}

	return &u, err
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	return r.Q(ctx).QueryRow(ctx, "INSERT INTO users (email) VALUES ($1) RETURNING id, created_at", u.Email).Scan(&u.ID, &u.CreatedAt)
}
