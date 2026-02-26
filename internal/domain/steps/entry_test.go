package steps

import (
	"errors"
	"testing"
	"time"

	"tracker/internal/domain/shared"
)

func TestParseSource_Manual(t *testing.T) {
	got, err := ParseSource(" manual ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != SourceManual {
		t.Fatalf("expected %q, got %q", SourceManual, got)
	}
}

func TestParseSource_Invalid(t *testing.T) {
	_, err := ParseSource("unknown")
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestNewDailyEntry_Success(t *testing.T) {
	now := time.Date(2026, 2, 26, 10, 0, 0, 0, time.UTC)
	date := shared.MustLocalDate("2026-02-26")

	entry, err := NewDailyEntry(42, date, 1234, SourceManual, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry")
	}
	if entry.UserID != 42 {
		t.Fatalf("expected userID=42, got %d", entry.UserID)
	}
	if entry.Date != date {
		t.Fatalf("expected date=%q, got %q", date, entry.Date)
	}
	if entry.Count != 1234 {
		t.Fatalf("expected count=1234, got %d", entry.Count)
	}
	if entry.Source != SourceManual {
		t.Fatalf("expected source=%q, got %q", SourceManual, entry.Source)
	}
	if !entry.CreatedAt.Equal(now.UTC()) {
		t.Fatalf("expected createdAt=%s, got %s", now.UTC(), entry.CreatedAt)
	}
	if !entry.UpdatedAt.Equal(now.UTC()) {
		t.Fatalf("expected updatedAt=%s, got %s", now.UTC(), entry.UpdatedAt)
	}
}

func TestNewDailyEntry_InvalidUserID(t *testing.T) {
	now := time.Date(2026, 2, 26, 10, 0, 0, 0, time.UTC)
	date := shared.MustLocalDate("2026-02-26")

	_, err := NewDailyEntry(0, date, 1000, SourceManual, now)
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestNewDailyEntry_InvalidCount(t *testing.T) {
	now := time.Date(2026, 2, 26, 10, 0, 0, 0, time.UTC)
	date := shared.MustLocalDate("2026-02-26")

	_, err := NewDailyEntry(42, date, -1, SourceManual, now)
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestNewDailyEntry_InvalidSource(t *testing.T) {
	now := time.Date(2026, 2, 26, 10, 0, 0, 0, time.UTC)
	date := shared.MustLocalDate("2026-02-26")

	_, err := NewDailyEntry(42, date, 1000, Source("bad"), now)
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestNewDailyEntry_InvalidDate(t *testing.T) {
	now := time.Date(2026, 2, 26, 10, 0, 0, 0, time.UTC)

	_, err := NewDailyEntry(42, shared.LocalDate(""), 1000, SourceManual, now)
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
