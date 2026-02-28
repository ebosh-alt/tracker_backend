package steps

import (
	"context"
	"errors"
	"testing"

	"tracker/internal/application"
	domainSteps "tracker/internal/domain/steps"
	"tracker/internal/domain/shared"
)

type addRepoStub struct {
	addOut   domainSteps.DailyEntry
	addErr   error
	addCalls int
	lastUser int64
	lastDate shared.LocalDate
	lastDelta int
	lastSource domainSteps.Source
	lastCtx  context.Context
}

func (s *addRepoStub) GetByDateRange(
	ctx context.Context,
	userID int64,
	from shared.LocalDate,
	to shared.LocalDate,
) ([]domainSteps.DailyEntry, error) {
	return nil, nil
}

func (s *addRepoStub) UpsertByDate(ctx context.Context, entry domainSteps.DailyEntry) (domainSteps.DailyEntry, error) {
	return domainSteps.DailyEntry{}, nil
}

func (s *addRepoStub) AddDelta(
	ctx context.Context,
	userID int64,
	date shared.LocalDate,
	delta int,
	source domainSteps.Source,
) (domainSteps.DailyEntry, error) {
	s.addCalls++
	s.lastCtx = ctx
	s.lastUser = userID
	s.lastDate = date
	s.lastDelta = delta
	s.lastSource = source
	if s.addErr != nil {
		return domainSteps.DailyEntry{}, s.addErr
	}
	if s.addOut.Date == "" {
		return domainSteps.DailyEntry{
			UserID: userID,
			Date:   date,
			Count:  delta,
			Source: source,
		}, nil
	}
	return s.addOut, nil
}

func (s *addRepoStub) DeleteByDate(ctx context.Context, userID int64, date shared.LocalDate) error {
	return nil
}

type addUOWStub struct {
	err        error
	calls      int
	executedFn bool
	fnCtx      context.Context
}

func (s *addUOWStub) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	s.calls++
	if s.err != nil {
		return s.err
	}
	txCtx := context.WithValue(ctx, addTxCtxKey{}, "tx")
	s.fnCtx = txCtx
	s.executedFn = true
	return fn(txCtx)
}

type addTxCtxKey struct{}

func TestAddUseCase_Success(t *testing.T) {
	repo := &addRepoStub{
		addOut: domainSteps.DailyEntry{
			UserID: 42,
			Date:   shared.MustLocalDate("2026-02-27"),
			Count:  1500,
			Source: domainSteps.SourceManual,
		},
	}
	uow := &addUOWStub{}
	uc := NewAddUseCase(repo, uow)

	out, err := uc.Execute(context.Background(), AddInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
		Delta:  500,
		Source: "manual",
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
	if repo.addCalls != 1 {
		t.Fatalf("expected AddDelta calls=1, got %d", repo.addCalls)
	}
	if repo.lastUser != 42 {
		t.Fatalf("expected userID=42, got %d", repo.lastUser)
	}
	if repo.lastDate != shared.LocalDate("2026-02-27") {
		t.Fatalf("expected date 2026-02-27, got %q", repo.lastDate)
	}
	if repo.lastDelta != 500 {
		t.Fatalf("expected delta=500, got %d", repo.lastDelta)
	}
	if repo.lastSource != domainSteps.SourceManual {
		t.Fatalf("expected source=%q, got %q", domainSteps.SourceManual, repo.lastSource)
	}
	if got, ok := repo.lastCtx.Value(addTxCtxKey{}).(string); !ok || got != "tx" {
		t.Fatalf("expected tx context marker, got %#v", repo.lastCtx.Value(addTxCtxKey{}))
	}
	if out.Entry.Count != 1500 {
		t.Fatalf("expected output count=1500, got %d", out.Entry.Count)
	}
}

func TestAddUseCase_InvalidUserID(t *testing.T) {
	repo := &addRepoStub{}
	uow := &addUOWStub{}
	uc := NewAddUseCase(repo, uow)

	_, err := uc.Execute(context.Background(), AddInput{
		UserID: 0,
		Date:   shared.MustLocalDate("2026-02-27"),
		Delta:  500,
		Source: "manual",
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if uow.calls != 0 {
		t.Fatalf("uow should not be called, got %d", uow.calls)
	}
}

func TestAddUseCase_ZeroDelta(t *testing.T) {
	repo := &addRepoStub{}
	uow := &addUOWStub{}
	uc := NewAddUseCase(repo, uow)

	_, err := uc.Execute(context.Background(), AddInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
		Delta:  0,
		Source: "manual",
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if uow.calls != 0 {
		t.Fatalf("uow should not be called, got %d", uow.calls)
	}
}

func TestAddUseCase_InvalidSource(t *testing.T) {
	repo := &addRepoStub{}
	uow := &addUOWStub{}
	uc := NewAddUseCase(repo, uow)

	_, err := uc.Execute(context.Background(), AddInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
		Delta:  500,
		Source: "invalid",
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if uow.calls != 0 {
		t.Fatalf("uow should not be called, got %d", uow.calls)
	}
}

func TestAddUseCase_UOWError(t *testing.T) {
	repo := &addRepoStub{}
	uowErr := errors.New("tx begin failed")
	uow := &addUOWStub{err: uowErr}
	uc := NewAddUseCase(repo, uow)

	_, err := uc.Execute(context.Background(), AddInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
		Delta:  500,
		Source: "manual",
	})
	if !errors.Is(err, uowErr) {
		t.Fatalf("expected uow error, got %v", err)
	}
	if repo.addCalls != 0 {
		t.Fatalf("repo should not be called, got %d", repo.addCalls)
	}
}

func TestAddUseCase_RepoError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &addRepoStub{addErr: repoErr}
	uow := &addUOWStub{}
	uc := NewAddUseCase(repo, uow)

	_, err := uc.Execute(context.Background(), AddInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
		Delta:  -300,
		Source: "manual",
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if uow.calls != 1 {
		t.Fatalf("expected uow calls=1, got %d", uow.calls)
	}
	if repo.addCalls != 1 {
		t.Fatalf("expected AddDelta calls=1, got %d", repo.addCalls)
	}
}

func TestAddUseCase_NegativeTotalRejected(t *testing.T) {
	// Domain contract for BE-012: negative resulting count must be rejected.
	// Repository is expected to enforce this atomically and return ErrInvalidInput.
	repo := &addRepoStub{addErr: shared.ErrInvalidInput}
	uow := &addUOWStub{}
	uc := NewAddUseCase(repo, uow)

	_, err := uc.Execute(context.Background(), AddInput{
		UserID: 42,
		Date:   shared.MustLocalDate("2026-02-27"),
		Delta:  -5000,
		Source: "manual",
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

var _ application.UnitOfWork = (*addUOWStub)(nil)
