package handlers

import (
	"context"
	"net/http"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type paymentService interface {
	Create(ctx context.Context, req *requests.CreatePayment) (*responses.Payment, error)
	Confirm(ctx context.Context, req *requests.ConfirmPayment) (*responses.Payment, error)
	Cancel(ctx context.Context, req *requests.CancelPayment) (*responses.Payment, error)
	ListByUserID(ctx context.Context, page, limit uint32) (*responses.Payments, error)
	GetByID(ctx context.Context, id uuid.UUID) (*responses.Payment, error)
	ListAllRates(ctx context.Context) (*responses.Rates, error)
	GetRateByID(ctx context.Context, id uuid.UUID) (*responses.Rate, error)
}

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
		panic("not impl")
	}
}

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

func (h *HandlerFuncs) GetPaymentByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse("id")
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

func (h *HandlerFuncs) GetRateByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := uuid.Parse("id")
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
