package handlers

import (
	"context"
	"net/http"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type offerService interface {
	Create(ctx context.Context, req *requests.CreateOffer) (*responses.Offer, error)
	Update(ctx context.Context, req *requests.UpdateOffer) (*responses.Offer, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*responses.Offer, error)
	ListByAuthorID(ctx context.Context, id uuid.UUID, page, limit uint32) (*responses.Offers, error)
}

// @Summary      Create the offer
// @Description  Create the offer
// @Tags         offer
// @Accept       json
// @Produce      json
// @Param        request body requests.CreateOffer true "data for creating"
// @Success      201  {object}  	responses.Offer
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /offers/offer [post]
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

// @Summary      Update the offer
// @Description  Update the fields of offer by its id
// @Tags         offer
// @Accept       json
// @Produce      json
// @Param        request body requests.UpdateOffer true "data for updating"
// @Success      200  {object}  	responses.Offer
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /offers/offer [patch]
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

// @Summary      Delete the offer
// @Description  Delete the offer by its id
// @Tags         offer
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID of offer" format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Success      200  {object}  	responses.Offer
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /offers/offer/{id} [delete]
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

// @Summary      Get the offer
// @Description  Get the offer by its id
// @Tags         offer
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID of offer" format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Success      200  {object}  	responses.Offer
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /offers/offer/{id} [get]
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

// @Summary      Get list of the offers
// @Description  Get list of the offers by author's id
// @Tags         offer
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID of offer" format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Param        page     query int    false "page" default(1)
// @Param        limit    query int    false "limit" default(100)
// @Success      200  {object}  	responses.Offers
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /offers/user/{authorID} [get]
func (h *HandlerFuncs) ListOffersByAuthorID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorID, err := uuid.Parse(ctx.Param("authorID"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: "invalid author's id",
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

		resp, err := h.offerService.ListByAuthorID(ctx, authorID, page, limit)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}
