package application

import "context"

// NoopUnitOfWork выполняет функцию без реальной транзакции.
// Для unit-тестов и read-only сценариев.
type NoopUnitOfWork struct{}

// WithinTx вызывает fn с исходным контекстом.
func (NoopUnitOfWork) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
