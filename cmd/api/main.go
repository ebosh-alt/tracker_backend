package main

import (
	"context"
	"log"
	"strings"
	"time"

	appauth "tracker/internal/application/auth"
	"tracker/internal/delivery/http/authapi"
	"tracker/internal/delivery/http/middleware"
	"tracker/internal/delivery/http/server"
	infraauth "tracker/internal/infra/auth"
	"tracker/internal/infra/config"
	"tracker/internal/infra/postgres"
	infratelegram "tracker/internal/infra/telegram"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("load config (CONFIG_ENV=dev|prod or CONFIG_PATH): %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := postgres.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	userReader := postgres.NewUserReader(db)
	sessionRepo := postgres.NewAuthSessionRepository(db)

	jwtManager := &infraauth.Manager{
		AccessSecret:  []byte(cfg.Auth.AccessSecret),
		RefreshSecret: []byte(cfg.Auth.RefreshSecret),
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    24 * time.Hour,
	}
	authMiddleware := middleware.RequireJWT(jwtManager)

	tgClient := infratelegram.NewClient(cfg.Telegram.BotToken)
	telegramVerifier := infratelegram.NewInitDataVerifierAdapter(tgClient)
	tokenIssuer := jwtTokenIssuerAdapter{manager: jwtManager}
	tokenHasher := infraauth.SHA256TokenHasher{}
	telegramAuthUC := appauth.NewTelegramAuthUseCase(telegramVerifier, userReader, tokenIssuer, tokenHasher, sessionRepo)
	refreshParser := jwtRefreshParserAdapter{manager: jwtManager}
	refreshUC := appauth.NewRefreshUseCase(refreshParser, sessionRepo, tokenIssuer, tokenHasher)

	authHandler := authapi.NewHandler(authapi.AuthUseCases{
		TelegramAuth: telegramAuthUC,
		Refresh:      refreshUC,
	})

	router := server.NewRouter(authMiddleware, authHandler)
	log.Printf(
		"api starting env=%s addr=%s",
		strings.TrimSpace(cfg.App.Env),
		cfg.Server.Addr(),
	)
	if err := router.Run(cfg.Server.Addr()); err != nil {
		log.Fatalf("run server: %v", err)
	}
}

type jwtTokenIssuerAdapter struct {
	manager *infraauth.Manager
}

func (a jwtTokenIssuerAdapter) IssueDetailed(userID int64) (appauth.IssuedPair, error) {
	pair, err := a.manager.IssueDetailed(userID)
	if err != nil {
		return appauth.IssuedPair{}, err
	}
	return appauth.IssuedPair{
		Access: appauth.IssuedToken{
			Value:     pair.AccessToken.Value,
			JTI:       pair.AccessToken.JTI,
			ExpiresAt: pair.AccessToken.ExpiresAt,
		},
		Refresh: appauth.IssuedToken{
			Value:     pair.RefreshToken.Value,
			JTI:       pair.RefreshToken.JTI,
			ExpiresAt: pair.RefreshToken.ExpiresAt,
		},
		IssuedAt: pair.IssuedAt,
	}, nil
}

func (a jwtTokenIssuerAdapter) Issue(userID int64) (appauth.TokenPair, error) {
	pair, err := a.manager.Issue(userID)
	if err != nil {
		return appauth.TokenPair{}, err
	}
	return appauth.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}

type jwtRefreshParserAdapter struct {
	manager *infraauth.Manager
}

func (a jwtRefreshParserAdapter) ParseDetailed(tokenStr, expectedType string) (*appauth.TokenMeta, error) {
	t, err := a.manager.ParseDetailed(tokenStr, expectedType)
	if err != nil {
		return nil, err
	}
	return &appauth.TokenMeta{
		UserID:    t.UserID,
		TokenType: t.TokenType,
		JTI:       t.JTI,
		ExpiresAt: t.ExpiresAt,
	}, nil
}
