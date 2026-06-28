package handlers

import (
	"context"
	"net/http"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/gin-gonic/gin"
)

type authService interface {
	Prepare(ctx context.Context, req *requests.PrepareUser) (*responses.User, error)
	Confirm(ctx context.Context, req *requests.ConfirmUser) (*responses.Token, error)
	Login(ctx context.Context, req *requests.Login) (*responses.Token, error)
	Update(ctx context.Context, req *requests.UpdateUser) (*responses.User, error)
	UpdateTGChatID(ctx context.Context, req *requests.UpdateTgChatID) (*responses.User, error)
	GetSelfUser(ctx context.Context) (*responses.User, error)
}

// @Summary      Prepare the user
// @Description  Prepares the user in the cache and waits for the confirmation code from the email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body requests.PrepareUser true "data for preparing"
// @Success      201  {object}  responses.User
// @Failure      400  {object}  responses.ErrMsg
// @Failure      409  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Router       /auth/prepare [post]
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

// @Summary      Confirm the prepared user
// @Description  Confirms the prepared user and saves him into db, also generates the jwt token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body requests.ConfirmUser true "data for confirming"
// @Success      200  {object}  responses.Token
// @Failure      400  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Router       /auth/confirm [post]
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

// @Summary      Login
// @Description  Checks passwords of user and generates the jwt token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body requests.Login true "data for logining"
// @Success      200  {object}  responses.Token
// @Failure      400  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Router       /auth/login [post]
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

// @Summary      Update the fields of user
// @Description  Applies the changes to user by his id from jwt token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body requests.UpdateUser true "data for updating"
// @Success      200  {object}  responses.User
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /auth/user [patch]
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

// @Summary      Get the self user by jwt token
// @Description  Returns the user from cache or db by jwt token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  responses.User
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /auth/user [get]
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

// @Summary      Update the field of user's telegram chat id
// @Description  Applies the change to user by his id from jwt token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body requests.UpdateTgChatID true "data for updating"
// @Success      200  {object}  responses.User
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      408  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /auth/tg [patch]
func (h *HandlerFuncs) UpdateTGChatID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.UpdateTgChatID
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.authService.UpdateTGChatID(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}
