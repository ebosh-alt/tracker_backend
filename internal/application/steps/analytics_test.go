package steps

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	domainSteps "tracker/internal/domain/steps"
	"tracker/internal/domain/shared"
)

type analyticsRepoStub struct {
	weekOut   []domainSteps.DailyEntry
	monthOut  []domainSteps.DailyEntry
	weekErr   error
	monthErr  error
	calls     int
	lastUser  int64
	lastFrom  []shared.LocalDate
	lastTo    []shared.LocalDate
	monthFrom shared.LocalDate
	monthTo   shared.LocalDate
}

func (s *analyticsRepoStub) GetByDateRange(
	ctx context.Context,
	userID int64,
	from shared.LocalDate,
	to shared.LocalDate,
) ([]domainSteps.DailyEntry, error) {
	s.calls++
	s.lastUser = userID
	s.lastFrom = append(s.lastFrom, from)
	s.lastTo = append(s.lastTo, to)

	if from == s.monthFrom && to == s.monthTo {
		if s.monthErr != nil {
			return nil, s.monthErr
		}
		return s.monthOut, nil
	}
	if s.weekErr != nil {
		return nil, s.weekErr
	}
	return s.weekOut, nil
}

func (s *analyticsRepoStub) UpsertByDate(ctx context.Context, entry domainSteps.DailyEntry) (domainSteps.DailyEntry, error) {
	return domainSteps.DailyEntry{}, nil
}

func (s *analyticsRepoStub) AddDelta(
	ctx context.Context,
	userID int64,
	date shared.LocalDate,
	delta int,
	source domainSteps.Source,
) (domainSteps.DailyEntry, error) {
	return domainSteps.DailyEntry{}, nil
}

func (s *analyticsRepoStub) DeleteByDate(ctx context.Context, userID int64, date shared.LocalDate) error {
	return nil
}

type analyticsUserReaderStub struct {
	out      domainSteps.AnalyticsUserSettings
	err      error
	calls    int
	lastUser int64
}

func (s *analyticsUserReaderStub) GetSettings(ctx context.Context, userID int64) (domainSteps.AnalyticsUserSettings, error) {
	s.calls++
	s.lastUser = userID
	if s.err != nil {
		return domainSteps.AnalyticsUserSettings{}, s.err
	}
	return s.out, nil
}

func TestAnalyticsUseCase_Success(t *testing.T) {
	month := mustLocalMonth(t, "2026-02")
	monthFrom := shared.MustLocalDate("2026-02-01")
	monthTo := shared.MustLocalDate("2026-02-28")

	repo := &analyticsRepoStub{
		weekOut: []domainSteps.DailyEntry{
			{UserID: 42, Date: shared.MustLocalDate("2026-02-23"), Count: 7200, Source: domainSteps.SourceManual},
			{UserID: 42, Date: shared.MustLocalDate("2026-02-24"), Count: 9400, Source: domainSteps.SourceManual},
		},
		monthOut: []domainSteps.DailyEntry{
			{UserID: 42, Date: shared.MustLocalDate("2026-02-01"), Count: 5000, Source: domainSteps.SourceManual},
			{UserID: 42, Date: shared.MustLocalDate("2026-02-02"), Count: 8000, Source: domainSteps.SourceManual},
			{UserID: 42, Date: shared.MustLocalDate("2026-02-24"), Count: 9400, Source: domainSteps.SourceManual},
		},
		monthFrom: monthFrom,
		monthTo:   monthTo,
	}
	users := &analyticsUserReaderStub{
		out: domainSteps.AnalyticsUserSettings{
			Timezone:  "Europe/Moscow",
			StepsGoal: 10000,
		},
	}
	uc := NewAnalyticsUseCase(repo, users)

	out, err := uc.Execute(context.Background(), AnalyticsInput{
		UserID: 42,
		Month:  month,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if users.calls != 1 || users.lastUser != 42 {
		t.Fatalf("expected user settings read once for user=42, got calls=%d user=%d", users.calls, users.lastUser)
	}
	if repo.calls != 2 {
		t.Fatalf("expected 2 repo calls (week+month), got %d", repo.calls)
	}
	if repo.lastUser != 42 {
		t.Fatalf("expected userID=42, got %d", repo.lastUser)
	}
	if out.GoalPerDay != 10000 {
		t.Fatalf("expected GoalPerDay=10000, got %d", out.GoalPerDay)
	}

	if out.Month.From != monthFrom || out.Month.To != monthTo {
		t.Fatalf("unexpected month range: from=%q to=%q", out.Month.From, out.Month.To)
	}
	if out.Month.Analytics.GoalTotal != 280000 {
		t.Fatalf("expected month goalTotal=280000, got %d", out.Month.Analytics.GoalTotal)
	}
	if out.Month.Analytics.FactTotal != 22400 {
		t.Fatalf("expected month factTotal=22400, got %d", out.Month.Analytics.FactTotal)
	}
	if len(out.Month.Analytics.Series) != 3 {
		t.Fatalf("expected month series len=3, got %d", len(out.Month.Analytics.Series))
	}

	if out.Week.Analytics.GoalTotal != 70000 {
		t.Fatalf("expected week goalTotal=70000, got %d", out.Week.Analytics.GoalTotal)
	}
	if out.Week.Analytics.FactTotal != 16600 {
		t.Fatalf("expected week factTotal=16600, got %d", out.Week.Analytics.FactTotal)
	}
	if len(out.Week.Analytics.Series) != 2 {
		t.Fatalf("expected week series len=2, got %d", len(out.Week.Analytics.Series))
	}

	wantWeekPct := 16600.0 / 70000.0 * 100
	if math.Abs(out.Week.Analytics.CompletionPercent-wantWeekPct) > 0.01 {
		t.Fatalf("unexpected week completion: got=%.4f want=%.4f", out.Week.Analytics.CompletionPercent, wantWeekPct)
	}
	wantMonthPct := 22400.0 / 280000.0 * 100
	if math.Abs(out.Month.Analytics.CompletionPercent-wantMonthPct) > 0.01 {
		t.Fatalf("unexpected month completion: got=%.4f want=%.4f", out.Month.Analytics.CompletionPercent, wantMonthPct)
	}

	// week range should be a valid 7-day window.
	if err := out.Week.From.Validate(); err != nil {
		t.Fatalf("invalid week from: %v", err)
	}
	if err := out.Week.To.Validate(); err != nil {
		t.Fatalf("invalid week to: %v", err)
	}
	weekFromTime, err := out.Week.From.TimeIn(time.UTC)
	if err != nil {
		t.Fatalf("week from parse failed: %v", err)
	}
	weekToTime, err := out.Week.To.TimeIn(time.UTC)
	if err != nil {
		t.Fatalf("week to parse failed: %v", err)
	}
	if weekToTime.Before(weekFromTime) {
		t.Fatalf("invalid week range: from=%q to=%q", out.Week.From, out.Week.To)
	}
	if int(weekToTime.Sub(weekFromTime).Hours()/24) != 6 {
		t.Fatalf("expected 7-day window (diff 6), got from=%q to=%q", out.Week.From, out.Week.To)
	}
}

func TestAnalyticsUseCase_InvalidInput(t *testing.T) {
	repo := &analyticsRepoStub{}
	users := &analyticsUserReaderStub{}
	uc := NewAnalyticsUseCase(repo, users)

	_, err := uc.Execute(context.Background(), AnalyticsInput{
		UserID: 0,
		Month:  mustLocalMonth(t, "2026-02"),
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if users.calls != 0 || repo.calls != 0 {
		t.Fatalf("expected no dependency calls, got users=%d repo=%d", users.calls, repo.calls)
	}
}

func TestAnalyticsUseCase_UserSettingsError(t *testing.T) {
	repo := &analyticsRepoStub{}
	userErr := errors.New("user settings failed")
	users := &analyticsUserReaderStub{err: userErr}
	uc := NewAnalyticsUseCase(repo, users)

	_, err := uc.Execute(context.Background(), AnalyticsInput{
		UserID: 42,
		Month:  mustLocalMonth(t, "2026-02"),
	})
	if !errors.Is(err, userErr) {
		t.Fatalf("expected user settings error, got %v", err)
	}
	if repo.calls != 0 {
		t.Fatalf("repo should not be called, got %d", repo.calls)
	}
}

func TestAnalyticsUseCase_InvalidUserSettings(t *testing.T) {
	repo := &analyticsRepoStub{}
	users := &analyticsUserReaderStub{
		out: domainSteps.AnalyticsUserSettings{
			Timezone:  "Bad/Timezone",
			StepsGoal: 0,
		},
	}
	uc := NewAnalyticsUseCase(repo, users)

	_, err := uc.Execute(context.Background(), AnalyticsInput{
		UserID: 42,
		Month:  mustLocalMonth(t, "2026-02"),
	})
	if !errors.Is(err, shared.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if repo.calls != 0 {
		t.Fatalf("repo should not be called, got %d", repo.calls)
	}
}

func TestAnalyticsUseCase_WeekRepoError(t *testing.T) {
	month := mustLocalMonth(t, "2026-02")
	repoErr := errors.New("week query failed")
	repo := &analyticsRepoStub{
		weekErr:   repoErr,
		monthFrom: shared.MustLocalDate("2026-02-01"),
		monthTo:   shared.MustLocalDate("2026-02-28"),
	}
	users := &analyticsUserReaderStub{
		out: domainSteps.AnalyticsUserSettings{
			Timezone:  "Europe/Moscow",
			StepsGoal: 10000,
		},
	}
	uc := NewAnalyticsUseCase(repo, users)

	_, err := uc.Execute(context.Background(), AnalyticsInput{
		UserID: 42,
		Month:  month,
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected week repo error, got %v", err)
	}
}

func TestAnalyticsUseCase_MonthRepoError(t *testing.T) {
	month := mustLocalMonth(t, "2026-02")
	repoErr := errors.New("month query failed")
	repo := &analyticsRepoStub{
		monthErr:  repoErr,
		monthFrom: shared.MustLocalDate("2026-02-01"),
		monthTo:   shared.MustLocalDate("2026-02-28"),
	}
	users := &analyticsUserReaderStub{
		out: domainSteps.AnalyticsUserSettings{
			Timezone:  "Europe/Moscow",
			StepsGoal: 10000,
		},
	}
	uc := NewAnalyticsUseCase(repo, users)

	_, err := uc.Execute(context.Background(), AnalyticsInput{
		UserID: 42,
		Month:  month,
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected month repo error, got %v", err)
	}
}

func mustLocalMonth(t *testing.T, raw string) shared.LocalMonth {
	t.Helper()
	v, err := shared.ParseLocalMonth(raw)
	if err != nil {
		t.Fatalf("parse month %q: %v", raw, err)
	}
	return v
}
