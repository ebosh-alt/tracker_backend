package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txContextKey struct{}

// UnitOfWork Postgres-реализация application.UnitOfWork.
type UnitOfWork struct {
	db *pgxpool.Pool
}

// NewUnitOfWork создает UoW поверх пула Postgres.
func NewUnitOfWork(db *pgxpool.Pool) *UnitOfWork {
	return &UnitOfWork{db: db}
}

// WithinTx запускает функцию внутри транзакции.
func (u *UnitOfWork) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txContextKey{}, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

type dbExecutor interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func executorFromContext(ctx context.Context, db *pgxpool.Pool) dbExecutor {
	tx, ok := ctx.Value(txContextKey{}).(pgx.Tx)
	if ok && tx != nil {
		return tx
	}
	return db
}
