package tghandlers

import (
	"context"
	"fmt"
	"log/slog"

	ticketapplication "github.com/braunkc/hive-engage/internal/application/ticket"
	engagedomain "github.com/braunkc/hive-engage/internal/domain/engage"
	"github.com/mymmrac/telego"
)

const (
	callbackTopicDelete = "topic_delete"
)

func (tgh *TgHandlers) prepareTopicForTicket(ctx context.Context, name string, ticket ticketapplication.Ticket) (engagedomain.ForumTopic, error) {
	tgh.log.Debug("preparing topic for ticket", slog.String("name", name))

	forumTopic, err := tgh.tg.CreateTopic(ctx, tgh.cfg.SupergroudID, name)
	if err != nil {
		return engagedomain.ForumTopic{}, err
	}

	text := fmt.Sprintf("id: %d\n\nclient id: %d\nusername: @%s\n\ntitle: %s\ndescription: %s",
		ticket.ID, ticket.ClientID.ID, ticket.ClientID.Username, ticket.Content.Title, ticket.Content.Description)
	btns := [][]engagedomain.InlineKeyboardButton{
		{
			{
				Text:         "Закрыть",
				Style:        telego.ButtonStyleDanger,
				CallbackData: fmt.Sprintf("%s_%d", callbackTicketClose, ticket.ID),
			},
		},
	}
	_, err = tgh.tg.SendMessageWithButtons(ctx, tgh.cfg.SupergroudID, forumTopic.ThreadID, text, btns)
	if err != nil {
		return engagedomain.ForumTopic{}, err
	}

	return forumTopic, nil
}

func (tgh *TgHandlers) handleTopicClose(ctx context.Context, ticket ticketapplication.Ticket) error {
	if ticket.ThreadID != 0 {
		btns := [][]engagedomain.InlineKeyboardButton{
			{
				{
					Text:         "Удалить топик",
					Style:        telego.ButtonStyleDanger,
					CallbackData: callbackTopicDelete,
				},
			},
		}
		_, err := tgh.tg.SendMessageWithButtons(ctx, tgh.cfg.SupergroudID, ticket.ThreadID, "Тикет закрыт", btns)
		if err != nil {
			return err
		}

		if err = tgh.tg.CloseTopic(ctx, tgh.cfg.SupergroudID, ticket.ThreadID); err != nil {
			return err
		}
	}

	return nil
}

func (tgh *TgHandlers) handleTopicDelete(ctx context.Context, upd telego.Update) error {
	tgh.log.Debug("handling topic delete")

	return tgh.tg.DeleteTopic(ctx, tgh.cfg.SupergroudID, upd.CallbackQuery.Message.Message().MessageThreadID)
}
