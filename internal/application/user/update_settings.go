package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"tracker/internal/application"
	"tracker/internal/domain/shared"
	domainUser "tracker/internal/domain/user"
)

type SettingsRepository interface {
	UpdateSettings(ctx context.Context, userID int64, patch domainUser.SettingsPatch) (domainUser.User, error)
}

type UpdateSettingsInput struct {
	UserID    int64
	Timezone  *string
	StepsGoal *int
}

type UpdateSettingsOutput struct {
	User domainUser.User
}

type UpdateSettingsUseCase struct {
	repo SettingsRepository
	uow  application.UnitOfWork
}

func NewUpdateSettingsUseCase(repo SettingsRepository, uow application.UnitOfWork) *UpdateSettingsUseCase {
	if repo == nil {
		panic(fmt.Errorf("%w: user settings repository dependency is required", shared.ErrInvalidInput))
	}
	if uow == nil {
		panic(fmt.Errorf("%w: nil unit of work dependency is required", shared.ErrInvalidInput))
	}

	return &UpdateSettingsUseCase{
		repo: repo,
		uow:  uow,
	}
}

func (uc *UpdateSettingsUseCase) Execute(ctx context.Context, in UpdateSettingsInput) (*UpdateSettingsOutput, error) {
	if in.UserID <= 0 {
		return nil, shared.ErrInvalidInput
	}
	if in.Timezone == nil && in.StepsGoal == nil {
		return nil, shared.ErrInvalidInput
	}

	var timezone *string
	if in.Timezone != nil {
		v := strings.TrimSpace(*in.Timezone)
		if v == "" {
			return nil, shared.ErrInvalidInput
		}
		if _, err := time.LoadLocation(v); err != nil {
			return nil, shared.ErrInvalidInput
		}
		timezone = &v
	}

	if in.StepsGoal != nil {
		if _, err := domainUser.NewStepsGoal(*in.StepsGoal); err != nil {
			return nil, err
		}
	}

	patch := domainUser.SettingsPatch{
		Timezone:  timezone,
		StepsGoal: in.StepsGoal,
	}

	var updated domainUser.User
	err := uc.uow.WithinTx(ctx, func(txCtx context.Context) error {
		user, err := uc.repo.UpdateSettings(txCtx, in.UserID, patch)
		if err != nil {
			return err
		}
		updated = user
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &UpdateSettingsOutput{User: updated}, nil
}
