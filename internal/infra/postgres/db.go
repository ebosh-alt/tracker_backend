package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool создает пул подключений к Postgres.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, databaseURL)
}
