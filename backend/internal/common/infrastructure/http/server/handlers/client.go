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

// @Summary      Upsert the client
// @Description Create or update the client by user's id
// @Tags         client
// @Accept       json
// @Produce      json
// @Param        request body requests.UpsertClient true "data for upserting"
// @Success      200  {object}  responses.Client
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /clients/client [post]
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

// @Summary      Returns the client by his id
// @Description  Returns the client by his id
// @Tags         client
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID of client" format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Success      200  {object}   responses.Client
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /clients/client/{id} [get]
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

// @Summary      Returns a list of clients
// @Description  Returns a list of clients by offer's id
// @Tags         client
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID of offer" format(uuid) example("550e8400-e29b-41d4-a716-446655440000")
// @Param        page     query int    false "page" default(1)
// @Param        limit    query int    false "limit" default(100)
// @Success      200  {object}   responses.Clients
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /clients/offer/{id} [get]
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

// @Summary      Returns a list of clients
// @Description  Returns a list of clients by jwt token of user
// @Tags         client
// @Accept       json
// @Produce      json
// @Param        page     query int    false "page" default(1)
// @Param        limit    query int    false "limit" default(100)
// @Success      200  {object}   responses.Clients
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /clients/user [get]
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
