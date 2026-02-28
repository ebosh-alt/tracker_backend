package postgres

import (
	"context"
	"errors"
	"testing"

	domainSteps "tracker/internal/domain/steps"
	"tracker/internal/domain/shared"
)

func TestNewStepsRepository_ImplementsDomainRepository(t *testing.T) {
	repo := NewStepsRepository(nil)
	if repo == nil {
		t.Fatal("expected repository instance")
	}
	var _ domainSteps.Repository = repo
}

func TestStepsRepository_GetByDateRange_InvalidUserID(t *testing.T) {
	repo := NewStepsRepository(nil)

	_, err := repo.GetByDateRange(
		context.Background(),
		0,
		shared.MustLocalDate("2026-02-01"),
		shared.MustLocalDate("2026-02-28"),
	)
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestStepsRepository_GetByDateRange_InvalidFrom(t *testing.T) {
	repo := NewStepsRepository(nil)

	_, err := repo.GetByDateRange(
		context.Background(),
		42,
		shared.LocalDate(""),
		shared.MustLocalDate("2026-02-28"),
	)
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestStepsRepository_GetByDateRange_InvalidTo(t *testing.T) {
	repo := NewStepsRepository(nil)

	_, err := repo.GetByDateRange(
		context.Background(),
		42,
		shared.MustLocalDate("2026-02-01"),
		shared.LocalDate(""),
	)
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestStepsRepository_GetByDateRange_InvalidRange(t *testing.T) {
	repo := NewStepsRepository(nil)

	_, err := repo.GetByDateRange(
		context.Background(),
		42,
		shared.MustLocalDate("2026-02-28"),
		shared.MustLocalDate("2026-02-01"),
	)
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestStepsRepository_UpsertByDate_InvalidUserID(t *testing.T) {
	repo := NewStepsRepository(nil)

	_, err := repo.UpsertByDate(context.Background(), domainSteps.DailyEntry{
		UserID: 0,
		Date:   shared.MustLocalDate("2026-02-27"),
		Count:  1000,
		Source: domainSteps.SourceManual,
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestStepsRepository_UpsertByDate_InvalidDate(t *testing.T) {
	repo := NewStepsRepository(nil)

	_, err := repo.UpsertByDate(context.Background(), domainSteps.DailyEntry{
		UserID: 42,
		Date:   shared.LocalDate(""),
		Count:  1000,
		Source: domainSteps.SourceManual,
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestStepsRepository_UpsertByDate_InvalidCount(t *testing.T) {
	repo := NewStepsRepository(nil)

	_, err := repo.UpsertByDate(context.Background(), domainSteps.DailyEntry{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
		Count:  -1,
		Source: domainSteps.SourceManual,
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestStepsRepository_UpsertByDate_InvalidSource(t *testing.T) {
	repo := NewStepsRepository(nil)

	_, err := repo.UpsertByDate(context.Background(), domainSteps.DailyEntry{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
		Count:  1000,
		Source: domainSteps.Source("bad"),
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestStepsRepository_AddDelta_InvalidUserID(t *testing.T) {
	repo := NewStepsRepository(nil)

	_, err := repo.AddDelta(
		context.Background(),
		0,
		shared.MustLocalDate("2026-02-27"),
		500,
		domainSteps.SourceManual,
	)
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestStepsRepository_AddDelta_InvalidDate(t *testing.T) {
	repo := NewStepsRepository(nil)

	_, err := repo.AddDelta(
		context.Background(),
		42,
		shared.LocalDate(""),
		500,
		domainSteps.SourceManual,
	)
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestStepsRepository_AddDelta_InvalidSource(t *testing.T) {
	repo := NewStepsRepository(nil)

	_, err := repo.AddDelta(
		context.Background(),
		42,
		shared.MustLocalDate("2026-02-27"),
		500,
		domainSteps.Source("bad"),
	)
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestStepsRepository_DeleteByDate_InvalidUserID(t *testing.T) {
	repo := NewStepsRepository(nil)

	err := repo.DeleteByDate(
		context.Background(),
		0,
		shared.MustLocalDate("2026-02-27"),
	)
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestStepsRepository_DeleteByDate_InvalidDate(t *testing.T) {
	repo := NewStepsRepository(nil)

	err := repo.DeleteByDate(
		context.Background(),
		42,
		shared.LocalDate(""),
	)
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

