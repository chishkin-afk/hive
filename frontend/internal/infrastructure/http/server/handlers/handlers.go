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
			"title":  "Hive",
			"page":   "landing",
			"og_url": ctx.Request.URL.Path,
		})
	}
}

func (h *handlers) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		action := ctx.DefaultQuery("action", "register")

		ctx.HTML(http.StatusOK, "base", gin.H{
			"title":  "Авторизация",
			"page":   "auth",
			"action": action,
			"og_url": ctx.Request.URL.Path,
		})
	}
}

func (h *handlers) Appointment() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title":  "Запись",
			"page":   "appointment",
			"og_url": ctx.Request.URL.Path,
		})
	}
}

func (h *handlers) Services() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title":  "Услуги",
			"page":   "services",
			"og_url": ctx.Request.URL.Path,
		})
	}
}

func (h *handlers) Profile() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title":  "Профиль",
			"page":   "profile",
			"og_url": ctx.Request.URL.Path,
		})
	}
}

func (h *handlers) Topup() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title":  "Тарифы",
			"page":   "topup",
			"og_url": ctx.Request.URL.Path,
		})
	}
}

func (h *handlers) Bookings() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title":  "Записи",
			"page":   "bookings",
			"og_url": ctx.Request.URL.Path,
		})
	}
}

func (h *handlers) Reset() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Query("token")
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title":       "Смена пароля",
			"page":        "reset",
			"reset_token": token,
			"og_url":      ctx.Request.URL.Path,
		})
	}
}

func (h *handlers) Terms() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title":  "Пользовательское соглашение",
			"page":   "terms",
			"og_url": ctx.Request.URL.Path,
		})
	}
}

func (h *handlers) Privacy() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title":  "Политика обработки данных",
			"page":   "privacy",
			"og_url": ctx.Request.URL.Path,
		})
	}
}

func (h *handlers) Offer() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title":  "Оферта",
			"page":   "offer",
			"og_url": ctx.Request.URL.Path,
		})
	}
}

func (h *handlers) Consent() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base", gin.H{
			"title":  "Согласие на обработку персональных данных",
			"page":   "consent",
			"og_url": ctx.Request.URL.Path,
		})
	}
}
