package postgres

import (
	"context"
	"errors"
	"testing"

	"tracker/internal/domain/shared"
	domainUser "tracker/internal/domain/user"
)

func TestUserRepo_UpdateSettings_InvalidUserID(t *testing.T) {
	repo := NewUserReader(nil)

	goal := 12000
	_, err := repo.UpdateSettings(context.Background(), 0, domainUser.SettingsPatch{
		StepsGoal: &goal,
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUserRepo_UpdateSettings_EmptyPatch(t *testing.T) {
	repo := NewUserReader(nil)

	_, err := repo.UpdateSettings(context.Background(), 42, domainUser.SettingsPatch{})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUserRepo_UpdateSettings_InvalidTimezone(t *testing.T) {
	repo := NewUserReader(nil)

	tz := "Bad/Timezone"
	_, err := repo.UpdateSettings(context.Background(), 42, domainUser.SettingsPatch{
		Timezone: &tz,
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUserRepo_UpdateSettings_InvalidStepsGoal(t *testing.T) {
	repo := NewUserReader(nil)

	goal := 1234
	_, err := repo.UpdateSettings(context.Background(), 42, domainUser.SettingsPatch{
		StepsGoal: &goal,
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
