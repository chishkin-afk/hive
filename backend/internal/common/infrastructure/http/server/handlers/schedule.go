package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type scheduleService interface {
	Create(ctx context.Context, req *requests.CreateSchedule) (*responses.Schedule, error)
	GetSchedulesByUserID(ctx context.Context, userID uuid.UUID) (*responses.Schedules, error)
	GetScheduleByUserID(ctx context.Context, userID uuid.UUID, dayOfWeek uint8) (*responses.Schedule, error)
	Update(ctx context.Context, req *requests.UpdateSchedule) (*responses.Schedule, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

func (h *HandlerFuncs) CreateSchedule() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.CreateSchedule
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.scheduleService.Create(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusCreated, resp)
	}
}

func (h *HandlerFuncs) UpdateSchedule() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.UpdateSchedule
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.scheduleService.Update(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *HandlerFuncs) GetScheduleByUserID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, err := uuid.Parse(ctx.Param("userID"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid user's id",
			})
			return
		}

		dayOfWeek, err := strconv.Atoi(ctx.Param("dayOfWeek"))
		if err != nil || dayOfWeek < 0 || dayOfWeek > 6 {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid day of week",
			})
		}

		resp, err := h.scheduleService.GetScheduleByUserID(ctx.Request.Context(), userID, uint8(dayOfWeek))
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *HandlerFuncs) GetSchedulesByUserID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, err := uuid.Parse(ctx.Param("userID"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid id",
			})
			return
		}

		resp, err := h.scheduleService.GetSchedulesByUserID(ctx.Request.Context(), userID)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *HandlerFuncs) DeleteSchedule() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid id",
			})
			return
		}

		if err := h.scheduleService.Delete(ctx.Request.Context(), id); err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}
