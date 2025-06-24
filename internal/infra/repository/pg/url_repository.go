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
	_, err := r.db.Pool.Exec(ctx, "INSERT INTO urls (short_code, encrypted_url, expires_at) VALUES ($1, $2, $3)", u.ShortCode, u.EncryptedURL, time.Now().Add(24*time.Hour))
	return err
}

func (r *URLRepository) FindByShortCode(ctx context.Context, shortCode string) (*url.URL, error) {
	u := url.URL{
		ShortCode:    shortCode,
		EncryptedURL: "",
	}
	err := r.db.Pool.QueryRow(ctx, "SELECT encrypted_url FROM urls WHERE short_code = $1 LIMIT 1", shortCode).Scan(&u.EncryptedURL)

	if err != nil {
		return nil, err
	}

	return &u, nil
}
