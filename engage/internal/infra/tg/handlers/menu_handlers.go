package tghandlers

import (
	"context"

	engagedomain "github.com/braunkc/hive-engage/internal/domain/engage"
	"github.com/mymmrac/telego"
)

const (
	callbackMenuMain     = "menu_main"
	callbackMenuSupport  = "menu_support"
	callbackTicketCreate = "ticket_create"
)

func (tgh *TgHandlers) mainMenu() (string, [][]engagedomain.InlineKeyboardButton) {
	text := "Главное меню"
	btns := [][]engagedomain.InlineKeyboardButton{
		{
			{
				Text:         "🛟 Поддержка",
				CallbackData: callbackMenuSupport,
			},
		},
	}

	return text, btns
}

func (tgh *TgHandlers) sendMenu(ctx context.Context, upd telego.Update) (engagedomain.Message, error) {
	text, btns := tgh.mainMenu()

	return tgh.tg.SendMessageWithButtons(ctx, upd.Message.Chat.ID, upd.Message.MessageThreadID, text, btns)
}

func (tgh *TgHandlers) changeMenuToMain(ctx context.Context, upd telego.Update) error {
	text, btns := tgh.mainMenu()

	return tgh.tg.EditMessageWithButtons(ctx, upd.CallbackQuery.Message.GetChat().ID, upd.CallbackQuery.Message.GetMessageID(), text, btns)
}

func (tgh *TgHandlers) changeMenuToSupport(ctx context.Context, upd telego.Update) error {
	text := "Поддержка"
	btns := [][]engagedomain.InlineKeyboardButton{
		{
			{
				Text:         "🎟 Создать тикет",
				CallbackData: callbackTicketCreate,
			},
		}, {
			{
				Text:         "◀️ Назад",
				CallbackData: callbackMenuMain,
			},
		},
	}

	return tgh.tg.EditMessageWithButtons(ctx, upd.CallbackQuery.Message.GetChat().ID, upd.CallbackQuery.Message.GetMessageID(), text, btns)
}
