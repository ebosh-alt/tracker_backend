package steps

import (
	"context"

	"tracker/internal/domain/shared"
)

type repositoryContractStub struct{}

func (repositoryContractStub) GetByDateRange(
	ctx context.Context,
	userID int64,
	from shared.LocalDate,
	to shared.LocalDate,
) ([]DailyEntry, error) {
	return nil, nil
}

func (repositoryContractStub) UpsertByDate(
	ctx context.Context,
	entry DailyEntry,
) (DailyEntry, error) {
	return DailyEntry{}, nil
}

func (repositoryContractStub) AddDelta(
	ctx context.Context,
	userID int64,
	date shared.LocalDate,
	delta int,
	source Source,
) (DailyEntry, error) {
	return DailyEntry{}, nil
}

func (repositoryContractStub) DeleteByDate(
	ctx context.Context,
	userID int64,
	date shared.LocalDate,
) error {
	return nil
}

var _ Repository = (*repositoryContractStub)(nil)

type analyticsUserReaderContractStub struct{}

func (analyticsUserReaderContractStub) GetSettings(ctx context.Context, userID int64) (AnalyticsUserSettings, error) {
	return AnalyticsUserSettings{}, nil
}

var _ AnalyticsUserReader = (*analyticsUserReaderContractStub)(nil)
