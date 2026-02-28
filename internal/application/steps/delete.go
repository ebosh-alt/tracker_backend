package steps

import (
	"context"
	"errors"
	"fmt"
	"tracker/internal/application"
	"tracker/internal/domain/shared"
	domainSteps "tracker/internal/domain/steps"
)

type DeleteInput struct {
	UserID int64
	Date   shared.LocalDate
}

type DeleteUseCase struct {
	repo domainSteps.Repository
	uow  application.UnitOfWork
}

func NewDeleteUseCase(repo domainSteps.Repository, uow application.UnitOfWork) *DeleteUseCase {
	if repo == nil {
		panic(fmt.Errorf("%w: steps repository dependency is required", shared.ErrInvalidInput))
	}

	if uow == nil {
		panic(fmt.Errorf("%w: nil unit of work dependency is required", shared.ErrInvalidInput))

	}
	return &DeleteUseCase{
		repo: repo,
		uow:  uow,
	}

}

func (uc *DeleteUseCase) Execute(ctx context.Context, in DeleteInput) error {
	if in.UserID <= 0 {
		return shared.ErrInvalidInput
	}

	err := uc.uow.WithinTx(ctx, func(txCtx context.Context) error {
		err := uc.repo.DeleteByDate(txCtx, in.UserID, in.Date)
		return err
	})
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}
