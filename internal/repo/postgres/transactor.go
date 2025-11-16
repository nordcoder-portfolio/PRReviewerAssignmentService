//go:build !mockery

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ Transactor = (*transactor)(nil)

type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
	Querier(ctx context.Context) Querier
}

type transactor struct {
	pool *pgxpool.Pool
}

func NewTransactor(pool *pgxpool.Pool) *transactor {
	return &transactor{pool: pool}
}

type ctxKeyTx struct{}

func (t *transactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(ctxKeyTx{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	ctxWithTx := context.WithValue(ctx, ctxKeyTx{}, tx)

	if err := fn(ctxWithTx); err != nil {
		if rbErr := tx.Rollback(ctxWithTx); rbErr != nil {
			return fmt.Errorf("rollback tx after err: %w", err)
		}
		return err
	}

	if err := tx.Commit(ctxWithTx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (t *transactor) Querier(ctx context.Context) Querier {
	if tx, ok := ctx.Value(ctxKeyTx{}).(pgx.Tx); ok {
		return tx
	}
	return t.pool
}
