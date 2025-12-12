package pg_repo

import (
	"context"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	base "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/base"
)

type URLRepository struct {
	base.BaseRepository
}

func NewURLRepository(q pg.Querier) *URLRepository {
	return &URLRepository{
		BaseRepository: base.NewBaseRepository(q),
	}
}

func (r *URLRepository) Exists(ctx context.Context, shortCode string) (bool, error) {
	var exists bool
	err := r.Q(ctx).QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)", shortCode).Scan(&exists)
	return exists, err
}

func (r *URLRepository) Save(ctx context.Context, u *domain.URL) error {
	_, err := r.Q(ctx).Exec(ctx, "INSERT INTO urls (short_code, encrypted_url, expires_at) VALUES ($1, $2, $3)", u.ShortCode, u.EncryptedURL, u.ExpiresAt.UTC())
	return err
}

func (r *URLRepository) FindByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	u := domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: "",
		ExpiresAt:    nil,
	}
	err := r.Q(ctx).QueryRow(ctx, "SELECT encrypted_url, expires_at FROM urls WHERE short_code = $1 LIMIT 1", shortCode).Scan(&u.EncryptedURL, &u.ExpiresAt)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *URLRepository) DeleteExpiredURLs(ctx context.Context) (int64, error) {
	query := `DELETE FROM urls WHERE expires_at IS NOT NULL AND expires_at < now()`
	result, err := r.Q(ctx).Exec(ctx, query)
	if err != nil {
		return 0, err
	}

	rows := result.RowsAffected()
	return rows, nil
}
