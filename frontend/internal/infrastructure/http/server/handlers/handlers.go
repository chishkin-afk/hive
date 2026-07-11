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
		ctx.HTML(http.StatusOK, "landing", gin.H{
			"title": "Hive",
		})
	}
}
