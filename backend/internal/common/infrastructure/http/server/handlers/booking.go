package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type bookingService interface {
	Create(ctx context.Context, req *requests.CreateBooking) (*responses.Booking, error)
	GetByID(ctx context.Context, id uuid.UUID) (*responses.Booking, error)
	ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) (*responses.Bookings, error)
	ListByUserID(ctx context.Context, page, limit uint32) (*responses.Bookings, error)
	GetFreeSlots(ctx context.Context, offerID uuid.UUID, date time.Time) (*responses.Slots, error)
	GetAllDays(ctx context.Context, offerID uuid.UUID, from, to time.Time) (*responses.Days, error)
	Update(ctx context.Context, req *requests.UpdateBooking) (*responses.Booking, error)
}

// @Summary      Creates the booking
// @Description  Creates the booking but before it validates the place of booking
// @Tags         booking
// @Accept       json
// @Produce      json
// @Param        request body requests.CreateBooking true "data for creating"
// @Success      201  {object}  responses.Booking
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /bookings/booking [post]
func (h *HandlerFuncs) CreateBooking() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req *requests.CreateBooking
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.bookingService.Create(ctx.Request.Context(), req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusCreated, resp)
	}
}

// @Summary      Returns the booking
// @Description  Returns the booking by its id
// @Tags         booking
// @Accept       json
// @Produce      json
// @Param id path string true "UUID of booking" format(uuid)
// @Success      200  {object}  responses.Booking
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /bookings/booking/{id} [get]
func (h *HandlerFuncs) GetBookingByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid id",
			})
			return
		}

		resp, err := h.bookingService.GetByID(ctx.Request.Context(), id)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

// @Summary      Returns the list of bookings
// @Description  Returns the list of booking by offer's id or user's id
// @Tags         booking
// @Accept       json
// @Produce      json
// @Param        offer_id query string false "UUID of offer" format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Param        page     query int    false "page" default(1)
// @Param        limit    query int    false "limit" default(100)
// @Success      200  {object}  responses.Bookings
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /bookings/bookings [get]
func (h *HandlerFuncs) ListBookings() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		page, limit, err := h.parsePagination(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		rawOfferID := ctx.Query("offer_id")
		if rawOfferID == "" {
			h.listBookingsByUserID(ctx, page, limit)
			return
		}

		offerID, err := uuid.Parse(rawOfferID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid offer's id",
			})
			return
		}

		h.listBookingsByOfferID(ctx, offerID, uint32(page), uint32(limit))
	}
}

func (h *HandlerFuncs) listBookingsByUserID(ctx *gin.Context, page, limit uint32) {
	resp, err := h.bookingService.ListByUserID(ctx.Request.Context(), page, limit)
	if err != nil {
		ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
			Error: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *HandlerFuncs) listBookingsByOfferID(ctx *gin.Context, offerID uuid.UUID, page, limit uint32) {
	resp, err := h.bookingService.ListByOfferID(ctx.Request.Context(), offerID, page, limit)
	if err != nil {
		ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
			Error: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// @Summary      Updates the fields of booking
// @Description  Updates the fields of booking by its id with check rights by jwt token
// @Tags         booking
// @Accept       json
// @Produce      json
// @Param        request body requests.UpdateBooking true "data for updating"
// @Success      200  {array}  responses.Booking
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /bookings/booking [patch]
func (h *HandlerFuncs) UpdateBooking() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req *requests.UpdateBooking
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.bookingService.Update(ctx.Request.Context(), req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

// @Summary      Returns the list of free slots
// @Description  Returns the list of free slots
// @Tags         booking
// @Accept       json
// @Produce      json
// @Param        offerID path string true "UUID of offer" format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Param        date query string true "date of booking" format(date)
// @Success      200  {object}  responses.Slots
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Router       /bookings/slots/{offerID} [get]
func (h *HandlerFuncs) GetFreeSlots() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawDate, err := time.Parse(time.DateOnly, ctx.Query("date"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid date",
			})
			return
		}

		offerID, err := uuid.Parse(ctx.Param("offerID"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid offer id",
			})
			return
		}

		resp, err := h.bookingService.GetFreeSlots(ctx.Request.Context(), offerID, time.Date(
			rawDate.Year(),
			rawDate.Month(),
			rawDate.Day(),
			0, 0, 0, 0, time.UTC,
		))
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

// @Summary      Returns the list of days
// @Description  Returns the list of days with is_active's field
// @Tags         booking
// @Accept       json
// @Produce      json
// @Param        offerID path string true "UUID of offer" format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Param        from query string true "date of booking" format(date)
// @Param        to query string true "date of booking" format(date)
// @Success      200  {object}  responses.Days
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Router       /bookings/days/{offerID} [get]
func (h *HandlerFuncs) GetAllDays() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawFrom, err := time.Parse(time.DateOnly, ctx.Query("from"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid from",
			})
			return
		}

		rawTo, err := time.Parse(time.DateOnly, ctx.Query("to"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid to",
			})
			return
		}

		offerID, err := uuid.Parse(ctx.Param("offerID"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid offer id",
			})
			return
		}

		resp, err := h.bookingService.GetAllDays(ctx.Request.Context(), offerID, rawFrom, rawTo)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}
