package shared

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	dateLayout      = "2006-01-02"
	monthLayout     = "2006-01"
	clockTimeLayout = "15:04"
)

// LocalDate хранит дату без времени в формате YYYY-MM-DD.
type LocalDate string

// ParseLocalDate валидирует и создает LocalDate.
func ParseLocalDate(raw string) (LocalDate, error) {
	if _, err := time.Parse(dateLayout, raw); err != nil {
		return "", fmt.Errorf("%w: invalid date format %q", ErrInvalidInput, raw)
	}
	return LocalDate(raw), nil
}

// MustLocalDate используется в тестах/fixtures, когда ошибка недопустима.
func MustLocalDate(raw string) LocalDate {
	d, err := ParseLocalDate(raw)
	if err != nil {
		panic(err)
	}
	return d
}

// String возвращает строковое представление даты.
func (d LocalDate) String() string {
	return string(d)
}

// Validate проверяет корректность формата даты.
func (d LocalDate) Validate() error {
	_, err := ParseLocalDate(d.String())
	return err
}

// TimeIn возвращает начало суток этой даты в заданной таймзоне.
func (d LocalDate) TimeIn(loc *time.Location) (time.Time, error) {
	if loc == nil {
		return time.Time{}, fmt.Errorf("%w: location is nil", ErrInvalidInput)
	}
	t, err := time.ParseInLocation(dateLayout, d.String(), loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: invalid date %q", ErrInvalidInput, d)
	}
	return t, nil
}

// Weekday это день недели в формате MO..SU.
type Weekday string

const (
	WeekdayMonday    Weekday = "MO"
	WeekdayTuesday   Weekday = "TU"
	WeekdayWednesday Weekday = "WE"
	WeekdayThursday  Weekday = "TH"
	WeekdayFriday    Weekday = "FR"
	WeekdaySaturday  Weekday = "SA"
	WeekdaySunday    Weekday = "SU"
)

var validWeekdays = map[Weekday]struct{}{
	WeekdayMonday:    {},
	WeekdayTuesday:   {},
	WeekdayWednesday: {},
	WeekdayThursday:  {},
	WeekdayFriday:    {},
	WeekdaySaturday:  {},
	WeekdaySunday:    {},
}

// ParseWeekday валидирует строку и возвращает доменный день недели.
func ParseWeekday(raw string) (Weekday, error) {
	wd := Weekday(strings.ToUpper(strings.TrimSpace(raw)))
	if _, ok := validWeekdays[wd]; !ok {
		return "", fmt.Errorf("%w: invalid weekday %q", ErrInvalidInput, raw)
	}
	return wd, nil
}

// ValidateWeekdays проверяет список дней недели, убирает дубликаты и сортирует по календарному порядку.
func ValidateWeekdays(days []Weekday) ([]Weekday, error) {
	if len(days) == 0 {
		return nil, fmt.Errorf("%w: weekdays should not be empty", ErrInvalidInput)
	}

	seen := make(map[Weekday]struct{}, len(days))
	out := make([]Weekday, 0, len(days))

	for _, d := range days {
		if _, ok := validWeekdays[d]; !ok {
			return nil, fmt.Errorf("%w: invalid weekday %q", ErrInvalidInput, d)
		}
		if _, dup := seen[d]; dup {
			continue
		}
		seen[d] = struct{}{}
		out = append(out, d)
	}

	sort.Slice(out, func(i, j int) bool {
		return weekdayOrder(out[i]) < weekdayOrder(out[j])
	})

	return out, nil
}

func weekdayOrder(day Weekday) int {
	switch day {
	case WeekdayMonday:
		return 1
	case WeekdayTuesday:
		return 2
	case WeekdayWednesday:
		return 3
	case WeekdayThursday:
		return 4
	case WeekdayFriday:
		return 5
	case WeekdaySaturday:
		return 6
	case WeekdaySunday:
		return 7
	default:
		return 99
	}
}

// WeekdayFromTime конвертирует time.Weekday в доменный Weekday.
func WeekdayFromTime(t time.Time) Weekday {
	switch t.Weekday() {
	case time.Monday:
		return WeekdayMonday
	case time.Tuesday:
		return WeekdayTuesday
	case time.Wednesday:
		return WeekdayWednesday
	case time.Thursday:
		return WeekdayThursday
	case time.Friday:
		return WeekdayFriday
	case time.Saturday:
		return WeekdaySaturday
	default:
		return WeekdaySunday
	}
}

// LocalClockTime хранит локальное время без даты в формате HH:mm.
type LocalClockTime string

// ParseLocalClockTime валидирует время HH:mm.
func ParseLocalClockTime(raw string) (LocalClockTime, error) {
	v := strings.TrimSpace(raw)
	if _, err := time.Parse(clockTimeLayout, v); err != nil {
		return "", fmt.Errorf("%w: invalid time format %q", ErrInvalidInput, raw)
	}
	return LocalClockTime(v), nil
}

// String возвращает строковое представление времени.
func (t LocalClockTime) String() string {
	return string(t)
}

// ValidateClockTimes проверяет список времен, убирает дубликаты и сортирует по возрастанию.
func ValidateClockTimes(times []LocalClockTime) ([]LocalClockTime, error) {
	if len(times) == 0 {
		return nil, fmt.Errorf("%w: times should not be empty", ErrInvalidInput)
	}

	seen := make(map[LocalClockTime]struct{}, len(times))
	out := make([]LocalClockTime, 0, len(times))

	for _, t := range times {
		if _, err := ParseLocalClockTime(t.String()); err != nil {
			return nil, err
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}

	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

// LocalMonth хранит месяц без дня в формате YYYY-MM.
type LocalMonth string

// ParseLocalMonth валидирует и создает LocalMonth.
func ParseLocalMonth(raw string) (LocalMonth, error) {
	if _, err := time.Parse(monthLayout, raw); err != nil {
		return "", fmt.Errorf("%w: invalid month format %q", ErrInvalidInput, raw)
	}
	return LocalMonth(raw), nil
}

// String возвращает строковое представление месяца.
func (m LocalMonth) String() string {
	return string(m)
}

// FirstDay возвращает первый день месяца в локальной таймзоне.
func (m LocalMonth) FirstDay(loc *time.Location) (time.Time, error) {
	if loc == nil {
		return time.Time{}, fmt.Errorf("%w: location is nil", ErrInvalidInput)
	}
	t, err := time.ParseInLocation(monthLayout, m.String(), loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: invalid month %q", ErrInvalidInput, m)
	}
	return t, nil
}

// DaysInMonth возвращает количество дней в месяце.
func (m LocalMonth) DaysInMonth(loc *time.Location) (int, error) {
	first, err := m.FirstDay(loc)
	if err != nil {
		return 0, err
	}
	next := first.AddDate(0, 1, 0)
	return int(next.Sub(first).Hours() / 24), nil
}

func (m LocalMonth) Validate() error {
	_, err := ParseLocalMonth(m.String())
	return err
}
