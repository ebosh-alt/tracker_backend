package steps

import (
	"errors"
	"testing"

	"tracker/internal/domain/shared"
)

func TestPointValidate_Success(t *testing.T) {
	p := Point{
		Date:  shared.MustLocalDate("2026-02-26"),
		Steps: 4200,
	}

	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPointValidate_NegativeSteps(t *testing.T) {
	p := Point{
		Date:  shared.MustLocalDate("2026-02-26"),
		Steps: -1,
	}

	err := p.Validate()
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestPointValidate_InvalidDate(t *testing.T) {
	p := Point{
		Date:  shared.LocalDate(""),
		Steps: 1000,
	}

	err := p.Validate()
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAnalyticsValidate_Success(t *testing.T) {
	a := Analytics{
		GoalTotal:         70000,
		FactTotal:         54230,
		CompletionPercent: 77.47,
		Series: []Point{
			{Date: shared.MustLocalDate("2026-02-24"), Steps: 7200},
			{Date: shared.MustLocalDate("2026-02-25"), Steps: 9400},
		},
	}

	if err := a.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyticsValidate_NegativeGoalTotal(t *testing.T) {
	a := Analytics{
		GoalTotal:         -1,
		FactTotal:         1,
		CompletionPercent: 100,
		Series:            nil,
	}

	err := a.Validate()
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAnalyticsValidate_NegativeFactTotal(t *testing.T) {
	a := Analytics{
		GoalTotal:         10,
		FactTotal:         -1,
		CompletionPercent: 0,
		Series:            nil,
	}

	err := a.Validate()
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAnalyticsValidate_NegativeCompletionPercent(t *testing.T) {
	a := Analytics{
		GoalTotal:         10,
		FactTotal:         10,
		CompletionPercent: -0.1,
		Series:            nil,
	}

	err := a.Validate()
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAnalyticsValidate_InvalidSeriesPoint(t *testing.T) {
	a := Analytics{
		GoalTotal:         10,
		FactTotal:         10,
		CompletionPercent: 100,
		Series: []Point{
			{Date: shared.MustLocalDate("2026-02-26"), Steps: 2000},
			{Date: shared.LocalDate(""), Steps: 10},
		},
	}

	err := a.Validate()
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
