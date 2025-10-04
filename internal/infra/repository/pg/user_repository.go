package pg_repo

import (
	"context"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/jackc/pgx/v5"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserProfile struct {
	ID        int64
	UserID    int64
	Name      string
	AvatarURL string
	Phone     string
	BirthDate *time.Time
}

type UserProvider struct {
	ID           int64
	UserID       int64
	Provider     string
	ProviderID   string
	AccessToken  string
	RefreshToken string
	CreatedAt    time.Time
}

type UserRepository struct {
	db *pg.Postgres
}

func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash string) (int64, error) {
	var id int64
	err := r.db.Pool.QueryRow(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		email, passwordHash).Scan(&id)
	return id, err
}

func (r *UserRepository) CreateuserWithProvider(ctx context.Context, email, provider, providerID, accessToken, refreshToken string) (int64, error) {
	var id int64
	tx, err := r.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	err = tx.QueryRow(ctx, "INSERT INTO users (email) VALUES ($1) RETURNING id", email).Scan(&id)
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	_, err = tx.Exec(ctx, "INSERT INTO user_providers (user_id, provider, provider_id, access_token, refresh_token) VALUES ($1, $2, $3, $4, $5)",
		id, provider, providerID, accessToken, refreshToken)

	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	return id, tx.Commit(ctx)
}
