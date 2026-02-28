package steps

import (
	"context"
	"errors"
	"testing"

	"tracker/internal/application"
	domainSteps "tracker/internal/domain/steps"
	"tracker/internal/domain/shared"
)

type putRepoStub struct {
	upsertOut   domainSteps.DailyEntry
	upsertErr   error
	upsertCalls int
	lastEntry   domainSteps.DailyEntry
}

func (s *putRepoStub) GetByDateRange(
	ctx context.Context,
	userID int64,
	from shared.LocalDate,
	to shared.LocalDate,
) ([]domainSteps.DailyEntry, error) {
	return nil, nil
}

func (s *putRepoStub) UpsertByDate(ctx context.Context, entry domainSteps.DailyEntry) (domainSteps.DailyEntry, error) {
	s.upsertCalls++
	s.lastEntry = entry
	if s.upsertErr != nil {
		return domainSteps.DailyEntry{}, s.upsertErr
	}
	if s.upsertOut.Date == "" {
		return entry, nil
	}
	return s.upsertOut, nil
}

func (s *putRepoStub) AddDelta(
	ctx context.Context,
	userID int64,
	date shared.LocalDate,
	delta int,
	source domainSteps.Source,
) (domainSteps.DailyEntry, error) {
	return domainSteps.DailyEntry{}, nil
}

func (s *putRepoStub) DeleteByDate(ctx context.Context, userID int64, date shared.LocalDate) error {
	return nil
}

type uowStub struct {
	err           error
	calls         int
	lastCtx       context.Context
	fnCtx         context.Context
	executedFn    bool
	executedFnErr error
}

func (s *uowStub) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	s.calls++
	s.lastCtx = ctx
	if s.err != nil {
		return s.err
	}
	txCtx := context.WithValue(ctx, txCtxMarkerKey{}, "tx")
	s.fnCtx = txCtx
	err := fn(txCtx)
	s.executedFn = true
	s.executedFnErr = err
	return err
}

type txCtxMarkerKey struct{}

func TestPutUseCase_Success(t *testing.T) {
	repo := &putRepoStub{}
	uow := &uowStub{}
	uc := NewPutUseCase(repo, uow)

	out, err := uc.Execute(context.Background(), PutInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
		Count:  1500,
		Source: "manual",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Entry.UserID != 42 {
		t.Fatalf("expected userID=42, got %d", out.Entry.UserID)
	}
	if out.Entry.Count != 1500 {
		t.Fatalf("expected count=1500, got %d", out.Entry.Count)
	}
	if out.Entry.Source != domainSteps.SourceManual {
		t.Fatalf("expected source=%q, got %q", domainSteps.SourceManual, out.Entry.Source)
	}
	if uow.calls != 1 {
		t.Fatalf("expected WithinTx calls=1, got %d", uow.calls)
	}
	if !uow.executedFn {
		t.Fatal("expected tx function to be executed")
	}
	if repo.upsertCalls != 1 {
		t.Fatalf("expected upsert calls=1, got %d", repo.upsertCalls)
	}
	if got, ok := uow.fnCtx.Value(txCtxMarkerKey{}).(string); !ok || got != "tx" {
		t.Fatalf("expected tx context marker, got %#v", uow.fnCtx.Value(txCtxMarkerKey{}))
	}
}

func TestPutUseCase_InvalidUserID(t *testing.T) {
	repo := &putRepoStub{}
	uow := &uowStub{}
	uc := NewPutUseCase(repo, uow)

	_, err := uc.Execute(context.Background(), PutInput{
		UserID: 0,
		Date:   shared.MustLocalDate("2026-02-27"),
		Count:  100,
		Source: "manual",
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if uow.calls != 0 {
		t.Fatalf("uow should not be called, got %d", uow.calls)
	}
}

func TestPutUseCase_InvalidCount(t *testing.T) {
	repo := &putRepoStub{}
	uow := &uowStub{}
	uc := NewPutUseCase(repo, uow)

	_, err := uc.Execute(context.Background(), PutInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
		Count:  -1,
		Source: "manual",
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if uow.calls != 0 {
		t.Fatalf("uow should not be called, got %d", uow.calls)
	}
}

func TestPutUseCase_InvalidSource(t *testing.T) {
	repo := &putRepoStub{}
	uow := &uowStub{}
	uc := NewPutUseCase(repo, uow)

	_, err := uc.Execute(context.Background(), PutInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
		Count:  100,
		Source: "invalid",
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if uow.calls != 0 {
		t.Fatalf("uow should not be called, got %d", uow.calls)
	}
}

func TestPutUseCase_UowError(t *testing.T) {
	repo := &putRepoStub{}
	uowErr := errors.New("tx begin failed")
	uow := &uowStub{err: uowErr}
	uc := NewPutUseCase(repo, uow)

	_, err := uc.Execute(context.Background(), PutInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
		Count:  100,
		Source: "manual",
	})
	if !errors.Is(err, uowErr) {
		t.Fatalf("expected uow error, got %v", err)
	}
	if repo.upsertCalls != 0 {
		t.Fatalf("repo should not be called, got %d", repo.upsertCalls)
	}
}

func TestPutUseCase_RepoError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &putRepoStub{upsertErr: repoErr}
	uow := &uowStub{}
	uc := NewPutUseCase(repo, uow)

	_, err := uc.Execute(context.Background(), PutInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
		Count:  100,
		Source: "manual",
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if uow.calls != 1 {
		t.Fatalf("expected uow calls=1, got %d", uow.calls)
	}
	if repo.upsertCalls != 1 {
		t.Fatalf("expected repo upsert calls=1, got %d", repo.upsertCalls)
	}
}

var _ application.UnitOfWork = (*uowStub)(nil)
