package medication

import (
	"context"
	"time"

	"tracker/internal/domain/shared"
)

// MedicationRepository определяет CRUD-операции для карточек БАД.
type MedicationRepository interface {
	ListByUser(ctx context.Context, userID int64) ([]Medication, error)
	GetByID(ctx context.Context, userID int64, medicationID int64) (Medication, error)
	Create(ctx context.Context, medication Medication) (Medication, error)
	Update(ctx context.Context, medication Medication) (Medication, error)
	Delete(ctx context.Context, userID int64, medicationID int64) error
}

// LogFilter задает параметры поиска логов приема БАД.
type LogFilter struct {
	From         time.Time
	To           time.Time
	MedicationID *int64
	Status       *Status
	Limit        int
	Offset       int
}

// LogRepository определяет операции над журналом приемов БАД.
type LogRepository interface {
	List(ctx context.Context, userID int64, filter LogFilter) ([]Log, error)
	Create(ctx context.Context, log Log) (Log, error)
	UpdateStatus(ctx context.Context, userID int64, logID int64, next Status) (Log, error)
	EnsureRange(ctx context.Context, userID int64, from time.Time, to time.Time) error
}

// SupplementQueryService обслуживает read-model endpoint'ы today/calendar/day.
type SupplementQueryService interface {
	Today(ctx context.Context, userID int64, date shared.LocalDate) (SupplementDayDetails, error)
	Day(ctx context.Context, userID int64, date shared.LocalDate) (SupplementDayDetails, error)
	Calendar(ctx context.Context, userID int64, month shared.LocalMonth) (map[shared.LocalDate]map[Category]DailyCategorySummary, error)
}
