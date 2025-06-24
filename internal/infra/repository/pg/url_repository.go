package pg_repo

import (
	"context"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
)

type URLRepository struct {
	db *pg.Postgres
}

func NewURLRepository(pg *pg.Postgres) *URLRepository {
	return &URLRepository{
		db: pg,
	}
}

func (r *URLRepository) Exists(ctx context.Context, shortCode string) (bool, error) {
	var exists bool
	err := r.db.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)", shortCode).Scan(&exists)
	return exists, err
}

func (r *URLRepository) Save(ctx context.Context, u *url.URL) error {
	_, err := r.db.Pool.Exec(ctx, "INSERT INTO urls (short_code, encrypted_url, expires_at) VALUES ($1, $2, $3)", u.ShortCode, u.EncryptedURL, u.ExpiresAt.UTC())
	return err
}

func (r *URLRepository) FindByShortCode(ctx context.Context, shortCode string) (*url.URL, error) {
	u := url.URL{
		ShortCode:    shortCode,
		EncryptedURL: "",
		ExpiresAt:    nil,
	}
	err := r.db.Pool.QueryRow(ctx, "SELECT encrypted_url, expires_at FROM urls WHERE short_code = $1 LIMIT 1", shortCode).Scan(&u.EncryptedURL, &u.ExpiresAt)

	if err != nil {
		return nil, err
	}

	if u.ExpiresAt != nil && time.Now().UTC().After(u.ExpiresAt.UTC()) {
		err := url.ErrExpiredURL
		return nil, err
	}

	return &u, nil
}

func (r *URLRepository) DeleteExpiredURLs(ctx context.Context) (int64, error) {
	query := `DELETE FROM urls WHERE expires_at IS NOT NULL AND expires_at < now()`
	result, err := r.db.Pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}

	rows := result.RowsAffected()
	return rows, nil
}
