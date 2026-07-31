package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type frontendService interface {
	Config() []byte
}

type handlers struct {
	fs frontendService
}

func (h *handlers) Config() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		js := h.fs.Config()
		ctx.Data(http.StatusOK, "application/javascript; charset=utf-8", []byte(js))
	}
}

func (h *handlers) Landing() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title": "Hive",
			"page":  "landing",
		})
	}
}

func (h *handlers) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		action := ctx.DefaultQuery("action", "register")

		ctx.HTML(http.StatusOK, "base", gin.H{
			"title":  "Auth",
			"page":   "auth",
			"action": action,
		})
	}
}

func (h *handlers) Appointment() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title":              "название услуги",
			"serviceTitle":       "название услуги",
			"serviceDescription": "описание услуги",
			"serviceAmount":      "стоимость услуги",
			"page":               "appointment",
		})
	}
}

func (h *handlers) Services() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title": "Услуги",
			"page":  "services",
		})
	}
}

func (h *handlers) Profile() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title": "Профиль",
			"page":  "profile",
		})
	}
}

func (h *handlers) Topup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title": "Тарифы",
			"page":  "topup",
		})
	}
}

func (h *handlers) Bookings() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title": "Записи",
			"page":  "bookings",
		})
	}
}
