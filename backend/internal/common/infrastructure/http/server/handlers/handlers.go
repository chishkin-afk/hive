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

type HandlerFuncs struct {
	authService   authService
	offerService  offerService
	clientService clientService
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
