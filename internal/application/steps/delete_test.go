package steps

import (
	"context"
	"errors"
	"testing"

	"tracker/internal/application"
	domainSteps "tracker/internal/domain/steps"
	"tracker/internal/domain/shared"
)

type deleteRepoStub struct {
	deleteErr   error
	deleteCalls int
	lastCtx     context.Context
	lastUser    int64
	lastDate    shared.LocalDate
}

func (s *deleteRepoStub) GetByDateRange(
	ctx context.Context,
	userID int64,
	from shared.LocalDate,
	to shared.LocalDate,
) ([]domainSteps.DailyEntry, error) {
	return nil, nil
}

func (s *deleteRepoStub) UpsertByDate(ctx context.Context, entry domainSteps.DailyEntry) (domainSteps.DailyEntry, error) {
	return domainSteps.DailyEntry{}, nil
}

func (s *deleteRepoStub) AddDelta(
	ctx context.Context,
	userID int64,
	date shared.LocalDate,
	delta int,
	source domainSteps.Source,
) (domainSteps.DailyEntry, error) {
	return domainSteps.DailyEntry{}, nil
}

func (s *deleteRepoStub) DeleteByDate(ctx context.Context, userID int64, date shared.LocalDate) error {
	s.deleteCalls++
	s.lastCtx = ctx
	s.lastUser = userID
	s.lastDate = date
	return s.deleteErr
}

type deleteUOWStub struct {
	err        error
	calls      int
	executedFn bool
	fnCtx      context.Context
}

func (s *deleteUOWStub) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	s.calls++
	if s.err != nil {
		return s.err
	}
	txCtx := context.WithValue(ctx, deleteTxCtxKey{}, "tx")
	s.fnCtx = txCtx
	s.executedFn = true
	return fn(txCtx)
}

type deleteTxCtxKey struct{}

func TestDeleteUseCase_Success(t *testing.T) {
	repo := &deleteRepoStub{}
	uow := &deleteUOWStub{}
	uc := NewDeleteUseCase(repo, uow)

	err := uc.Execute(context.Background(), DeleteInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uow.calls != 1 {
		t.Fatalf("expected WithinTx calls=1, got %d", uow.calls)
	}
	if !uow.executedFn {
		t.Fatal("expected tx function to be executed")
	}
	if repo.deleteCalls != 1 {
		t.Fatalf("expected DeleteByDate calls=1, got %d", repo.deleteCalls)
	}
	if repo.lastUser != 42 {
		t.Fatalf("expected userID=42, got %d", repo.lastUser)
	}
	if repo.lastDate != shared.LocalDate("2026-02-27") {
		t.Fatalf("expected date 2026-02-27, got %q", repo.lastDate)
	}
	if got, ok := repo.lastCtx.Value(deleteTxCtxKey{}).(string); !ok || got != "tx" {
		t.Fatalf("expected tx context marker, got %#v", repo.lastCtx.Value(deleteTxCtxKey{}))
	}
}

func TestDeleteUseCase_InvalidUserID(t *testing.T) {
	repo := &deleteRepoStub{}
	uow := &deleteUOWStub{}
	uc := NewDeleteUseCase(repo, uow)

	err := uc.Execute(context.Background(), DeleteInput{
		UserID: 0,
		Date:   shared.MustLocalDate("2026-02-27"),
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if uow.calls != 0 {
		t.Fatalf("uow should not be called, got %d", uow.calls)
	}
}

func TestDeleteUseCase_UOWError(t *testing.T) {
	repo := &deleteRepoStub{}
	uowErr := errors.New("tx begin failed")
	uow := &deleteUOWStub{err: uowErr}
	uc := NewDeleteUseCase(repo, uow)

	err := uc.Execute(context.Background(), DeleteInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
	})
	if !errors.Is(err, uowErr) {
		t.Fatalf("expected uow error, got %v", err)
	}
	if repo.deleteCalls != 0 {
		t.Fatalf("repo should not be called, got %d", repo.deleteCalls)
	}
}

func TestDeleteUseCase_RepoError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &deleteRepoStub{deleteErr: repoErr}
	uow := &deleteUOWStub{}
	uc := NewDeleteUseCase(repo, uow)

	err := uc.Execute(context.Background(), DeleteInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if uow.calls != 1 {
		t.Fatalf("expected uow calls=1, got %d", uow.calls)
	}
	if repo.deleteCalls != 1 {
		t.Fatalf("expected DeleteByDate calls=1, got %d", repo.deleteCalls)
	}
}

func TestDeleteUseCase_NotFoundIsIdempotent(t *testing.T) {
	repo := &deleteRepoStub{deleteErr: shared.ErrNotFound}
	uow := &deleteUOWStub{}
	uc := NewDeleteUseCase(repo, uow)

	err := uc.Execute(context.Background(), DeleteInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
	})
	if err != nil {
		t.Fatalf("expected nil error for not found, got %v", err)
	}
	if uow.calls != 1 {
		t.Fatalf("expected uow calls=1, got %d", uow.calls)
	}
	if repo.deleteCalls != 1 {
		t.Fatalf("expected DeleteByDate calls=1, got %d", repo.deleteCalls)
	}
}

var _ application.UnitOfWork = (*deleteUOWStub)(nil)
