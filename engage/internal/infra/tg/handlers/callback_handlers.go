package tghandlers

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func (tgh *TgHandlers) setupCallbackQueryHandlers(bh *th.BotHandler) {
	bh.Handle(tgh.callbackQueryMenuHandler, th.CallbackDataPrefix("menu_"))
	bh.Handle(tgh.callbackQueryTicketHandler, th.CallbackDataPrefix("ticket_"))
	bh.Handle(tgh.callbackQueryTopicHandler, th.CallbackDataPrefix("topic_"))
}

func (tgh *TgHandlers) callbackQueryMenuHandler(ctx *th.Context, upd telego.Update) error {
	switch upd.CallbackQuery.Data {
	case callbackMenuMain:
		return tgh.handleError(ctx, upd, tgh.changeMenuToMain(ctx.Context(), upd))
	case callbackMenuSupport:
		return tgh.handleError(ctx, upd, tgh.changeMenuToSupport(ctx.Context(), upd))
	default:
		tgh.log.Warn("invalid callback data", slog.String("data", upd.CallbackQuery.Data))
		return nil
	}
}

func (tgh *TgHandlers) callbackQueryTicketHandler(ctx *th.Context, upd telego.Update) error {
	data := upd.CallbackQuery.Data

	switch {
	case data == callbackTicketCreate:
		return tgh.handleError(ctx.Context(), upd, tgh.handleTicketCreate(ctx.Context(), upd))
	case data == callbackTicketConfirm:
		return tgh.handleError(ctx.Context(), upd, tgh.handleTicketConfirm(ctx.Context(), upd))
	case data == callbackTicketCancel:
		return tgh.handleError(ctx.Context(), upd, tgh.handleTicketCancel(ctx.Context(), upd))
	case strings.HasPrefix(data, callbackTicketAssign):
		ticketID, err := strconv.ParseInt(strings.TrimPrefix(data, fmt.Sprintf("%s_", callbackTicketAssign)), 10, 64)
		if err != nil {
			return err
		}
		return tgh.handleError(ctx.Context(), upd, tgh.handleTicketAssign(ctx.Context(), upd, ticketID))
	case strings.HasPrefix(data, callbackTicketClose):
		ticketID, err := strconv.ParseInt(strings.TrimPrefix(data, fmt.Sprintf("%s_", callbackTicketClose)), 10, 64)
		if err != nil {
			return err
		}
		return tgh.handleError(ctx.Context(), upd, tgh.handleTicketClose(ctx.Context(), upd, ticketID))
	default:
		tgh.log.Warn("invalid callback data", slog.String("data", upd.CallbackQuery.Data))
		return nil
	}
}

func (tgh *TgHandlers) callbackQueryTopicHandler(ctx *th.Context, upd telego.Update) error {
	switch upd.CallbackQuery.Data {
	case callbackTopicDelete:
		return tgh.handleError(ctx.Context(), upd, tgh.handleTopicDelete(ctx.Context(), upd))
	default:
		tgh.log.Warn("invalid callback data", slog.String("data", upd.CallbackQuery.Data))
		return nil
	}
}
