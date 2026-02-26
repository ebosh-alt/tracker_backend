package medication

import (
	"fmt"
	"strings"
	"time"

	"tracker/internal/domain/shared"
)

// Category определяет продуктовую категорию БАД.
type Category string

const (
	CategoryTablet         Category = "tablet"
	CategorySportNutrition Category = "sport_nutrition"
)

// Validate проверяет корректность категории.
func (c Category) Validate() error {
	switch c {
	case CategoryTablet, CategorySportNutrition:
		return nil
	default:
		return fmt.Errorf("%w: invalid medication category %q", shared.ErrInvalidInput, c)
	}
}

// Color определяет цветовую маркировку БАД в UI.
type Color string

const (
	ColorLime   Color = "lime"
	ColorViolet Color = "violet"
	ColorRose   Color = "rose"
	ColorInfo   Color = "info"
)

// Validate проверяет корректность цветового enum.
func (c Color) Validate() error {
	switch c {
	case ColorLime, ColorViolet, ColorRose, ColorInfo:
		return nil
	default:
		return fmt.Errorf("%w: invalid medication color %q", shared.ErrInvalidInput, c)
	}
}

// Status отражает состояние конкретного запланированного приема.
type Status string

const (
	StatusPending Status = "pending"
	StatusTaken   Status = "taken"
	StatusMissed  Status = "missed"
)

// NormalizeStatus переводит входящее значение в канонический статус.
// Поддерживается alias skipped -> missed для обратной совместимости.
func NormalizeStatus(raw string) (Status, error) {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case string(StatusPending):
		return StatusPending, nil
	case string(StatusTaken):
		return StatusTaken, nil
	case string(StatusMissed), "skipped":
		return StatusMissed, nil
	default:
		return "", fmt.Errorf("%w: invalid medication log status %q", shared.ErrInvalidInput, raw)
	}
}

// CanTransitionTo проверяет допустимость перехода между статусами.
func (s Status) CanTransitionTo(next Status, allowCorrections bool) bool {
	if s == next {
		return true
	}

	switch s {
	case StatusPending:
		return next == StatusTaken || next == StatusMissed
	case StatusTaken:
		return allowCorrections && next == StatusMissed
	case StatusMissed:
		return allowCorrections && next == StatusTaken
	default:
		return false
	}
}

// Schedule описывает правило генерации приемов: дни недели и времена суток.
type Schedule struct {
	ByDay []shared.Weekday
	Times []shared.LocalClockTime
}

// NewSchedule создает валидное расписание БАД.
func NewSchedule(byDay []shared.Weekday, times []shared.LocalClockTime) (Schedule, error) {
	validDays, err := shared.ValidateWeekdays(byDay)
	if err != nil {
		return Schedule{}, err
	}
	validTimes, err := shared.ValidateClockTimes(times)
	if err != nil {
		return Schedule{}, err
	}
	return Schedule{
		ByDay: validDays,
		Times: validTimes,
	}, nil
}

// IsScheduledOnDate проверяет, должна ли быть доза в указанную дату.
func (s Schedule) IsScheduledOnDate(date shared.LocalDate, loc *time.Location) (bool, error) {
	t, err := date.TimeIn(loc)
	if err != nil {
		return false, err
	}
	day := shared.WeekdayFromTime(t)
	for _, d := range s.ByDay {
		if d == day {
			return true, nil
		}
	}
	return false, nil
}

// DailyCategorySummary хранит счетчики статусов для категории за день.
type DailyCategorySummary struct {
	Taken   int
	Missed  int
	Pending int
}
