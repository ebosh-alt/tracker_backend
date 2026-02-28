package steps

import (
	"context"
	"tracker/internal/domain/shared"
)

type Repository interface {
	GetByDateRange(ctx context.Context, userID int64, from shared.LocalDate, to shared.LocalDate) ([]DailyEntry, error)
	UpsertByDate(ctx context.Context, entry DailyEntry) (DailyEntry, error)
	AddDelta(ctx context.Context, userID int64, date shared.LocalDate, delta int, source Source) (DailyEntry, error)
	DeleteByDate(ctx context.Context, userID int64, date shared.LocalDate) error
}

type AnalyticsUserSettings struct {
	Timezone  string
	StepsGoal int
}

type AnalyticsUserReader interface {
	GetSettings(ctx context.Context, userID int64) (AnalyticsUserSettings, error)
}
