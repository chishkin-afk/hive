package handlers

import (
	"context"
	"net/http"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type clientService interface {
	Upsert(ctx context.Context, req *requests.UpsertClient) (*responses.Client, error)
	GetByID(ctx context.Context, id uuid.UUID) (*responses.Client, error)
	ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) (*responses.Clients, error)
	ListByUserID(ctx context.Context, page, limit uint32) (*responses.Clients, error)
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

		page, limit, err := h.parsePagination(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.clientService.ListByOfferID(ctx.Request.Context(), offerID, page, limit)
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
		page, limit, err := h.parsePagination(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.clientService.ListByUserID(ctx.Request.Context(), page, limit)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}
