package meapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	appuser "tracker/internal/application/user"
	"tracker/internal/delivery/http/middleware"
	"tracker/internal/domain/shared"
)

type UpdateSettingsUseCase interface {
	Execute(ctx context.Context, in appuser.UpdateSettingsInput) (*appuser.UpdateSettingsOutput, error)
}

type MeUseCases struct {
	UpdateSettings UpdateSettingsUseCase
}

type Handler struct {
	useCases MeUseCases
}

func NewHandler(useCases MeUseCases) *Handler {
	return &Handler{useCases: useCases}
}

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {
	api.PATCH("/me/settings", handler.UpdateSettings)
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, shared.ErrUnauthorized)
		return
	}

	var body updateSettingsRequestATO
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, errors.Join(shared.ErrInvalidInput, err))
		return
	}

	out, err := h.useCases.UpdateSettings.Execute(c.Request.Context(), appuser.UpdateSettingsInput{
		UserID:    userID,
		Timezone:  body.Timezone,
		StepsGoal: body.StepsGoal,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	if out == nil {
		writeError(c, errors.New("update settings use case returned nil output"))
		return
	}

	c.JSON(http.StatusOK, mapUpdateSettingsResponse(out.User))
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
