package authapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	appauth "tracker/internal/application/auth"
	"tracker/internal/domain/shared"
	domainUser "tracker/internal/domain/user"
)

type telegramAuthUseCaseStub struct {
	executeFn func(ctx context.Context, in appauth.TelegramAuthInput) (appauth.TelegramAuthOutput, error)
}

func (s telegramAuthUseCaseStub) Execute(ctx context.Context, in appauth.TelegramAuthInput) (appauth.TelegramAuthOutput, error) {
	if s.executeFn == nil {
		return appauth.TelegramAuthOutput{}, nil
	}
	return s.executeFn(ctx, in)
}

type refreshUseCaseStub struct {
	executeFn func(ctx context.Context, in *appauth.RefreshInput) (*appauth.RefreshOutput, error)
}

func (s refreshUseCaseStub) Execute(ctx context.Context, in *appauth.RefreshInput) (*appauth.RefreshOutput, error) {
	if s.executeFn == nil {
		return &appauth.RefreshOutput{}, nil
	}
	return s.executeFn(ctx, in)
}

func TestTelegramAuthHandlerSuccess(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	goal, _ := domainUser.NewStepsGoal(10000)
	uc := telegramAuthUseCaseStub{
		executeFn: func(_ context.Context, in appauth.TelegramAuthInput) (appauth.TelegramAuthOutput, error) {
			if in.InitData != "query_id=ok" {
				t.Fatalf("expected initData to be passed")
			}
			if in.Timezone == nil || *in.Timezone != "Europe/Moscow" {
				t.Fatalf("expected timezone from header, got %#v", in.Timezone)
			}

			return appauth.TelegramAuthOutput{
				User: domainUser.User{
					ID:         11,
					TelegramID: 100500,
					Username:   "john",
					FirstName:  "John",
					LastName:   "Doe",
					Timezone:   "Europe/Moscow",
					StepsGoal:  goal,
					Streak:     3,
					CreatedAt:  time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC),
					UpdatedAt:  time.Date(2026, 2, 24, 11, 0, 0, 0, time.UTC),
				},
				Token: appauth.TokenPair{
					AccessToken:  "access",
					RefreshToken: "refresh",
				},
			}, nil
		},
	}
	router := buildRouter(uc, refreshUseCaseStub{})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/telegram", bytes.NewBufferString(`{"initData":"query_id=ok"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Timezone", "Europe/Moscow")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	token := body["token"].(map[string]any)
	if token["accessToken"] != "access" {
		t.Fatalf("expected access token")
	}
	user := body["user"].(map[string]any)
	if user["id"].(float64) != 11 {
		t.Fatalf("expected user id 11")
	}
}

func TestTelegramAuthHandlerInvalidPayload(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := buildRouter(telegramAuthUseCaseStub{}, refreshUseCaseStub{})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/telegram", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTelegramAuthHandlerUnauthorized(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	uc := telegramAuthUseCaseStub{
		executeFn: func(_ context.Context, _ appauth.TelegramAuthInput) (appauth.TelegramAuthOutput, error) {
			return appauth.TelegramAuthOutput{}, shared.ErrUnauthorized
		},
	}
	router := buildRouter(uc, refreshUseCaseStub{})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/telegram", bytes.NewBufferString(`{"initData":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRefreshHandlerSuccess(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	refreshUC := refreshUseCaseStub{
		executeFn: func(_ context.Context, in *appauth.RefreshInput) (*appauth.RefreshOutput, error) {
			if in == nil {
				t.Fatal("expected refresh input")
			}
			if in.RefreshToken != "refresh.raw.jwt" {
				t.Fatalf("expected refresh token, got %q", in.RefreshToken)
			}
			if in.IP == nil || *in.IP == "" {
				t.Fatal("expected non-empty client ip")
			}
			if in.UserAgent == nil || *in.UserAgent != "tracker-test-agent" {
				t.Fatalf("expected user agent header, got %#v", in.UserAgent)
			}
			return &appauth.RefreshOutput{
				Token: appauth.TokenPair{
					AccessToken:  "access.new.jwt",
					RefreshToken: "refresh.new.jwt",
				},
			}, nil
		},
	}
	router := buildRouter(telegramAuthUseCaseStub{}, refreshUC)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewBufferString(`{"refreshToken":"refresh.raw.jwt"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "tracker-test-agent")
	req.RemoteAddr = "10.10.10.10:12345"
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	token := body["token"].(map[string]any)
	if token["accessToken"] != "access.new.jwt" || token["refreshToken"] != "refresh.new.jwt" {
		t.Fatalf("unexpected refresh response: %v", body)
	}
}

func TestRefreshHandlerInvalidPayload(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	router := buildRouter(telegramAuthUseCaseStub{}, refreshUseCaseStub{})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRefreshHandlerUnauthorized(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	refreshUC := refreshUseCaseStub{
		executeFn: func(_ context.Context, in *appauth.RefreshInput) (*appauth.RefreshOutput, error) {
			return nil, shared.ErrUnauthorized
		},
	}
	router := buildRouter(telegramAuthUseCaseStub{}, refreshUC)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewBufferString(`{"refreshToken":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func buildRouter(authUC TelegramAuthUseCase, refreshUC RefreshUseCase) *gin.Engine {
	router := gin.New()
	api := router.Group("/api")
	RegisterRoutes(api, NewHandler(AuthUseCases{
		TelegramAuth: authUC,
		Refresh:      refreshUC,
	}))
	return router
}
