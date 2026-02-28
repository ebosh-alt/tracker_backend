package meapi

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

	appuser "tracker/internal/application/user"
	"tracker/internal/domain/shared"
	domainUser "tracker/internal/domain/user"
)

type updateSettingsUseCaseStub struct {
	executeFn func(ctx context.Context, in appuser.UpdateSettingsInput) (*appuser.UpdateSettingsOutput, error)
}

func (s updateSettingsUseCaseStub) Execute(ctx context.Context, in appuser.UpdateSettingsInput) (*appuser.UpdateSettingsOutput, error) {
	if s.executeFn == nil {
		return &appuser.UpdateSettingsOutput{}, nil
	}
	return s.executeFn(ctx, in)
}

func TestUpdateSettingsHandler_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	uc := updateSettingsUseCaseStub{
		executeFn: func(_ context.Context, in appuser.UpdateSettingsInput) (*appuser.UpdateSettingsOutput, error) {
			if in.UserID != 42 {
				t.Fatalf("expected userID=42, got %d", in.UserID)
			}
			if in.Timezone == nil || *in.Timezone != "Europe/Moscow" {
				t.Fatalf("unexpected timezone input: %#v", in.Timezone)
			}
			if in.StepsGoal == nil || *in.StepsGoal != 12000 {
				t.Fatalf("unexpected stepsGoal input: %#v", in.StepsGoal)
			}
			goal, _ := domainUser.NewStepsGoal(12000)
			return &appuser.UpdateSettingsOutput{
				User: domainUser.User{
					ID:         42,
					TelegramID: 123456789,
					Username:   "john",
					FirstName:  "John",
					LastName:   "Doe",
					Timezone:   "Europe/Moscow",
					StepsGoal:  goal,
					Streak:     3,
					CreatedAt:  time.Date(2026, 2, 22, 10, 0, 0, 0, time.UTC),
					UpdatedAt:  time.Date(2026, 2, 27, 12, 0, 0, 0, time.UTC),
				},
			}, nil
		},
	}

	router := buildMeRouter(uc, true)

	req := httptest.NewRequest(http.MethodPatch, "/api/me/settings", bytes.NewBufferString(`{"timezone":"Europe/Moscow","stepsGoal":12000}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	user, ok := body["user"].(map[string]any)
	if !ok {
		t.Fatalf("expected user object, got %#v", body["user"])
	}
	if user["stepsGoal"].(float64) != 12000 {
		t.Fatalf("expected stepsGoal=12000, got %v", user["stepsGoal"])
	}
}

func TestUpdateSettingsHandler_Unauthorized(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := buildMeRouter(updateSettingsUseCaseStub{}, false)
	req := httptest.NewRequest(http.MethodPatch, "/api/me/settings", bytes.NewBufferString(`{"stepsGoal":12000}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateSettingsHandler_InvalidJSON(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := buildMeRouter(updateSettingsUseCaseStub{}, true)
	req := httptest.NewRequest(http.MethodPatch, "/api/me/settings", bytes.NewBufferString(`{"stepsGoal":`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateSettingsHandler_ErrorMapping(t *testing.T) {
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
			router := buildMeRouter(updateSettingsUseCaseStub{
				executeFn: func(_ context.Context, _ appuser.UpdateSettingsInput) (*appuser.UpdateSettingsOutput, error) {
					return nil, tc.err
				},
			}, true)

			req := httptest.NewRequest(http.MethodPatch, "/api/me/settings", bytes.NewBufferString(`{"stepsGoal":12000}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Fatalf("expected status %d, got %d: %s", tc.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func buildMeRouter(updateUC UpdateSettingsUseCase, withUser bool) *gin.Engine {
	router := gin.New()
	if withUser {
		router.Use(func(c *gin.Context) {
			c.Set("user_id", int64(42))
			c.Next()
		})
	}
	api := router.Group("/api")
	RegisterRoutes(api, NewHandler(MeUseCases{
		UpdateSettings: updateUC,
	}))
	return router
}
