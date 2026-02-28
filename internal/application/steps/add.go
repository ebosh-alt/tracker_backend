package steps

import (
	"context"
	"fmt"
	"tracker/internal/application"
	"tracker/internal/domain/shared"
	domainSteps "tracker/internal/domain/steps"
)

type AddInput struct {
	UserID int64
	Date   shared.LocalDate
	Delta  int
	Source string
}

type AddOutput struct {
	Entry domainSteps.DailyEntry
}

type AddUseCase struct {
	repo domainSteps.Repository
	uow  application.UnitOfWork
}

func NewAddUseCase(repo domainSteps.Repository, uow application.UnitOfWork) *AddUseCase {
	if repo == nil {
		panic(fmt.Errorf("%w: steps repository dependency is required", shared.ErrInvalidInput))
	}

	if uow == nil {
		panic(fmt.Errorf("%w: nil unit of work dependency is required", shared.ErrInvalidInput))

	}
	return &AddUseCase{
		repo: repo,
		uow:  uow,
	}
}

func (uc *AddUseCase) Execute(ctx context.Context, input AddInput) (*AddOutput, error) {
	if input.UserID <= 0 {
		return nil, shared.ErrInvalidInput
	}
	if input.Delta == 0 {
		return nil, shared.ErrInvalidInput
	}

	source, err := domainSteps.ParseSource(input.Source)
	if err != nil {
		return nil, err
	}

	var out domainSteps.DailyEntry
	err = uc.uow.WithinTx(ctx, func(txCtx context.Context) error {
		out, err = uc.repo.AddDelta(txCtx, input.UserID, input.Date, input.Delta, source)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &AddOutput{
		Entry: out,
	}, nil
}
