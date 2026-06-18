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
	Update(ctx context.Context, req *requests.UpdateBooking) (*responses.Booking, error)
}

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
