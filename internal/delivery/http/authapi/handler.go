package authapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	appauth "tracker/internal/application/auth"
	"tracker/internal/domain/shared"
)

// TelegramAuthUseCase контракт use case авторизации через Telegram.
type TelegramAuthUseCase interface {
	Execute(ctx context.Context, in appauth.TelegramAuthInput) (appauth.TelegramAuthOutput, error)
}

type RefreshUseCase interface {
	Execute(ctx context.Context, in *appauth.RefreshInput) (*appauth.RefreshOutput, error)
}

// AuthUseCases агрегирует зависимости auth handler.
type AuthUseCases struct {
	TelegramAuth TelegramAuthUseCase
	Refresh      RefreshUseCase
}

// Handler HTTP-обработчики auth API.
type Handler struct {
	useCases AuthUseCases
}

// NewHandler создает handler auth API.
func NewHandler(useCases AuthUseCases) *Handler {
	return &Handler{
		useCases: useCases,
	}
}

// RegisterRoutes подключает публичные auth endpoint'ы.
func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {
	api.POST("/auth/telegram", handler.TelegramAuth)
	api.POST("/auth/refresh", handler.Refresh)
}

// TelegramAuth обрабатывает вход через Telegram initData.
func (h *Handler) TelegramAuth(c *gin.Context) {
	var body telegramAuthRequestATO
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, errors.Join(shared.ErrInvalidInput, err))
		return
	}

	var timezone *string
	if value := strings.TrimSpace(c.GetHeader("X-Timezone")); value != "" {
		timezone = &value
	}
	ipRaw := strings.TrimSpace(c.ClientIP())
	uaRaw := strings.TrimSpace(c.GetHeader("User-Agent"))

	var ip *string
	if ipRaw != "" {
		ip = &ipRaw
	}
	var ua *string
	if uaRaw != "" {
		ua = &uaRaw
	}

	out, err := h.useCases.TelegramAuth.Execute(c.Request.Context(), appauth.TelegramAuthInput{
		InitData:  body.InitData,
		Timezone:  timezone,
		IP:        ip,
		UserAgent: ua,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapTelegramAuthResponse(out))
}

func (h *Handler) Refresh(c *gin.Context) {
	var body refreshRequestATO
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, errors.Join(shared.ErrInvalidInput, err))
		return
	}
	ipRaw := strings.TrimSpace(c.ClientIP())
	uaRaw := strings.TrimSpace(c.GetHeader("User-Agent"))

	var ip *string
	if ipRaw != "" {
		ip = &ipRaw
	}
	var ua *string
	if uaRaw != "" {
		ua = &uaRaw
	}

	appRefresh := appauth.RefreshInput{
		RefreshToken: body.RefreshToken,
		IP:           ip,
		UserAgent:    ua,
	}
	refresh, err := h.useCases.Refresh.Execute(c.Request.Context(), &appRefresh)
	if err != nil {
		writeError(c, err)
		return
	}
	resp := mapRefreshResponse(*refresh)
	c.JSON(http.StatusOK, resp)

}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, shared.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"code": "invalid_input", "error": err.Error()})
	case errors.Is(err, shared.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "error": err.Error()})
	case errors.Is(err, shared.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"code": "not_found", "error": err.Error()})
	case errors.Is(err, shared.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"code": "conflict", "error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "error": "internal server error"})
	}
}
