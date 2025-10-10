package pg_repo

import (
	"context"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *pg.Postgres
}

func NewUserRepository(pg *pg.Postgres) *UserRepository {
	return &UserRepository{
		db: pg,
	}
}

func (r *UserRepository) Exists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	return exists, err
}

func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	var id uuid.UUID
	var createdAt time.Time

	err := r.db.Pool.QueryRow(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, created_at",
		email, passwordHash).Scan(&id, &createdAt)
	return &domain.User{
		ID:        id,
		Email:     email,
		CreatedAt: createdAt,
	}, err
}

func (r *UserRepository) CreateUserWithProvider(ctx context.Context, email, provider, providerID, accessToken, refreshToken string) (*domain.User, error) {
	var id uuid.UUID
	var createdAt time.Time

	tx, err := r.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	err = tx.QueryRow(ctx, "INSERT INTO users (email) VALUES ($1) RETURNING id, created_at", email).Scan(&id, &createdAt)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	_, err = tx.Exec(ctx, "INSERT INTO user_providers (user_id, provider, provider_id, access_token, refresh_token) VALUES ($1, $2, $3, $4, $5)",
		id, provider, providerID, accessToken, refreshToken)

	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	return &domain.User{
		ID:        id,
		Email:     email,
		CreatedAt: createdAt,
	}, tx.Commit(ctx)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.db.Pool.QueryRow(ctx,
		"SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email=$1", email)

	var u domain.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}

func (r *UserRepository) GetByProvider(ctx context.Context, provider, providerID string) (*domain.User, error) {
	row := r.db.Pool.QueryRow(ctx,
		`SELECT u.id, u.email, u.password_hash, u.created_at, u.updated_at
		 FROM users u
		 INNER JOIN user_providers p ON p.user_id = u.id
		 WHERE p.provider=$1 AND p.provider_id=$2`,
		provider, providerID,
	)

	var u domain.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}
