package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/devathh/hive/backend/internal/modules/auth/domain/user"
	"github.com/devathh/hive/backend/pkg/errs"
	"github.com/gin-gonic/gin"
)

type authService interface {
	Prepare(ctx context.Context, req *requests.PrepareUser) (*responses.User, error)
	Confirm(ctx context.Context, req *requests.ConfirmUser) (*responses.Token, error)
	Login(ctx context.Context, req *requests.Login) (*responses.Token, error)
	Update(ctx context.Context, req *requests.UpdateUser) (*responses.User, error)
	GetSelfUser(ctx context.Context) (*responses.User, error)
}

type HandlerFuncs struct {
	authService authService
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

func (h *HandlerFuncs) getCodeFromError(err error) int {
	switch {
	case errors.Is(err, user.ErrAlreadyOffRest),
		errors.Is(err, user.ErrAlreadyOnRest),
		errors.Is(err, user.ErrNotificationTgDisabled),
		errors.Is(err, errs.ErrInvalidCode),
		errors.Is(err, user.ErrInvalidEmail),
		errors.Is(err, user.ErrInvalidFirstname),
		errors.Is(err, user.ErrInvalidLastname),
		errors.Is(err, user.ErrInvalidTgChatID),
		errors.Is(err, user.ErrAlreadyOffRest),
		errors.Is(err, user.ErrAlreadyOnRest),
		errors.Is(err, user.ErrNotificationTgDisabled),
		errors.Is(err, user.ErrInvalidPassword):
		return http.StatusBadRequest
	case errors.Is(err, user.ErrUserAlreadyExists),
		errors.Is(err, errs.ErrUserAlreadyRegistered):
		return http.StatusConflict
	case errors.Is(err, user.ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, errs.ErrInvalidToken),
		errors.Is(err, errs.ErrInvalidCredentials):
		return http.StatusUnauthorized
	}

	return http.StatusInternalServerError
}
