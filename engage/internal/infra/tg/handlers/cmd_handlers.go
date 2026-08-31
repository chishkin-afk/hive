package tghandlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/braunkc/hive-engage/internal/infra/http/dtos"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func (tgh *TgHandlers) setupCMDHandlers(bh *th.BotHandler) {
	bh.Handle(tgh.startHandler, th.CommandEqual("start"))
}

func (tgh *TgHandlers) startHandler(ctx *th.Context, upd telego.Update) error {
	args := tgh.getArgs(upd)
	if len(args) > 0 {
		return tgh.handleError(ctx, upd, tgh.prepareStartWithQuery(ctx.Context(), upd, args[0]))
	}

	return tgh.handleError(ctx, upd, tgh.prepareStart(ctx.Context(), upd))
}

func (tgh *TgHandlers) prepareStart(ctx context.Context, upd telego.Update) error {
	_, err := tgh.sendMenu(ctx, upd)
	if err != nil {
		return err
	}

	return nil
}

func (tgh *TgHandlers) prepareStartWithQuery(ctx context.Context, upd telego.Update, arg string) error {
	user, err := tgh.tg.GetUser(ctx, upd.Message.From.ID)
	if err != nil {
		return err
	}

	if err = tgh.client.BindTg(ctx, &dtos.BindTgRequest{
		BindingID: arg,
		TgChatID:  upd.Message.From.ID,
	}); err != nil {
		return err
	}

	text := fmt.Sprintf("%s, успешная авторизация", user.FirstName)
	if _, err = tgh.tg.SendTempMessage(ctx, upd.Message.Chat.ID, 0, text, time.Minute); err != nil {
		return err
	}

	_, err = tgh.sendMenu(ctx, upd)
	if err != nil {
		return err
	}

	return nil
}

func (tgh *TgHandlers) getArgs(upd telego.Update) []string {
	if upd.Message == nil {
		return []string{}
	}

	text := strings.TrimSpace(upd.Message.Text)
	if text == "" {
		return []string{}
	}

	parts := strings.Fields(text)

	return parts[1:]
}
