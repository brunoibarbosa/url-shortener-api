package base

import (
	"context"

	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
)

type BaseRepository struct {
	q pg.Querier
}

func NewBaseRepository(q pg.Querier) BaseRepository {
	return BaseRepository{q: q}
}

func (r BaseRepository) Q(ctx context.Context) pg.Querier {
	return pg.GetQuerier(ctx, r.q)
}
