package stepsapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	appsteps "tracker/internal/application/steps"
	domainSteps "tracker/internal/domain/steps"
	"tracker/internal/domain/shared"
)

type listUseCaseStub struct {
	executeFn func(ctx context.Context, in appsteps.ListInput) (*appsteps.ListOutput, error)
}

func (s listUseCaseStub) Execute(ctx context.Context, in appsteps.ListInput) (*appsteps.ListOutput, error) {
	if s.executeFn == nil {
		return &appsteps.ListOutput{}, nil
	}
	return s.executeFn(ctx, in)
}

type putUseCaseStub struct {
	executeFn func(ctx context.Context, in appsteps.PutInput) (*appsteps.PutOutput, error)
}

func (s putUseCaseStub) Execute(ctx context.Context, in appsteps.PutInput) (*appsteps.PutOutput, error) {
	if s.executeFn == nil {
		return &appsteps.PutOutput{}, nil
	}
	return s.executeFn(ctx, in)
}

type addUseCaseStub struct {
	executeFn func(ctx context.Context, in appsteps.AddInput) (*appsteps.AddOutput, error)
}

func (s addUseCaseStub) Execute(ctx context.Context, in appsteps.AddInput) (*appsteps.AddOutput, error) {
	if s.executeFn == nil {
		return &appsteps.AddOutput{}, nil
	}
	return s.executeFn(ctx, in)
}

type deleteUseCaseStub struct {
	executeFn func(ctx context.Context, in appsteps.DeleteInput) error
}

func (s deleteUseCaseStub) Execute(ctx context.Context, in appsteps.DeleteInput) error {
	if s.executeFn == nil {
		return nil
	}
	return s.executeFn(ctx, in)
}

type analyticsUseCaseStub struct {
	executeFn func(ctx context.Context, in appsteps.AnalyticsInput) (*appsteps.AnalyticsOutput, error)
}

func (s analyticsUseCaseStub) Execute(ctx context.Context, in appsteps.AnalyticsInput) (*appsteps.AnalyticsOutput, error) {
	if s.executeFn == nil {
		return &appsteps.AnalyticsOutput{}, nil
	}
	return s.executeFn(ctx, in)
}

func TestListStepsHandler_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	uc := listUseCaseStub{
		executeFn: func(_ context.Context, in appsteps.ListInput) (*appsteps.ListOutput, error) {
			if in.UserID != 42 {
				t.Fatalf("expected userID=42, got %d", in.UserID)
			}
			if in.From != shared.LocalDate("2026-02-24") || in.To != shared.LocalDate("2026-02-27") {
				t.Fatalf("unexpected range: from=%q to=%q", in.From, in.To)
			}
			return &appsteps.ListOutput{
				Entries: []domainSteps.DailyEntry{
					{
						UserID:    42,
						Date:      shared.MustLocalDate("2026-02-24"),
						Count:     7200,
						Source:    domainSteps.SourceManual,
						CreatedAt: time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2026, 2, 24, 11, 0, 0, 0, time.UTC),
					},
				},
			}, nil
		},
	}

	router := buildStepsRouter(uc, putUseCaseStub{}, addUseCaseStub{}, deleteUseCaseStub{}, analyticsUseCaseStub{}, true)

	req := httptest.NewRequest(http.MethodGet, "/api/steps?from=2026-02-24&to=2026-02-27", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	entries, ok := body["entries"].([]any)
	if !ok || len(entries) != 1 {
		t.Fatalf("expected entries len=1, got %v", body["entries"])
	}
	entry := entries[0].(map[string]any)
	if entry["count"].(float64) != 7200 {
		t.Fatalf("expected count=7200, got %v", entry["count"])
	}
}

func TestPutStepsHandler_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	uc := putUseCaseStub{
		executeFn: func(_ context.Context, in appsteps.PutInput) (*appsteps.PutOutput, error) {
			if in.UserID != 42 {
				t.Fatalf("expected userID=42, got %d", in.UserID)
			}
			if in.Date != shared.LocalDate("2026-02-27") {
				t.Fatalf("unexpected date: %q", in.Date)
			}
			if in.Count != 1500 || in.Source != "manual" {
				t.Fatalf("unexpected input: %+v", in)
			}
			return &appsteps.PutOutput{
				Entry: domainSteps.DailyEntry{
					UserID:    42,
					Date:      shared.MustLocalDate("2026-02-27"),
					Count:     1500,
					Source:    domainSteps.SourceManual,
					CreatedAt: time.Date(2026, 2, 27, 10, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2026, 2, 27, 10, 5, 0, 0, time.UTC),
				},
			}, nil
		},
	}
	router := buildStepsRouter(listUseCaseStub{}, uc, addUseCaseStub{}, deleteUseCaseStub{}, analyticsUseCaseStub{}, true)

	req := httptest.NewRequest(http.MethodPut, "/api/steps/2026-02-27", bytes.NewBufferString(`{"count":1500,"source":"manual"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAddStepsHandler_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	uc := addUseCaseStub{
		executeFn: func(_ context.Context, in appsteps.AddInput) (*appsteps.AddOutput, error) {
			if in.UserID != 42 {
				t.Fatalf("expected userID=42, got %d", in.UserID)
			}
			if in.Date != shared.LocalDate("2026-02-27") || in.Delta != 500 || in.Source != "manual" {
				t.Fatalf("unexpected input: %+v", in)
			}
			return &appsteps.AddOutput{
				Entry: domainSteps.DailyEntry{
					UserID:    42,
					Date:      shared.MustLocalDate("2026-02-27"),
					Count:     2000,
					Source:    domainSteps.SourceManual,
					CreatedAt: time.Date(2026, 2, 27, 10, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2026, 2, 27, 10, 10, 0, 0, time.UTC),
				},
			}, nil
		},
	}
	router := buildStepsRouter(listUseCaseStub{}, putUseCaseStub{}, uc, deleteUseCaseStub{}, analyticsUseCaseStub{}, true)

	req := httptest.NewRequest(http.MethodPost, "/api/steps/add", bytes.NewBufferString(`{"date":"2026-02-27","delta":500,"source":"manual"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteStepsHandler_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	uc := deleteUseCaseStub{
		executeFn: func(_ context.Context, in appsteps.DeleteInput) error {
			if in.UserID != 42 {
				t.Fatalf("expected userID=42, got %d", in.UserID)
			}
			if in.Date != shared.LocalDate("2026-02-27") {
				t.Fatalf("unexpected date: %q", in.Date)
			}
			return nil
		},
	}
	router := buildStepsRouter(listUseCaseStub{}, putUseCaseStub{}, addUseCaseStub{}, uc, analyticsUseCaseStub{}, true)

	req := httptest.NewRequest(http.MethodDelete, "/api/steps/2026-02-27", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAnalyticsStepsHandler_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	uc := analyticsUseCaseStub{
		executeFn: func(_ context.Context, in appsteps.AnalyticsInput) (*appsteps.AnalyticsOutput, error) {
			if in.UserID != 42 {
				t.Fatalf("expected userID=42, got %d", in.UserID)
			}
			if in.Month != shared.LocalMonth("2026-02") {
				t.Fatalf("unexpected month: %q", in.Month)
			}
			return &appsteps.AnalyticsOutput{
				GoalPerDay: 10000,
				Week: appsteps.Period{
					From: shared.MustLocalDate("2026-02-21"),
					To:   shared.MustLocalDate("2026-02-27"),
					Analytics: domainSteps.Analytics{
						GoalTotal:         70000,
						FactTotal:         54230,
						CompletionPercent: 77.47,
						Series: []domainSteps.Point{
							{Date: shared.MustLocalDate("2026-02-24"), Steps: 7200},
						},
					},
				},
				Month: appsteps.Period{
					From: shared.MustLocalDate("2026-02-01"),
					To:   shared.MustLocalDate("2026-02-28"),
					Analytics: domainSteps.Analytics{
						GoalTotal:         280000,
						FactTotal:         163450,
						CompletionPercent: 58.38,
						Series: []domainSteps.Point{
							{Date: shared.MustLocalDate("2026-02-01"), Steps: 4500},
						},
					},
				},
			}, nil
		},
	}
	router := buildStepsRouter(listUseCaseStub{}, putUseCaseStub{}, addUseCaseStub{}, deleteUseCaseStub{}, uc, true)

	req := httptest.NewRequest(http.MethodGet, "/api/steps/analytics?month=2026-02", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["goalPerDay"].(float64) != 10000 {
		t.Fatalf("expected goalPerDay=10000, got %v", body["goalPerDay"])
	}
}

func TestStepsHandler_UnauthorizedWhenUserMissing(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := buildStepsRouter(listUseCaseStub{}, putUseCaseStub{}, addUseCaseStub{}, deleteUseCaseStub{}, analyticsUseCaseStub{}, false)

	req := httptest.NewRequest(http.MethodGet, "/api/steps?from=2026-02-24&to=2026-02-27", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStepsHandler_BadRequestOnInvalidPayload(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := buildStepsRouter(listUseCaseStub{}, putUseCaseStub{}, addUseCaseStub{}, deleteUseCaseStub{}, analyticsUseCaseStub{}, true)

	req := httptest.NewRequest(http.MethodPost, "/api/steps/add", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStepsHandler_ErrorMapping(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name string
		err  error
		want int
	}{
		{name: "invalid input", err: shared.ErrInvalidInput, want: http.StatusBadRequest},
		{name: "unauthorized", err: shared.ErrUnauthorized, want: http.StatusUnauthorized},
		{name: "not found", err: shared.ErrNotFound, want: http.StatusNotFound},
		{name: "conflict", err: shared.ErrConflict, want: http.StatusConflict},
		{name: "internal", err: errors.New("db down"), want: http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := buildStepsRouter(
				listUseCaseStub{
					executeFn: func(_ context.Context, _ appsteps.ListInput) (*appsteps.ListOutput, error) {
						return nil, tc.err
					},
				},
				putUseCaseStub{},
				addUseCaseStub{},
				deleteUseCaseStub{},
				analyticsUseCaseStub{},
				true,
			)

			req := httptest.NewRequest(http.MethodGet, "/api/steps?from=2026-02-24&to=2026-02-27", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Fatalf("expected status %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func buildStepsRouter(
	listUC ListUseCase,
	putUC PutUseCase,
	addUC AddUseCase,
	deleteUC DeleteUseCase,
	analyticsUC AnalyticsUseCase,
	withUser bool,
) *gin.Engine {
	router := gin.New()
	if withUser {
		router.Use(func(c *gin.Context) {
			c.Set("user_id", int64(42))
			c.Next()
		})
	}
	api := router.Group("/api")
	RegisterRoutes(api, NewHandler(StepsUseCases{
		List:      listUC,
		Put:       putUC,
		Add:       addUC,
		Delete:    deleteUC,
		Analytics: analyticsUC,
	}))
	return router
}
