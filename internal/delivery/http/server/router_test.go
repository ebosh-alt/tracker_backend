package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	appauth "tracker/internal/application/auth"
	appsteps "tracker/internal/application/steps"
	appuser "tracker/internal/application/user"
	"tracker/internal/delivery/http/authapi"
	"tracker/internal/delivery/http/meapi"
	"tracker/internal/delivery/http/stepsapi"
)

type authTelegramStub struct{}

func (authTelegramStub) Execute(ctx context.Context, in appauth.TelegramAuthInput) (appauth.TelegramAuthOutput, error) {
	return appauth.TelegramAuthOutput{}, nil
}

type authRefreshStub struct{}

func (authRefreshStub) Execute(ctx context.Context, in *appauth.RefreshInput) (*appauth.RefreshOutput, error) {
	return &appauth.RefreshOutput{}, nil
}

type stepsListStub struct{}

func (stepsListStub) Execute(ctx context.Context, in appsteps.ListInput) (*appsteps.ListOutput, error) {
	return &appsteps.ListOutput{}, nil
}

type stepsPutStub struct{}

func (stepsPutStub) Execute(ctx context.Context, in appsteps.PutInput) (*appsteps.PutOutput, error) {
	return &appsteps.PutOutput{}, nil
}

type stepsAddStub struct{}

func (stepsAddStub) Execute(ctx context.Context, in appsteps.AddInput) (*appsteps.AddOutput, error) {
	return &appsteps.AddOutput{}, nil
}

type stepsDeleteStub struct{}

func (stepsDeleteStub) Execute(ctx context.Context, in appsteps.DeleteInput) error {
	return nil
}

type stepsAnalyticsStub struct{}

func (stepsAnalyticsStub) Execute(ctx context.Context, in appsteps.AnalyticsInput) (*appsteps.AnalyticsOutput, error) {
	return &appsteps.AnalyticsOutput{}, nil
}

type meUpdateSettingsStub struct{}

func (meUpdateSettingsStub) Execute(ctx context.Context, in appuser.UpdateSettingsInput) (*appuser.UpdateSettingsOutput, error) {
	return &appuser.UpdateSettingsOutput{}, nil
}

func TestNewRouter_StepsRoutesAreProtected(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	authHandler := authapi.NewHandler(authapi.AuthUseCases{
		TelegramAuth: authTelegramStub{},
		Refresh:      authRefreshStub{},
	})
	stepsHandler := stepsapi.NewHandler(stepsapi.StepsUseCases{
		List:      stepsListStub{},
		Put:       stepsPutStub{},
		Add:       stepsAddStub{},
		Delete:    stepsDeleteStub{},
		Analytics: stepsAnalyticsStub{},
	})
	meHandler := meapi.NewHandler(meapi.MeUseCases{
		UpdateSettings: meUpdateSettingsStub{},
	})

	mwCalls := 0
	authMiddleware := func(c *gin.Context) {
		mwCalls++
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"code":  "unauthorized",
			"error": "blocked by test middleware",
		})
	}

	router, err := NewRouter(authMiddleware, authHandler, stepsHandler, meHandler, []string{"127.0.0.1"})
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/steps?from=2026-02-24&to=2026-02-27", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if mwCalls == 0 {
		t.Fatal("expected auth middleware to be called for /api/steps")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}
