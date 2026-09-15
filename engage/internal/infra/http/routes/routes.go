package httproutes

import (
	"github.com/braunkc/hive-engage/config"
	httphandlers "github.com/braunkc/hive-engage/internal/infra/http/handlers"
	"github.com/braunkc/hive-engage/internal/infra/http/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/mymmrac/telego"
)

func Setup(r *gin.Engine, h *httphandlers.HTTPHandlers, webhookChan chan<- telego.Update, cfg config.TgWebhookConfig) {
	api := r.Group("/api")
	api.Use(middlewares.InternalOnly())
	{
		v1 := api.Group("/v1")
		{
			v1.POST("/telegram/message/send", h.SendMessage())
		}
	}

	if webhookChan != nil {
		r.POST("/telegram/webhook", h.TelegramWebhook(webhookChan, cfg.Secret))
	}
}
