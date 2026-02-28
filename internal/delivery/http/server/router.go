package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"tracker/internal/delivery/http/authapi"
	"tracker/internal/delivery/http/meapi"
	"tracker/internal/delivery/http/stepsapi"
)

// NewRouter собирает HTTP-роутер приложения.
func NewRouter(
	authMiddleware gin.HandlerFunc,
	authHandler *authapi.Handler,
	stepsHandler *stepsapi.Handler,
	meHandler *meapi.Handler,
	trustedProxies []string,
) (*gin.Engine, error) {
	router := gin.New()
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		return nil, fmt.Errorf("set trusted proxies: %w", err)
	}
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	publicAPI := router.Group("/api")
	authapi.RegisterRoutes(publicAPI, authHandler)

	protectedAPI := router.Group("/api")
	protectedAPI.Use(authMiddleware)
	stepsapi.RegisterRoutes(protectedAPI, stepsHandler)
	meapi.RegisterRoutes(protectedAPI, meHandler)

	return router, nil
}
