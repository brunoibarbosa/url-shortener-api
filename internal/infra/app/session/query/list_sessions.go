package query

import (
	"context"
	"fmt"

	"github.com/brunoibarbosa/url-shortener/internal/app/session/query"
	"github.com/brunoibarbosa/url-shortener/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ListSessionsHandler struct {
	db *pgxpool.Pool
}

func NewListSessionsHandler(pg *pgxpool.Pool) *ListSessionsHandler {
	return &ListSessionsHandler{
		db: pg,
	}
}

func (h *ListSessionsHandler) Handle(ctx context.Context, p query.ListSessionsParams) ([]query.ListSessionsDTO, uint64, error) {
	pagination := h.getPagination(p)
	sort := h.getOrderByField(p)

	var count uint64
	if pagination != "" {
		if err := h.db.QueryRow(ctx, `
			SELECT COUNT(s.id)
			FROM sessions s
			WHERE s.revoked_at IS NULL
		`,
		).Scan(&count); err != nil {
			return nil, 0, err
		}
	}

	rows, err := h.db.Query(ctx, `
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

	sessions := []query.ListSessionsDTO{}
	for rows.Next() {
		var s query.ListSessionsDTO
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

func (*ListSessionsHandler) getOrderByField(p query.ListSessionsParams) string {
	sortBy := ""
	switch p.SortBy {
	case query.ListSessionsSortByUserAgent:
		sortBy = "user_agent"
	case query.ListSessionsSortByIPAddress:
		sortBy = "ip_address"
	case query.ListSessionsSortByCreatedAt:
		sortBy = "created_at"
	case query.ListSessionsSortByExpiresAt:
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

func (*ListSessionsHandler) getPagination(p query.ListSessionsParams) string {
	if p.Pagination.Size <= 0 {
		return ""
	}

	offset := p.Pagination.Size * (p.Pagination.Number - 1)
	pgn := fmt.Sprintf(" LIMIT %d OFFSET %d", p.Pagination.Size, offset)

	return pgn
}
