package steps

import (
	"context"
	"fmt"
	"time"
	"tracker/internal/application"
	"tracker/internal/domain/shared"
	domainSteps "tracker/internal/domain/steps"
)

type PutInput struct {
	UserID int64
	Date   shared.LocalDate
	Count  int
	Source string
}

type PutOutput struct {
	Entry domainSteps.DailyEntry
}

type PutUseCase struct {
	repo domainSteps.Repository
	uow  application.UnitOfWork
}

func NewPutUseCase(repo domainSteps.Repository, uow application.UnitOfWork) *PutUseCase {
	if repo == nil {
		panic(fmt.Errorf("%w: steps repository dependency is required", shared.ErrInvalidInput))
	}

	if uow == nil {
		panic(fmt.Errorf("%w: nil unit of work dependency is required", shared.ErrInvalidInput))

	}

	return &PutUseCase{
		repo: repo,
		uow:  uow,
	}
}

func (uc *PutUseCase) Execute(ctx context.Context, input PutInput) (*PutOutput, error) {
	if input.UserID <= 0 {
		return nil, shared.ErrInvalidInput
	}
	if input.Count < 0 {
		return nil, shared.ErrInvalidInput
	}

	source, err := domainSteps.ParseSource(input.Source)
	if err != nil {
		return nil, err
	}
	daily, err := domainSteps.NewDailyEntry(input.UserID, input.Date, input.Count, source, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	var saved domainSteps.DailyEntry
	err = uc.uow.WithinTx(ctx, func(txCtx context.Context) error {
		var err error
		saved, err = uc.repo.UpsertByDate(txCtx, *daily)
		return err
	})

	if err != nil {
		return nil, err
	}

	return &PutOutput{Entry: saved}, nil

}
