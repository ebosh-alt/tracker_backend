package application

import "context"

// UnitOfWork выполняет функцию в рамках транзакции.
// Внутри fn должен передаваться ctx, полученный из WithinTx, чтобы репозитории могли использовать текущий tx.
type UnitOfWork interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
