package steps

import (
	"context"
	"errors"
	"testing"

	"tracker/internal/domain/shared"
	domainSteps "tracker/internal/domain/steps"
)

type listRepoStub struct {
	out      []domainSteps.DailyEntry
	err      error
	calls    int
	lastUser int64
	lastFrom shared.LocalDate
	lastTo   shared.LocalDate
}

func (s *listRepoStub) GetByDateRange(
	ctx context.Context,
	userID int64,
	from shared.LocalDate,
	to shared.LocalDate,
) ([]domainSteps.DailyEntry, error) {
	s.calls++
	s.lastUser = userID
	s.lastFrom = from
	s.lastTo = to
	if s.err != nil {
		return nil, s.err
	}
	return s.out, nil
}

func (s *listRepoStub) UpsertByDate(ctx context.Context, entry domainSteps.DailyEntry) (domainSteps.DailyEntry, error) {
	return domainSteps.DailyEntry{}, nil
}

func (s *listRepoStub) AddDelta(
	ctx context.Context,
	userID int64,
	date shared.LocalDate,
	delta int,
	source domainSteps.Source,
) (domainSteps.DailyEntry, error) {
	return domainSteps.DailyEntry{}, nil
}

func (s *listRepoStub) DeleteByDate(ctx context.Context, userID int64, date shared.LocalDate) error {
	return nil
}

func TestListUseCase_Success(t *testing.T) {
	repo := &listRepoStub{
		out: []domainSteps.DailyEntry{
			{
				UserID: 42,
				Date:   shared.MustLocalDate("2026-02-24"),
				Count:  7200,
				Source: domainSteps.SourceManual,
			},
			{
				UserID: 42,
				Date:   shared.MustLocalDate("2026-02-25"),
				Count:  9400,
				Source: domainSteps.SourceManual,
			},
		},
	}
	uc := NewListUseCase(repo)

	out, err := uc.Execute(context.Background(), ListInput{
		UserID: 42,
		From:   shared.MustLocalDate("2026-02-24"),
		To:     shared.MustLocalDate("2026-02-25"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.calls != 1 {
		t.Fatalf("expected 1 repo call, got %d", repo.calls)
	}
	if repo.lastUser != 42 {
		t.Fatalf("expected userID=42, got %d", repo.lastUser)
	}
	if repo.lastFrom != shared.LocalDate("2026-02-24") || repo.lastTo != shared.LocalDate("2026-02-25") {
		t.Fatalf("unexpected range: from=%q to=%q", repo.lastFrom, repo.lastTo)
	}
	if len(out.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out.Entries))
	}
}

func TestListUseCase_InvalidUserID(t *testing.T) {
	repo := &listRepoStub{}
	uc := NewListUseCase(repo)

	_, err := uc.Execute(context.Background(), ListInput{
		UserID: 0,
		From:   shared.MustLocalDate("2026-02-24"),
		To:     shared.MustLocalDate("2026-02-25"),
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if repo.calls != 0 {
		t.Fatalf("repo should not be called, got %d", repo.calls)
	}
}

func TestListUseCase_InvalidRange(t *testing.T) {
	repo := &listRepoStub{}
	uc := NewListUseCase(repo)

	_, err := uc.Execute(context.Background(), ListInput{
		UserID: 42,
		From:   shared.MustLocalDate("2026-02-26"),
		To:     shared.MustLocalDate("2026-02-25"),
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if repo.calls != 0 {
		t.Fatalf("repo should not be called, got %d", repo.calls)
	}
}

func TestListUseCase_RepoError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &listRepoStub{err: repoErr}
	uc := NewListUseCase(repo)

	_, err := uc.Execute(context.Background(), ListInput{
		UserID: 42,
		From:   shared.MustLocalDate("2026-02-24"),
		To:     shared.MustLocalDate("2026-02-25"),
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
