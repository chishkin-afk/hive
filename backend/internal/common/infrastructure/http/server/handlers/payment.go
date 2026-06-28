package handlers

import (
	"context"
	"net/http"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	EventSucceeded = "payment.succeeded"
	EventCanceled  = "payment.canceled"
)

type paymentService interface {
	Create(ctx context.Context, req *requests.CreatePayment) (*responses.UnconfirmedPayment, error)
	Confirm(ctx context.Context, req *requests.ConfirmPayment) (*responses.Payment, error)
	Cancel(ctx context.Context, req *requests.CancelPayment) (*responses.Payment, error)
	ListByUserID(ctx context.Context, page, limit uint32) (*responses.Payments, error)
	GetByID(ctx context.Context, id uuid.UUID) (*responses.Payment, error)
	ListAllRates(ctx context.Context) (*responses.Rates, error)
	GetRateByID(ctx context.Context, id uuid.UUID) (*responses.Rate, error)
	TakeTrial(ctx context.Context) error
}

// @Summary      Create the payment
// @Description  Create the unconfirmed payment in db
// @Tags         payment
// @Accept       json
// @Produce      json
// @Param        request body requests.CreatePayment true "data for creating"
// @Success      201  {object} 	responses.UnconfirmedPayment
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /payments/payment [post]
func (h *HandlerFuncs) CreatePayment() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.CreatePayment
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.paymentService.Create(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusCreated, resp)
	}
}

func (h *HandlerFuncs) YookassaAction() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.ConfirmPaymentYookassa
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		switch req.Event {
		case EventCanceled:
			h.cancelPayment(ctx, &requests.CancelPayment{
				UserID:    req.Object.Metadata.UserID,
				PaymentID: req.Object.PaymentID,
			})
			return
		case EventSucceeded:
			h.confirmPayment(ctx, &requests.ConfirmPayment{
				UserID:    req.Object.Metadata.UserID,
				PaymentID: req.Object.PaymentID,
				RateID:    req.Object.Metadata.RateID,
			})
			return
		}

		ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
			Error: "invalid event",
		})
	}
}

func (h *HandlerFuncs) confirmPayment(ctx *gin.Context, req *requests.ConfirmPayment) {
	resp, err := h.paymentService.Confirm(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
			Error: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *HandlerFuncs) cancelPayment(ctx *gin.Context, req *requests.CancelPayment) {
	resp, err := h.paymentService.Cancel(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
			Error: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// @Summary      Returns the list of payments
// @Description  Returns the list of payments by user's id
// @Tags         payment
// @Accept       json
// @Produce      json
// @Param        page     query int    false "page" default(1)
// @Param        limit    query int    false "limit" default(100)
// @Success      200  {object} 	responses.Payments
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /payments/user [get]
func (h *HandlerFuncs) ListPayments() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		page, limit, err := h.parsePagination(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.paymentService.ListByUserID(ctx.Request.Context(), page, limit)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

// @Summary      Returns the list of payments
// @Description  Returns the list of payments by user's id
// @Tags         payment
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID of payment" format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Success      200  {object} 	responses.Payment
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /payments/payment/{id} [get]
func (h *HandlerFuncs) GetPaymentByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid id of payment",
			})
			return
		}

		resp, err := h.paymentService.GetByID(ctx.Request.Context(), id)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

// @Summary      Returns the list of rates
// @Description  Returns the list of rates
// @Tags         payment
// @Accept       json
// @Produce      json
// @Success      200  {object} 	responses.Rates
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /payments/rates [get]
func (h *HandlerFuncs) ListAllRates() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp, err := h.paymentService.ListAllRates(ctx.Request.Context())
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

// @Summary      Returns the rate
// @Description  Returns the rate by its id
// @Tags         payment
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID of rate" format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Success      200  {object} 	responses.Rate
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /payments/rate/{id} [get]
func (h *HandlerFuncs) GetRateByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid id of rate",
			})
			return
		}

		resp, err := h.paymentService.GetRateByID(ctx.Request.Context(), id)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

// @Summary      Take trial days
// @Description  Add trial days to user by jwt token
// @Tags         payment
// @Accept       json
// @Produce      json
// @Success      204
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /payments/take-trial [post]
func (h *HandlerFuncs) TakeTrial() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := h.paymentService.TakeTrial(ctx.Request.Context()); err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}
