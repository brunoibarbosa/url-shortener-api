package pg_repo

import (
	"context"
	"fmt"

	"github.com/brunoibarbosa/url-shortener/internal/domain"
	url_domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	base "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/base"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ListUserURLsRepository struct {
	base.BaseRepository
}

func NewListUserURLsRepository(q *pgxpool.Pool) *ListUserURLsRepository {
	return &ListUserURLsRepository{
		BaseRepository: base.NewBaseRepository(q),
	}
}

func (r *ListUserURLsRepository) ListByUserID(ctx context.Context, userID uuid.UUID, params url_domain.ListURLsParams) ([]url_domain.ListURLsDTO, uint64, error) {
	pagination := r.getPagination(params)
	sort := r.getOrderByField(params)

	var count uint64
	if pagination != "" {
		if err := r.Q(ctx).QueryRow(ctx, `
			SELECT COUNT(u.id)
			FROM urls u
			WHERE u.user_id = $1
		`,
			userID,
		).Scan(&count); err != nil {
			return nil, 0, err
		}
	}

	rows, err := r.Q(ctx).Query(ctx, `
		SELECT short_code, expires_at, created_at
		FROM urls
		WHERE user_id = $1
		`+sort+
		pagination,
		userID,
	)
	if err != nil {
		return nil, 0, err
	}

	urls := []url_domain.ListURLsDTO{}
	for rows.Next() {
		var u url_domain.ListURLsDTO
		if err := rows.Scan(
			&u.ShortCode,
			&u.ExpiresAt,
			&u.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		urls = append(urls, u)
	}

	if count == 0 {
		count = uint64(len(urls))
	}

	return urls, count, nil
}

func (*ListUserURLsRepository) getOrderByField(p url_domain.ListURLsParams) string {
	if p.SortBy == url_domain.ListURLsSortByNone {
		return ""
	}

	sortBy := ""
	switch p.SortBy {
	case url_domain.ListURLsSortByCreatedAt:
		sortBy = "created_at"
	case url_domain.ListURLsSortByExpiresAt:
		sortBy = "expires_at"
	}

	sortKind := ""
	switch p.SortKind {
	case domain.SortDesc:
		sortKind = "DESC"
	default:
		sortKind = "ASC"
	}

	return fmt.Sprintf(" ORDER BY %s %s", sortBy, sortKind)
}

func (*ListUserURLsRepository) getPagination(p url_domain.ListURLsParams) string {
	if p.Pagination.Size <= 0 {
		return ""
	}

	offset := p.Pagination.Size * (p.Pagination.Number - 1)
	pgn := fmt.Sprintf(" LIMIT %d OFFSET %d", p.Pagination.Size, offset)

	return pgn
}
