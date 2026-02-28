package steps

import (
	"fmt"
	"strings"
	"time"
	"tracker/internal/domain/shared"
)

const (
	SourceManual Source = "manual"
)

type Source string

func ParseSource(raw string) (Source, error) {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch Source(v) {
	case SourceManual:
		return SourceManual, nil
	default:
		return "", fmt.Errorf("%w: invalid steps source %q", shared.ErrInvalidInput, raw)
	}
}

type DailyEntry struct {
	UserID    int64
	Count     int
	Source    Source
	Date      shared.LocalDate
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewDailyEntry(userID int64, date shared.LocalDate, count int, source Source, now time.Time) (*DailyEntry, error) {
	if userID == 0 {
		return nil, shared.ErrInvalidInput
	}
	if count < 0 {
		return nil, shared.ErrInvalidInput
	}

	if err := date.Validate(); err != nil {
		return nil, shared.ErrInvalidInput
	}
	if _, err := ParseSource(string(source)); err != nil {
		return nil, shared.ErrInvalidInput
	}
	return &DailyEntry{
		UserID:    userID,
		Count:     count,
		Source:    source,
		Date:      date,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC(),
	}, nil
}
