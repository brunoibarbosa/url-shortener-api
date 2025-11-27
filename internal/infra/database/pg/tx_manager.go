package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type querierKey struct{}

type TxStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type TxManager struct {
	s TxStarter
}

func NewTxManager(s TxStarter) *TxManager {
	return &TxManager{s: s}
}

func (m *TxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.s.Begin(ctx)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, querierKey{}, tx)

	if err := fn(ctx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func GetQuerier(ctx context.Context, fallback Querier) Querier {
	if q, ok := ctx.Value(querierKey{}).(Querier); ok {
		return q
	}
	return fallback
}
