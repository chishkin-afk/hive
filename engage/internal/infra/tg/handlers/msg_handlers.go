package tghandlers

import (
	"context"
	"errors"
	"time"

	ticketapplication "github.com/braunkc/hive-engage/internal/application/ticket"
	ticketdomain "github.com/braunkc/hive-engage/internal/domain/ticket"
	statecache "github.com/braunkc/hive-engage/internal/infra/repository/cache/state"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func (tgh *TgHandlers) setupMsgHandlers(bh *th.BotHandler) {
	bh.Handle(tgh.msgHandler, th.AnyMessage())
}

func (tgh *TgHandlers) msgHandler(ctx *th.Context, upd telego.Update) error {
	switch upd.Message.Chat.Type {
	case telego.ChatTypePrivate:
		return tgh.handleError(ctx.Context(), upd, tgh.handlePrivateMsg(ctx, upd))
	default:
		return tgh.handleError(ctx.Context(), upd, tgh.handleGroupMsg(ctx, upd))
	}
}

func (tgh *TgHandlers) handlePrivateMsg(ctx *th.Context, upd telego.Update) error {
	if upd.Message.Text != "" {
		state, err := tgh.stateRepo.Get(ctx.Context(), upd.Message.From.ID)
		if err == nil {
			return tgh.handleError(ctx.Context(), upd, tgh.prepareState(ctx.Context(), upd, state))
		}
		if !errors.Is(err, statecache.ErrRecordNotFound) {
			return tgh.handleError(ctx, upd, err)
		}
	}

	ticket, err := tgh.ticketService.GetActiveByClientID(ctx.Context(), ticketapplication.GetRequest{ID: upd.Message.From.ID})
	if err == nil {
		return tgh.handleError(ctx, upd, tgh.handleEcho(ctx.Context(), upd, ticket))
	}
	if !errors.Is(err, ticketdomain.ErrTicketNotFound) {
		return tgh.handleError(ctx, upd, err)
	}

	return tgh.handleError(ctx.Context(), upd, tgh.sendHelp(ctx.Context(), upd))
}

func (tgh *TgHandlers) handleGroupMsg(ctx *th.Context, upd telego.Update) error {
	if upd.Message.MessageThreadID != 0 && upd.Message.ForumTopicClosed == nil {
		ticket, err := tgh.ticketService.GetActiveByThreadID(ctx.Context(), ticketapplication.GetByThreadIDRequest{ID: upd.Message.MessageThreadID})
		if err == nil {
			return tgh.handleError(ctx, upd, tgh.handleEcho(ctx.Context(), upd, ticket))
		}
		if !errors.Is(err, ticketdomain.ErrTicketNotFound) {
			return tgh.handleError(ctx, upd, err)
		}
	}

	return nil
}

func (tgh *TgHandlers) sendHelp(ctx context.Context, upd telego.Update) error {
	if _, err := tgh.tg.SendTempMessage(ctx, upd.Message.Chat.ID, upd.Message.MessageThreadID, "Чтобы посмотреть меню, нажмите /start", 10*time.Second); err != nil {
		return err
	}

	return nil
}
