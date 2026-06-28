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

// @Summary      Create the schedule
// @Description  Create the schedule of the user by jwt token
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        request body requests.CreateSchedule true "data for creating"
// @Success      201  {object} 	responses.Schedule
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /schedules/schedule [post]
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

// @Summary      Update the schedule
// @Description  Update the fields of schedule by its token
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        request body requests.UpdateSchedule true "data for updating"
// @Success      200  {object} 	responses.Schedule
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /schedules/schedule [patch]
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

// @Summary      Get the schedule
// @Description  Get the schedule by user's id and day of week
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        userID path string true "UUID of user" format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Param        dayOfWeek path string true "day of week"
// @Success      200  {object} 	responses.Schedule
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /schedules/schedule/{userID}/d/{dayOfWeek} [get]
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

// @Summary      Get the schedules
// @Description  Get the schedules by user's id
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        userID path string true "UUID of user" format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Success      200  {object} 	responses.Schedule
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /schedules/user/{userID} [get]
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

// @Summary      Get the schedule
// @Description  Get the schedule by user's id and day of week
// @Tags         schedule
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID of schedule" format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Success      200  {object} 	responses.Schedule
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /schedules/schedule/{id} [delete]
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
