package medication

import (
	"fmt"
	"strings"
	"time"

	"tracker/internal/domain/shared"
)

// Medication представляет карточку БАД с расписанием и метаданными.
type Medication struct {
	ID        int64
	UserID    int64
	Name      string
	Category  Category
	Dose      float64
	Unit      string
	Color     Color
	Schedule  Schedule
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewMedication создает новую сущность БАД с проверкой инвариантов.
func NewMedication(
	userID int64,
	name string,
	category Category,
	dose float64,
	unit string,
	color Color,
	schedule Schedule,
	active bool,
	now time.Time,
) (*Medication, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: user id should be positive", shared.ErrInvalidInput)
	}
	n := strings.TrimSpace(name)
	if n == "" {
		return nil, fmt.Errorf("%w: medication name should not be empty", shared.ErrInvalidInput)
	}
	if dose < 0 {
		return nil, fmt.Errorf("%w: medication dose should be >= 0", shared.ErrInvalidInput)
	}
	if err := category.Validate(); err != nil {
		return nil, err
	}
	if err := color.Validate(); err != nil {
		return nil, err
	}
	// Переиспользуем конструктор для гарантии валидности расписания.
	if _, err := NewSchedule(schedule.ByDay, schedule.Times); err != nil {
		return nil, err
	}

	ts := now.UTC()
	return &Medication{
		UserID:    userID,
		Name:      n,
		Category:  category,
		Dose:      dose,
		Unit:      strings.TrimSpace(unit),
		Color:     color,
		Schedule:  schedule,
		Active:    active,
		CreatedAt: ts,
		UpdatedAt: ts,
	}, nil
}

// Update меняет редактируемые поля карточки БАД.
func (m *Medication) Update(
	name string,
	category Category,
	dose float64,
	unit string,
	color Color,
	schedule Schedule,
	active bool,
	now time.Time,
) (*Medication, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("%w: medication name should not be empty", shared.ErrInvalidInput)
	}
	if dose < 0 {
		return nil, fmt.Errorf("%w: medication dose should be >= 0", shared.ErrInvalidInput)
	}
	if err := category.Validate(); err != nil {
		return nil, err
	}
	if err := color.Validate(); err != nil {
		return nil, err
	}
	if _, err := NewSchedule(schedule.ByDay, schedule.Times); err != nil {
		return nil, err
	}

	m.Name = strings.TrimSpace(name)
	m.Category = category
	m.Dose = dose
	m.Unit = strings.TrimSpace(unit)
	m.Color = color
	m.Schedule = schedule
	m.Active = active
	m.UpdatedAt = now.UTC()
	return m, nil
}

// SupplementDayItem read-model для drill-down страницы дня.
type SupplementDayItem struct {
	MedicationID int64
	Name         string
	Category     Category
	ScheduledAt  time.Time
	Status       Status
}

// SupplementDayDetails read-model для ответа /supplements/day.
type SupplementDayDetails struct {
	Date     shared.LocalDate
	IsPast   bool
	IsFuture bool
	Items    []SupplementDayItem
	Summary  map[Category]DailyCategorySummary
}
