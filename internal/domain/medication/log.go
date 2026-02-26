package medication

import (
	"fmt"
	"time"

	"tracker/internal/domain/shared"
)

// Log отражает единичный запланированный прием с фактическим статусом.
type Log struct {
	ID           int64
	MedicationID int64
	UserID       int64
	ScheduledAt  time.Time
	TakenAt      *time.Time
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewLog создает новый лог приема БАД.
func NewLog(
	medicationID int64,
	userID int64,
	scheduledAt time.Time,
	status Status,
	now time.Time,
) (*Log, error) {
	if medicationID <= 0 {
		return nil, fmt.Errorf("%w: medication id should be positive", shared.ErrInvalidInput)
	}
	if userID <= 0 {
		return nil, fmt.Errorf("%w: user id should be positive", shared.ErrInvalidInput)
	}
	if scheduledAt.IsZero() {
		return nil, fmt.Errorf("%w: scheduledAt is required", shared.ErrInvalidInput)
	}
	normalizedStatus, err := NormalizeStatus(string(status))
	if err != nil {
		return nil, err
	}

	ts := now.UTC()
	log := Log{
		MedicationID: medicationID,
		UserID:       userID,
		ScheduledAt:  scheduledAt.UTC(),
		Status:       normalizedStatus,
		CreatedAt:    ts,
		UpdatedAt:    ts,
	}

	// takenAt должен выставляться автоматически при status=taken.
	if normalizedStatus == StatusTaken {
		takenAt := ts
		log.TakenAt = &takenAt
	}

	return &log, nil
}

// SetStatus выполняет переход статуса с учетом правил домена.
func (l *Log) SetStatus(next Status, now time.Time, allowCorrections bool) error {
	normalizedNext, err := NormalizeStatus(string(next))
	if err != nil {
		return err
	}
	if !l.Status.CanTransitionTo(normalizedNext, allowCorrections) {
		return fmt.Errorf(
			"%w: invalid status transition %q -> %q",
			shared.ErrInvalidInput,
			l.Status,
			normalizedNext,
		)
	}

	l.Status = normalizedNext
	l.UpdatedAt = now.UTC()

	// Требование: при переводе в taken поле takenAt проставляется автоматически.
	if normalizedNext == StatusTaken {
		takenAt := now.UTC()
		l.TakenAt = &takenAt
	}
	return nil
}

// LogKey служит для идемпотентной генерации логов по unique (medication_id, scheduled_at).
type LogKey struct {
	MedicationID int64
	ScheduledAt  time.Time
}

// Key возвращает ключ уникальности доменной записи.
func (l *Log) Key() *LogKey {
	return &LogKey{
		MedicationID: l.MedicationID,
		ScheduledAt:  l.ScheduledAt.UTC(),
	}
}
