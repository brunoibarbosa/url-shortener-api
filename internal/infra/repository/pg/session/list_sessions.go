package pg_repo

import (
	"context"
	"fmt"

	"github.com/brunoibarbosa/url-shortener/internal/domain"
	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ListSessionsRepository struct {
	db *pgxpool.Pool
}

func NewListSessionsRepository(pg *pgxpool.Pool) *ListSessionsRepository {
	return &ListSessionsRepository{
		db: pg,
	}
}

func (r *ListSessionsRepository) List(ctx context.Context, p session_domain.ListSessionsParams) ([]session_domain.ListSessionsDTO, uint64, error) {
	pagination := r.getPagination(p)
	sort := r.getOrderByField(p)

	var count uint64
	if pagination != "" {
		if err := r.db.QueryRow(ctx, `
			SELECT COUNT(s.id)
			FROM sessions s
			WHERE s.revoked_at IS NULL
		`,
		).Scan(&count); err != nil {
			return nil, 0, err
		}
	}

	rows, err := r.db.Query(ctx, `
		SELECT
			s.user_agent,
			s.ip_address,
			s.created_at,
			s.expires_at
		FROM sessions s
		WHERE s.revoked_at IS NULL
		`+sort+
		pagination,
	)
	if err != nil {
		return nil, 0, err
	}

	sessions := []session_domain.ListSessionsDTO{}
	for rows.Next() {
		var s session_domain.ListSessionsDTO
		if err := rows.Scan(
			&s.UserAgent,
			&s.IPAddress,
			&s.CreatedAt,
			&s.ExpiresAt,
		); err != nil {
			return nil, 0, err
		}

		sessions = append(sessions, s)
	}

	if count == 0 {
		count = uint64(len(sessions))
	}

	return sessions, count, nil
}

func (*ListSessionsRepository) getOrderByField(p session_domain.ListSessionsParams) string {
	sortBy := ""
	switch p.SortBy {
	case session_domain.ListSessionsSortByUserAgent:
		sortBy = "user_agent"
	case session_domain.ListSessionsSortByIPAddress:
		sortBy = "ip_address"
	case session_domain.ListSessionsSortByCreatedAt:
		sortBy = "created_at"
	case session_domain.ListSessionsSortByExpiresAt:
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

func (*ListSessionsRepository) getPagination(p session_domain.ListSessionsParams) string {
	if p.Pagination.Size <= 0 {
		return ""
	}

	offset := p.Pagination.Size * (p.Pagination.Number - 1)
	pgn := fmt.Sprintf(" LIMIT %d OFFSET %d", p.Pagination.Size, offset)

	return pgn
}
