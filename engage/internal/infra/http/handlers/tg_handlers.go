package httphandlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/braunkc/hive-engage/internal/infra/http/dtos"
	"github.com/gin-gonic/gin"
	"github.com/mymmrac/telego"
	telegoerr "github.com/mymmrac/telego/telegoapi"
)

func (h *HTTPHandlers) SendMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.log.Debug("reveived send message request")

		var req dtos.SendMessageRequest
		if err := c.ShouldBindBodyWithJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		_, err := h.tg.SendMessage(c.Request.Context(), req.ChatID, req.ThreadID, req.Text)
		if err != nil {
			c.AbortWithStatusJSON(h.errorMap(err))
			return
		}

		c.Status(http.StatusOK)
	}
}

func (h *HTTPHandlers) TelegramWebhook(webhookChan chan<- telego.Update, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret != "" && c.GetHeader(telego.WebhookSecretTokenHeader) != secret {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		var update telego.Update
		if err := c.ShouldBindBodyWithJSON(&update); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		webhookChan <- update
		c.Status(http.StatusOK)
	}
}

func (h *HTTPHandlers) errorMap(err error) (int, gin.H) {
	if err == nil {
		return http.StatusOK, nil
	}

	if tgErr, ok := errors.AsType[*telegoerr.Error](err); ok {
		return tgErr.ErrorCode, gin.H{"error": tgErr.Description}
	}

	h.log.Error("internal error", slog.Any("err", err))
	return http.StatusInternalServerError, gin.H{"error": "internal error"}
}
