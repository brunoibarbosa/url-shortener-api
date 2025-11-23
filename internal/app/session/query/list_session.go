package query

import (
	"context"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain"
)

type ListSessionsDTO struct {
	UserAgent string
	IPAddress string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type ListSessionsSortBy uint8

const (
	ListSessionsSortByNone ListSessionsSortBy = iota
	ListSessionsSortByUserAgent
	ListSessionsSortByIPAddress
	ListSessionsSortByCreatedAt
	ListSessionsSortByExpiresAt
)

type ListSessionsParams struct {
	SortBy     ListSessionsSortBy
	SortKind   domain.SortKind
	Pagination domain.Pagination
}

type ListSessionsHandler interface {
	Handle(ctx context.Context, p ListSessionsParams) ([]ListSessionsDTO, uint64, error)
}
