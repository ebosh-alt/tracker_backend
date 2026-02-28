package user

import (
	"context"
	"errors"
	"testing"

	"tracker/internal/application"
	"tracker/internal/domain/shared"
	domainUser "tracker/internal/domain/user"
)

type settingsRepoStub struct {
	updateOut   domainUser.User
	updateErr   error
	updateCalls int
	lastCtx     context.Context
	lastUserID  int64
	lastPatch   domainUser.SettingsPatch
}

func (s *settingsRepoStub) UpdateSettings(ctx context.Context, userID int64, patch domainUser.SettingsPatch) (domainUser.User, error) {
	s.updateCalls++
	s.lastCtx = ctx
	s.lastUserID = userID
	s.lastPatch = patch
	if s.updateErr != nil {
		return domainUser.User{}, s.updateErr
	}
	if s.updateOut.ID == 0 {
		goal, _ := domainUser.NewStepsGoal(10000)
		return domainUser.User{
			ID:        userID,
			Timezone:  "UTC",
			StepsGoal: goal,
		}, nil
	}
	return s.updateOut, nil
}

type settingsUOWStub struct {
	err        error
	calls      int
	executedFn bool
	fnCtx      context.Context
}

func (s *settingsUOWStub) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	s.calls++
	if s.err != nil {
		return s.err
	}
	txCtx := context.WithValue(ctx, settingsTxMarkerKey{}, "tx")
	s.fnCtx = txCtx
	s.executedFn = true
	return fn(txCtx)
}

type settingsTxMarkerKey struct{}

func TestUpdateSettingsUseCase_Success(t *testing.T) {
	repo := &settingsRepoStub{}
	uow := &settingsUOWStub{}
	uc := NewUpdateSettingsUseCase(repo, uow)

	tz := "Europe/Moscow"
	goal := 12000
	out, err := uc.Execute(context.Background(), UpdateSettingsInput{
		UserID:    42,
		Timezone:  &tz,
		StepsGoal: &goal,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.User.ID != 42 {
		t.Fatalf("expected user id=42, got %d", out.User.ID)
	}
	if repo.updateCalls != 1 {
		t.Fatalf("expected update calls=1, got %d", repo.updateCalls)
	}
	if repo.lastUserID != 42 {
		t.Fatalf("expected user id=42, got %d", repo.lastUserID)
	}
	if repo.lastPatch.Timezone == nil || *repo.lastPatch.Timezone != "Europe/Moscow" {
		t.Fatalf("unexpected timezone patch: %#v", repo.lastPatch.Timezone)
	}
	if repo.lastPatch.StepsGoal == nil || *repo.lastPatch.StepsGoal != 12000 {
		t.Fatalf("unexpected stepsGoal patch: %#v", repo.lastPatch.StepsGoal)
	}
	if uow.calls != 1 {
		t.Fatalf("expected uow calls=1, got %d", uow.calls)
	}
	if !uow.executedFn {
		t.Fatal("expected tx function to be executed")
	}
	if got, ok := repo.lastCtx.Value(settingsTxMarkerKey{}).(string); !ok || got != "tx" {
		t.Fatalf("expected tx marker, got %#v", repo.lastCtx.Value(settingsTxMarkerKey{}))
	}
}

func TestUpdateSettingsUseCase_InvalidUserID(t *testing.T) {
	repo := &settingsRepoStub{}
	uow := &settingsUOWStub{}
	uc := NewUpdateSettingsUseCase(repo, uow)

	tz := "UTC"
	_, err := uc.Execute(context.Background(), UpdateSettingsInput{
		UserID:   0,
		Timezone: &tz,
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if uow.calls != 0 {
		t.Fatalf("uow should not be called, got %d", uow.calls)
	}
}

func TestUpdateSettingsUseCase_EmptyPatch(t *testing.T) {
	repo := &settingsRepoStub{}
	uow := &settingsUOWStub{}
	uc := NewUpdateSettingsUseCase(repo, uow)

	_, err := uc.Execute(context.Background(), UpdateSettingsInput{
		UserID: 42,
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if uow.calls != 0 {
		t.Fatalf("uow should not be called, got %d", uow.calls)
	}
}

func TestUpdateSettingsUseCase_InvalidTimezone(t *testing.T) {
	repo := &settingsRepoStub{}
	uow := &settingsUOWStub{}
	uc := NewUpdateSettingsUseCase(repo, uow)

	tz := "Bad/Timezone"
	_, err := uc.Execute(context.Background(), UpdateSettingsInput{
		UserID:   42,
		Timezone: &tz,
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if uow.calls != 0 {
		t.Fatalf("uow should not be called, got %d", uow.calls)
	}
}

func TestUpdateSettingsUseCase_InvalidStepsGoal(t *testing.T) {
	repo := &settingsRepoStub{}
	uow := &settingsUOWStub{}
	uc := NewUpdateSettingsUseCase(repo, uow)

	goal := 1234
	_, err := uc.Execute(context.Background(), UpdateSettingsInput{
		UserID:    42,
		StepsGoal: &goal,
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if uow.calls != 0 {
		t.Fatalf("uow should not be called, got %d", uow.calls)
	}
}

func TestUpdateSettingsUseCase_UOWError(t *testing.T) {
	repo := &settingsRepoStub{}
	uowErr := errors.New("begin tx failed")
	uow := &settingsUOWStub{err: uowErr}
	uc := NewUpdateSettingsUseCase(repo, uow)

	tz := "UTC"
	_, err := uc.Execute(context.Background(), UpdateSettingsInput{
		UserID:   42,
		Timezone: &tz,
	})
	if !errors.Is(err, uowErr) {
		t.Fatalf("expected uow error, got %v", err)
	}
	if repo.updateCalls != 0 {
		t.Fatalf("repo should not be called, got %d", repo.updateCalls)
	}
}

func TestUpdateSettingsUseCase_RepoError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &settingsRepoStub{updateErr: repoErr}
	uow := &settingsUOWStub{}
	uc := NewUpdateSettingsUseCase(repo, uow)

	goal := 15000
	_, err := uc.Execute(context.Background(), UpdateSettingsInput{
		UserID:    42,
		StepsGoal: &goal,
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if uow.calls != 1 {
		t.Fatalf("expected uow calls=1, got %d", uow.calls)
	}
	if repo.updateCalls != 1 {
		t.Fatalf("expected repo calls=1, got %d", repo.updateCalls)
	}
}

var _ application.UnitOfWork = (*settingsUOWStub)(nil)
