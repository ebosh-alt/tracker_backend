package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tracker/internal/delivery/http/authapi"
)

// NewRouter собирает HTTP-роутер приложения.
func NewRouter(
	authMiddleware gin.HandlerFunc,
	authHandler *authapi.Handler,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	publicAPI := router.Group("/api")
	authapi.RegisterRoutes(publicAPI, authHandler)

	protectedAPI := router.Group("/api")
	protectedAPI.Use(authMiddleware)

	return router
}
