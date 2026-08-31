package httphandlers

import (
	"log/slog"

	engagedomain "github.com/braunkc/hive-engage/internal/domain/engage"
	th "github.com/mymmrac/telego/telegohandler"
)

type HTTPHandlers struct {
	tg  engagedomain.EngageAPI
	bh  *th.BotHandler
	log *slog.Logger
}

func New(tg engagedomain.EngageAPI, bh *th.BotHandler, log *slog.Logger) *HTTPHandlers {
	return &HTTPHandlers{
		tg:  tg,
		bh:  bh,
		log: log,
	}
}
