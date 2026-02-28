package steps

import (
	"context"
	"time"
	"tracker/internal/domain/shared"
	domainSteps "tracker/internal/domain/steps"
)

type AnalyticsInput struct {
	UserID int64
	Month  shared.LocalMonth
}

type Period struct {
	From      shared.LocalDate
	To        shared.LocalDate
	Analytics domainSteps.Analytics
}
type AnalyticsOutput struct {
	GoalPerDay int
	Week       Period
	Month      Period
}

type AnalyticsUseCase struct {
	repo  domainSteps.Repository
	users domainSteps.AnalyticsUserReader
}

func NewAnalyticsUseCase(repo domainSteps.Repository, users domainSteps.AnalyticsUserReader) *AnalyticsUseCase {
	return &AnalyticsUseCase{
		repo:  repo,
		users: users,
	}
}

func (uc *AnalyticsUseCase) Execute(ctx context.Context, in AnalyticsInput) (*AnalyticsOutput, error) {
	if in.UserID <= 0 {
		return nil, shared.ErrInvalidInput
	}
	settings, err := uc.users.GetSettings(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	loc, err := time.LoadLocation(settings.Timezone)
	if err != nil {
		return nil, shared.ErrInvalidInput
	}
	if settings.StepsGoal <= 0 {
		return nil, shared.ErrInvalidInput
	}
	first, err := in.Month.FirstDay(loc)
	if err != nil {
		return nil, shared.ErrInvalidInput
	}
	days, err := in.Month.DaysInMonth(loc)
	if err != nil {
		return nil, shared.ErrInvalidInput
	}
	monthFrom := shared.LocalDate(first.Format("2006-01-02"))
	monthTo := shared.LocalDate(first.AddDate(0, 0, days-1).Format("2006-01-02"))
	monthEntries, err := uc.repo.GetByDateRange(ctx, in.UserID, monthFrom, monthTo)
	if err != nil {
		return nil, err
	}

	todayLocal := time.Now().In(loc)
	weekToTime := time.Date(todayLocal.Year(), todayLocal.Month(), todayLocal.Day(), 0, 0, 0, 0, loc)
	weekFromTime := weekToTime.AddDate(0, 0, -6)
	weekFrom := shared.LocalDate(weekFromTime.Format("2006-01-02"))
	weekTo := shared.LocalDate(weekToTime.Format("2006-01-02"))
	weekEntries, err := uc.repo.GetByDateRange(ctx, in.UserID, weekFrom, weekTo)
	if err != nil {
		return nil, err
	}

	weekAnalytics, err := buildAnalytics(weekEntries, settings.StepsGoal, 7)
	if err != nil {
		return nil, err
	}
	daysInPeriod, err := in.Month.DaysInMonth(loc)
	monthAnalytics, err := buildAnalytics(monthEntries, settings.StepsGoal, daysInPeriod)
	if err != nil {
		return nil, err
	}
	return &AnalyticsOutput{
		GoalPerDay: settings.StepsGoal,
		Week: Period{
			From:      weekFrom,
			To:        weekTo,
			Analytics: weekAnalytics,
		},
		Month: Period{
			From:      monthFrom,
			To:        monthTo,
			Analytics: monthAnalytics,
		},
	}, nil
}

func buildAnalytics(
	entries []domainSteps.DailyEntry,
	goalPerDay int,
	daysInPeriod int,
) (domainSteps.Analytics, error) {
	goalTotal := goalPerDay * daysInPeriod

	factTotal := 0
	series := make([]domainSteps.Point, 0, len(entries))
	for _, e := range entries {
		factTotal += e.Count
		series = append(series, domainSteps.Point{
			Date:  e.Date,
			Steps: e.Count,
		})
	}

	completion := 0.0
	if goalTotal > 0 {
		completion = float64(factTotal) / float64(goalTotal) * 100
	}

	a := domainSteps.Analytics{
		GoalTotal:         goalTotal,
		FactTotal:         factTotal,
		CompletionPercent: completion,
		Series:            series,
	}
	if err := a.Validate(); err != nil {
		return domainSteps.Analytics{}, err
	}
	return a, nil
}
