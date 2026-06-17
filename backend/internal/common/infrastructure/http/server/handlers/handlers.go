package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/devathh/hive/backend/pkg/errs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TODO: раскидать роуты сервисов по файлам

type authService interface {
	Prepare(ctx context.Context, req *requests.PrepareUser) (*responses.User, error)
	Confirm(ctx context.Context, req *requests.ConfirmUser) (*responses.Token, error)
	Login(ctx context.Context, req *requests.Login) (*responses.Token, error)
	Update(ctx context.Context, req *requests.UpdateUser) (*responses.User, error)
	GetSelfUser(ctx context.Context) (*responses.User, error)
}

type offerService interface {
	Create(ctx context.Context, req *requests.CreateOffer) (*responses.Offer, error)
	Update(ctx context.Context, req *requests.UpdateOffer) (*responses.Offer, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*responses.Offer, error)
	ListByAuthorID(ctx context.Context, id uuid.UUID, page, limit uint32) (*responses.Offers, error)
}

type clientService interface {
	Upsert(ctx context.Context, req *requests.UpsertClient) (*responses.Client, error)
	GetByID(ctx context.Context, id uuid.UUID) (*responses.Client, error)
	ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) (*responses.Clients, error)
	ListByUserID(ctx context.Context, page, limit uint32) (*responses.Clients, error)
}

type scheduleService interface {
	Create(ctx context.Context, req *requests.CreateSchedule) (*responses.Schedule, error)
	GetSchedulesByUserID(ctx context.Context, userID uuid.UUID) (*responses.Schedules, error)
	GetScheduleByUserID(ctx context.Context, userID uuid.UUID, dayOfWeek uint8) (*responses.Schedule, error)
	Update(ctx context.Context, req *requests.UpdateSchedule) (*responses.Schedule, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type bookingService interface {
	Create(ctx context.Context, req *requests.CreateBooking) (*responses.Booking, error)
	GetByID(ctx context.Context, id uuid.UUID) (*responses.Booking, error)
	ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) (*responses.Bookings, error)
	ListByUserID(ctx context.Context, page, limit uint32) (*responses.Bookings, error)
	Update(ctx context.Context, req *requests.UpdateBooking) (*responses.Booking, error)
}

type HandlerFuncs struct {
	authService     authService
	offerService    offerService
	clientService   clientService
	scheduleService scheduleService
	bookingService  bookingService
}

func (h *HandlerFuncs) Prepare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.PrepareUser
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid request",
			})
			return
		}

		resp, err := h.authService.Prepare(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusCreated, resp)
	}
}

func (h *HandlerFuncs) Confirm() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.ConfirmUser
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid request",
			})
			return
		}

		resp, err := h.authService.Confirm(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *HandlerFuncs) Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.Login
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid request",
			})
			return
		}

		resp, err := h.authService.Login(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *HandlerFuncs) UpdateUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.UpdateUser
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.authService.Update(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *HandlerFuncs) GetSelfUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp, err := h.authService.GetSelfUser(ctx.Request.Context())
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *HandlerFuncs) CreateOffer() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.CreateOffer
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.offerService.Create(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusCreated, resp)
	}
}

func (h *HandlerFuncs) UpdateOffer() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.UpdateOffer
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.offerService.Update(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *HandlerFuncs) DeleteOffer() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid id of offer",
			})
			return
		}

		if err := h.offerService.Delete(ctx.Request.Context(), id); err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}

func (h *HandlerFuncs) GetOfferByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid id of offer",
			})
			return
		}

		resp, err := h.offerService.GetByID(ctx.Request.Context(), id)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *HandlerFuncs) ListOffersByAuthorID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorID, err := uuid.Parse(ctx.Param("authorID"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid author's id",
			})
			return
		}

		page, err := strconv.Atoi(ctx.Query("page"))
		if err != nil || page < 1 {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid page",
			})
			return
		}

		limit, err := strconv.Atoi(ctx.Query("limit"))
		if err != nil || limit < 1 {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid limit",
			})
			return
		}

		resp, err := h.offerService.ListByAuthorID(ctx, authorID, uint32(page), uint32(limit))
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *HandlerFuncs) UpsertClient() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.UpsertClient
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.clientService.Upsert(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *HandlerFuncs) GetClientByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid id",
			})
			return
		}

		resp, err := h.clientService.GetByID(ctx.Request.Context(), id)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *HandlerFuncs) ListClientsByOfferID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		offerID, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid offer's id",
			})
			return
		}

		page, err := strconv.Atoi(ctx.Query("page"))
		if err != nil || page < 1 {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid page",
			})
			return
		}

		limit, err := strconv.Atoi(ctx.Query("limit"))
		if err != nil || limit < 1 {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid limit",
			})
			return
		}

		resp, err := h.clientService.ListByOfferID(ctx.Request.Context(), offerID, uint32(page), uint32(limit))
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *HandlerFuncs) ListClientsByUserID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		page, err := strconv.Atoi(ctx.Query("page"))
		if err != nil || page < 1 {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid page",
			})
			return
		}

		limit, err := strconv.Atoi(ctx.Query("limit"))
		if err != nil || limit < 1 {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid limit",
			})
			return
		}

		resp, err := h.clientService.ListByUserID(ctx.Request.Context(), uint32(page), uint32(limit))
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
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
			h.listBookingsByUserID(ctx, uint32(page), uint32(limit))
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

func (h *HandlerFuncs) parsePagination(ctx *gin.Context) (uint32, uint32, error) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 || page > 1000 {
		return 0, 0, errs.ErrInvalidPage
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		return 0, 0, errs.ErrInvalidLimit
	}

	return uint32(page), uint32(limit), nil
}

func (h *HandlerFuncs) getCodeFromError(err error) int {
	var errKind *errs.ErrorKind
	if !errors.As(err, &errKind) {
		return http.StatusInternalServerError
	}

	switch errKind.Kind() {
	case errs.KindRequest:
		return http.StatusBadRequest
	case errs.KindInternal:
		return http.StatusInternalServerError
	case errs.KindUnauthorized:
		return http.StatusUnauthorized
	case errs.KindConflict:
		return http.StatusConflict
	case errs.KindTimeout:
		return http.StatusRequestTimeout
	case errs.KindNotFound:
		return http.StatusNotFound
	case errs.KindPermissionDenied:
		return http.StatusForbidden
	}

	return http.StatusInternalServerError
}
