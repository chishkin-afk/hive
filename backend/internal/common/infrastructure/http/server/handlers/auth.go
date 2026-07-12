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
	ResendCode(ctx context.Context, req *requests.ResendCode) (*responses.CodeResent, error)
	Confirm(ctx context.Context, req *requests.ConfirmUser) (*responses.Token, error)
	Login(ctx context.Context, req *requests.Login) (*responses.Token, error)
	Update(ctx context.Context, req *requests.UpdateUser) (*responses.User, error)
	GetSelfUser(ctx context.Context) (*responses.User, error)
	GetTgBindingURL(ctx context.Context) (*responses.TgBinding, error)
	BindTG(ctx context.Context, req *requests.BindTG) (*responses.User, error)
	Lookup(ctx context.Context, req *requests.Lookup) (*responses.LookupUsers, error)
	ForgetPassword(ctx context.Context, req *requests.ForgetPassword) error
	ChangePassword(ctx context.Context, req *requests.ChangePassword) (*responses.Token, error)
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

// @Summary      Get the self tg's binding url by jwt token
// @Description  Returns the tg's binding url to bot
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  responses.TgBinding
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /auth/tgbinding [get]
func (h *HandlerFuncs) GetTgBindingURL() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp, err := h.authService.GetTgBindingURL(ctx.Request.Context())
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

// @Summary      Update the telegram's fields of user
// @Description  Update the telegram's fields of user
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
// @Router       /auth/tgbinding [post]
func (h *HandlerFuncs) BindTG() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.BindTG
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.authService.BindTG(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

// @Summary      Resend the auth code
// @Description  Resend the auth code
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      202  {object}  responses.CodeResent
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /auth/resend [post]
func (h *HandlerFuncs) ResendCode() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.ResendCode
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.authService.ResendCode(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusAccepted, resp)
	}
}

func (h *HandlerFuncs) Lookup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.Lookup
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.authService.Lookup(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}

// @Summary      Send the change password url
// @Description  Send the change password url
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      202  {null}
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /auth/forget-password [post]
func (h *HandlerFuncs) ForgetPassword() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.ForgetPassword
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		if err := h.authService.ForgetPassword(ctx.Request.Context(), &req); err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.Status(http.StatusAccepted)
	}
}

// @Summary      Change the password's field
// @Description  Change the password's field by reset token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  responses.Token
// @Failure      400  {object}  responses.ErrMsg
// @Failure      401  {object}  responses.ErrMsg
// @Failure      403  {object}  responses.ErrMsg
// @Failure      404  {object}  responses.ErrMsg
// @Failure      500  {object}  responses.ErrMsg
// @Security     BearerAuth
// @Router       /auth/password [patch]
func (h *HandlerFuncs) ChangePassword() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req requests.ChangePassword
		if err := ctx.BindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		resp, err := h.authService.ChangePassword(ctx.Request.Context(), &req)
		if err != nil {
			ctx.JSON(h.getCodeFromError(err), &responses.ErrMsg{
				Error: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resp)
	}
}
