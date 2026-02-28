package steps

import (
	"context"
	"fmt"
	"tracker/internal/domain/shared"
	domainSteps "tracker/internal/domain/steps"
)

type ListInput struct {
	UserID int64
	From   shared.LocalDate
	To     shared.LocalDate
}

type ListOutput struct {
	Entries []domainSteps.DailyEntry
}

type ListUseCase struct {
	repo domainSteps.Repository
}

func NewListUseCase(repo domainSteps.Repository) *ListUseCase {
	if repo == nil {
		panic(fmt.Errorf("%w: steps repository dependency is required", shared.ErrInvalidInput))
	}
	return &ListUseCase{
		repo: repo,
	}
}

func (uc *ListUseCase) Execute(ctx context.Context, input ListInput) (*ListOutput, error) {
	if input.UserID <= 0 {
		return nil, shared.ErrInvalidInput
	}
	if string(input.From) > string(input.To) {
		return nil, shared.ErrInvalidInput
	}
	entries, err := uc.repo.GetByDateRange(ctx, input.UserID, input.From, input.To)
	if err != nil {
		return nil, err
	}
	out := &ListOutput{
		Entries: entries,
	}
	return out, nil
}
