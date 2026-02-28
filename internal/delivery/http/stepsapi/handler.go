package stepsapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	appsteps "tracker/internal/application/steps"
	"tracker/internal/delivery/http/middleware"
	"tracker/internal/domain/shared"
)

type ListUseCase interface {
	Execute(ctx context.Context, in appsteps.ListInput) (*appsteps.ListOutput, error)
}

type PutUseCase interface {
	Execute(ctx context.Context, in appsteps.PutInput) (*appsteps.PutOutput, error)
}

type AddUseCase interface {
	Execute(ctx context.Context, in appsteps.AddInput) (*appsteps.AddOutput, error)
}

type DeleteUseCase interface {
	Execute(ctx context.Context, in appsteps.DeleteInput) error
}

type AnalyticsUseCase interface {
	Execute(ctx context.Context, in appsteps.AnalyticsInput) (*appsteps.AnalyticsOutput, error)
}

type StepsUseCases struct {
	List      ListUseCase
	Put       PutUseCase
	Add       AddUseCase
	Delete    DeleteUseCase
	Analytics AnalyticsUseCase
}

type Handler struct {
	useCases StepsUseCases
}

func NewHandler(useCases StepsUseCases) *Handler {
	return &Handler{useCases: useCases}
}

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {
	api.GET("/steps", handler.List)
	api.PUT("/steps/:date", handler.Put)
	api.POST("/steps/add", handler.Add)
	api.DELETE("/steps/:date", handler.Delete)
	api.GET("/steps/analytics", handler.Analytics)
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, shared.ErrUnauthorized)
		return
	}

	from, err := shared.ParseLocalDate(strings.TrimSpace(c.Query("from")))
	if err != nil {
		writeError(c, errors.Join(shared.ErrInvalidInput, err))
		return
	}
	to, err := shared.ParseLocalDate(strings.TrimSpace(c.Query("to")))
	if err != nil {
		writeError(c, errors.Join(shared.ErrInvalidInput, err))
		return
	}

	out, err := h.useCases.List.Execute(c.Request.Context(), appsteps.ListInput{
		UserID: userID,
		From:   from,
		To:     to,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	if out == nil {
		writeError(c, errors.New("list use case returned nil output"))
		return
	}

	c.JSON(http.StatusOK, mapListResponse(*out))
}

func (h *Handler) Put(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, shared.ErrUnauthorized)
		return
	}

	date, err := shared.ParseLocalDate(strings.TrimSpace(c.Param("date")))
	if err != nil {
		writeError(c, errors.Join(shared.ErrInvalidInput, err))
		return
	}

	var body putRequestATO
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, errors.Join(shared.ErrInvalidInput, err))
		return
	}

	out, err := h.useCases.Put.Execute(c.Request.Context(), appsteps.PutInput{
		UserID: userID,
		Date:   date,
		Count:  body.Count,
		Source: body.Source,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	if out == nil {
		writeError(c, errors.New("put use case returned nil output"))
		return
	}

	c.JSON(http.StatusOK, mapPutResponse(*out))
}

func (h *Handler) Add(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, shared.ErrUnauthorized)
		return
	}

	var body addRequestATO
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, errors.Join(shared.ErrInvalidInput, err))
		return
	}

	date, err := shared.ParseLocalDate(strings.TrimSpace(body.Date))
	if err != nil {
		writeError(c, errors.Join(shared.ErrInvalidInput, err))
		return
	}

	out, err := h.useCases.Add.Execute(c.Request.Context(), appsteps.AddInput{
		UserID: userID,
		Date:   date,
		Delta:  body.Delta,
		Source: body.Source,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	if out == nil {
		writeError(c, errors.New("add use case returned nil output"))
		return
	}

	c.JSON(http.StatusOK, mapAddResponse(*out))
}

func (h *Handler) Delete(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, shared.ErrUnauthorized)
		return
	}

	date, err := shared.ParseLocalDate(strings.TrimSpace(c.Param("date")))
	if err != nil {
		writeError(c, errors.Join(shared.ErrInvalidInput, err))
		return
	}

	err = h.useCases.Delete.Execute(c.Request.Context(), appsteps.DeleteInput{
		UserID: userID,
		Date:   date,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) Analytics(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		writeError(c, shared.ErrUnauthorized)
		return
	}

	month, err := shared.ParseLocalMonth(strings.TrimSpace(c.Query("month")))
	if err != nil {
		writeError(c, errors.Join(shared.ErrInvalidInput, err))
		return
	}

	out, err := h.useCases.Analytics.Execute(c.Request.Context(), appsteps.AnalyticsInput{
		UserID: userID,
		Month:  month,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	if out == nil {
		writeError(c, errors.New("analytics use case returned nil output"))
		return
	}

	c.JSON(http.StatusOK, mapAnalyticsResponse(*out))
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
