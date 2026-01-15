package pg_repo

import (
	"context"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	base "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/base"
	"github.com/google/uuid"
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
	_, err := r.Q(ctx).Exec(ctx, "INSERT INTO urls (short_code, encrypted_url, user_id, expires_at) VALUES ($1, $2, $3, $4)", u.ShortCode, u.EncryptedURL, u.UserID, u.ExpiresAt.UTC())
	return err
}

func (r *URLRepository) FindByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	u := domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: "",
		UserID:       nil,
		ExpiresAt:    nil,
		DeletedAt:    nil,
	}
	err := r.Q(ctx).QueryRow(ctx, "SELECT encrypted_url, user_id, expires_at, deleted_at FROM urls WHERE short_code = $1 AND deleted_at IS NULL LIMIT 1", shortCode).Scan(&u.EncryptedURL, &u.UserID, &u.ExpiresAt, &u.DeletedAt)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *URLRepository) SoftDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (string, error) {
	var shortCode string
	query := `UPDATE urls SET deleted_at = now() WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL RETURNING short_code`
	err := r.Q(ctx).QueryRow(ctx, query, id, userID).Scan(&shortCode)
	return shortCode, err
}
