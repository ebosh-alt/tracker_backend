package dashboard

import (
	"context"

	"tracker/internal/domain/shared"
)

// QueryService описывает read-only сборку данных для /api/dashboard.
type QueryService interface {
	BuildByDate(ctx context.Context, userID int64, date shared.LocalDate) (Model, error)
}
