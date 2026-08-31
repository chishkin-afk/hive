package tghandlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	ticketapplication "github.com/braunkc/hive-engage/internal/application/ticket"
	engagedomain "github.com/braunkc/hive-engage/internal/domain/engage"
	ticketdomain "github.com/braunkc/hive-engage/internal/domain/ticket"
	statecache "github.com/braunkc/hive-engage/internal/infra/repository/cache/state"
	"github.com/mymmrac/telego"
)

const (
	callbackTicketConfirm = "ticket_confirm"
	callbackTicketCancel  = "ticket_cancel"
	callbackTicketAssign  = "ticket_assign"
	callbackTicketClose   = "ticket_close"

	stepTitle       = "title"
	stepDescription = "description"
	stepConfirm     = "confirm"

	cacheKeyMessageID   = "messageID"
	cacheKeyTitle       = "title"
	cacheKeyDescription = "description"
)

var (
	ErrNoMessageID        = errors.New("no message id in state data")
	ErrNoChatID           = errors.New("no chat id in update")
	ErrNoTitle            = errors.New("no ticket title in state data")
	ErrNoDescription      = errors.New("no ticket description in state data")
	ErrDescriptionTooLong = errors.New("description too long")
)

func (tgh *TgHandlers) renderTicketStepTitle(ctx context.Context, upd telego.Update, state statecache.State) error {
	tgh.log.Debug("rendering ticket step title", slog.Int64("clientID", state.ClientID))

	text := "Напишите краткий заголовок проблемы"
	btns := [][]engagedomain.InlineKeyboardButton{
		{
			{
				Text:         "Отмена",
				Style:        telego.ButtonStyleDanger,
				CallbackData: callbackTicketCancel,
			},
		},
	}

	return tgh.editMessage(ctx, upd, state, text, btns)
}

func (tgh *TgHandlers) renderTicketStepDescription(ctx context.Context, upd telego.Update, state statecache.State) error {
	tgh.log.Debug("rendering ticket step description", slog.Int64("clientID", state.ClientID))

	text := "Опишите проблему, максимум 500 символов"
	btns := [][]engagedomain.InlineKeyboardButton{
		{
			{
				Text:         "Отмена",
				Style:        telego.ButtonStyleDanger,
				CallbackData: callbackTicketCancel,
			},
		},
	}

	return tgh.editMessage(ctx, upd, state, text, btns)
}

func (tgh *TgHandlers) renderTicketStepConfirm(ctx context.Context, upd telego.Update, state statecache.State) error {
	tgh.log.Debug("rendering ticket step confirm", slog.Int64("clientID", state.ClientID))

	title, ok := state.Data[cacheKeyTitle].(string)
	if !ok {
		return ErrNoTitle
	}

	description, ok := state.Data[cacheKeyDescription].(string)
	if !ok {
		return ErrNoDescription
	}

	text := fmt.Sprintf("Подтвердите данные тикета\n\nзаголовок: %s\nописание: %s", title, description)
	btns := [][]engagedomain.InlineKeyboardButton{
		{
			{
				Text:         "Подтвердить",
				Style:        telego.ButtonStyleSuccess,
				CallbackData: callbackTicketConfirm,
			}, {
				Text:         "Отмена",
				Style:        telego.ButtonStyleDanger,
				CallbackData: callbackTicketCancel,
			},
		},
	}

	return tgh.editMessage(ctx, upd, state, text, btns)
}

func (tgh *TgHandlers) handleTicketStepTitle(ctx context.Context, upd telego.Update, state statecache.State) error {
	tgh.log.Debug("handling ticket step title", slog.Int64("clientID", state.ClientID))

	if upd.Message == nil || upd.Message.From == nil {
		return ErrNilPointer
	}

	state.Data[cacheKeyTitle] = upd.Message.Text
	state.Step = stepDescription

	if err := tgh.stateRepo.Save(ctx, state); err != nil {
		return err
	}

	if err := tgh.tg.DeleteMessage(ctx, upd.Message.Chat.ID, upd.Message.MessageID); err != nil {
		return err
	}

	return tgh.renderTicketStepDescription(ctx, upd, state)
}

func (tgh *TgHandlers) handleTicketStepDescription(ctx context.Context, upd telego.Update, state statecache.State) error {
	tgh.log.Debug("handling ticket step description", slog.Int64("clientID", state.ClientID))

	if upd.Message == nil || upd.Message.From == nil {
		return ErrNilPointer
	}

	text := upd.Message.Text
	if len(text) > 500 {
		return ErrDescriptionTooLong
	}

	state.Data[cacheKeyDescription] = text
	state.Step = stepConfirm

	if err := tgh.stateRepo.Save(ctx, state); err != nil {
		return err
	}

	if err := tgh.tg.DeleteMessage(ctx, upd.Message.Chat.ID, upd.Message.MessageID); err != nil {
		return err
	}

	return tgh.renderTicketStepConfirm(ctx, upd, state)
}

func (tgh *TgHandlers) handleTicketCreate(ctx context.Context, upd telego.Update) error {
	tgh.log.Debug("handling ticket create", slog.Int64("clientID", upd.CallbackQuery.From.ID))

	if upd.CallbackQuery == nil || upd.CallbackQuery.Message == nil {
		return ErrNilPointer
	}

	_, err := tgh.ticketService.GetOpenByClientID(ctx, ticketapplication.GetRequest{ID: upd.CallbackQuery.From.ID})
	if err == nil {
		if err = tgh.tg.AnswerCallbackQuery(ctx, upd.CallbackQuery.ID, "Вы не можете создать еще 1 тикет", false, 60); err != nil {
			tgh.log.Warn("failed to answer callback query", slog.Any("err", err))
		}

		return nil
	}
	if !errors.Is(err, ticketdomain.ErrTicketNotFound) {
		return err
	}

	state := statecache.State{
		ClientID: upd.CallbackQuery.From.ID,
		Step:     stepTitle,
		Data: map[string]any{
			cacheKeyMessageID: upd.CallbackQuery.Message.GetMessageID(),
		},
	}

	if err := tgh.stateRepo.Save(ctx, state); err != nil {
		return err
	}

	return tgh.renderTicketStepTitle(ctx, upd, state)
}

func (tgh *TgHandlers) handleTicketConfirm(ctx context.Context, upd telego.Update) error {
	tgh.log.Debug("handling ticket confirm", slog.Int64("clientID", upd.CallbackQuery.From.ID))

	if upd.CallbackQuery == nil {
		return ErrNilPointer
	}

	state, err := tgh.stateRepo.Get(ctx, upd.CallbackQuery.From.ID)
	if err != nil {
		return err
	}

	title, ok := state.Data[cacheKeyTitle].(string)
	if !ok || title == "" {
		return ErrNoTitle
	}

	description, ok := state.Data[cacheKeyDescription].(string)
	if !ok || description == "" {
		return ErrNoDescription
	}

	ticket, err := tgh.ticketService.Create(ctx, ticketapplication.CreateRequest{
		ClientID:    state.ClientID,
		Username:    upd.CallbackQuery.From.Username,
		Title:       title,
		Description: description,
	})
	if err != nil {
		return err
	}

	text := fmt.Sprintf("Имя: %s\n@%s\n\ntitle: %s\ndesc: %s",
		upd.CallbackQuery.From.FirstName, upd.CallbackQuery.From.Username, title, description)
	btns := [][]engagedomain.InlineKeyboardButton{
		{
			{
				Text:         "Взять",
				Style:        telego.ButtonStylePrimary,
				CallbackData: fmt.Sprintf("%s_%d", callbackTicketAssign, ticket.ID),
			}, {
				Text:         "Закрыть",
				Style:        telego.ButtonStyleDanger,
				CallbackData: fmt.Sprintf("%s_%d", callbackTicketClose, ticket.ID),
			},
		},
	}
	if _, err = tgh.tg.SendMessageWithButtons(ctx, tgh.cfg.SupergroudID, tgh.cfg.TicketsThreadID, text, btns); err != nil {
		return err
	}

	if err = tgh.clearTicketCache(ctx, upd); err != nil {
		tgh.log.Warn("failed to clear ticket cache", slog.Any("err", err))
	}

	if err = tgh.changeMenuToMain(ctx, upd); err != nil {
		return err
	}

	if err = tgh.tg.AnswerCallbackQuery(ctx, upd.CallbackQuery.ID, "Тикет создан", false, 1); err != nil {
		tgh.log.Warn("failed to answer callback query", slog.Any("err", err))
	}

	return nil
}

func (tgh *TgHandlers) handleTicketCancel(ctx context.Context, upd telego.Update) error {
	tgh.log.Debug("handling ticket cancel", slog.Int64("clientID", upd.CallbackQuery.From.ID))

	if upd.CallbackQuery == nil || upd.CallbackQuery.Message == nil {
		return ErrNilPointer
	}

	if err := tgh.clearTicketCache(ctx, upd); err != nil {
		tgh.log.Warn("failed to clear ticket cache", slog.Any("err", err))
	}

	if err := tgh.changeMenuToMain(ctx, upd); err != nil {
		return err
	}

	if err := tgh.tg.AnswerCallbackQuery(ctx, upd.CallbackQuery.ID, "Отменено", false, 1); err != nil {
		tgh.log.Warn("failed to answer callback query", slog.Any("err", err))
	}

	return nil
}

func (tgh *TgHandlers) handleTicketAssign(ctx context.Context, upd telego.Update, ticketID int64) error {
	tgh.log.Debug("handling ticket assign",
		slog.Int64("ticketID", ticketID),
		slog.Int64("takerID", upd.CallbackQuery.From.ID),
	)

	if upd.CallbackQuery == nil || upd.CallbackQuery.Message == nil {
		return ErrNilPointer
	}

	ticket, err := tgh.ticketService.Get(ctx, ticketapplication.GetRequest{ID: ticketID})
	if err != nil {
		return err
	}

	forumTopic, err := tgh.prepareTopicForTicket(ctx, fmt.Sprintf("%d %s", ticketID, upd.CallbackQuery.From.Username), ticket)
	if err != nil {
		return err
	}

	text := "Тикет взят в работу, с этого момента до закрытия тикета все отправленные сообщения будут направлены в поддержку"
	_, err = tgh.tg.SendMessage(ctx, ticket.ClientID.ID, 0, text)
	if err != nil {
		return err
	}

	ticket, err = tgh.ticketService.Assign(ctx, ticketapplication.AssignRequest{
		TicketID:         ticketID,
		ThreadID:         forumTopic.ThreadID,
		AssignToID:       upd.CallbackQuery.From.ID,
		AssignToUsername: upd.CallbackQuery.From.Username,
	})
	if err != nil {
		return err
	}

	if err = tgh.tg.DeleteMessage(ctx, upd.CallbackQuery.Message.GetChat().ID, upd.CallbackQuery.Message.GetMessageID()); err != nil {
		return err
	}

	return nil
}

func (tgh *TgHandlers) handleTicketClose(ctx context.Context, upd telego.Update, ticketID int64) error {
	tgh.log.Debug("handling ticket close",
		slog.Int64("ticketID", ticketID),
	)

	if upd.CallbackQuery == nil || upd.CallbackQuery.Message == nil {
		return ErrNilPointer
	}

	ticket, err := tgh.ticketService.Close(ctx, ticketapplication.CloseRequest{
		TicketID: ticketID,
	})
	if err != nil {
		return err
	}

	_, err = tgh.tg.SendMessage(ctx, ticket.ClientID.ID, 0, "Тикет закрыт поддержкой")
	if err != nil {
		return err
	}

	if err = tgh.handleTopicClose(ctx, ticket); err != nil {
		return err
	}

	if err = tgh.tg.AnswerCallbackQuery(ctx, upd.CallbackQuery.ID, "Тикет закрыт", false, 1); err != nil {
		return err
	}

	if err = tgh.tg.DeleteMessage(ctx, upd.CallbackQuery.Message.GetChat().ID, upd.CallbackQuery.Message.GetMessageID()); err != nil {
		tgh.log.Warn("failed to delete message", slog.Any("err", err))
	}

	return nil
}

func (tgh *TgHandlers) handleEcho(ctx context.Context, upd telego.Update, ticket ticketapplication.Ticket) error {
	tgh.log.Debug("handling ticket echo",
		slog.Int64("ticketID", ticket.ID),
	)

	var chatID int64
	var threadID int
	if upd.Message.Chat.ID == ticket.ClientID.ID {
		chatID = tgh.cfg.SupergroudID
		threadID = ticket.ThreadID
	} else {
		chatID = ticket.ClientID.ID
	}

	_, err := tgh.tg.CopyMessage(ctx, chatID, upd.Message.Chat.ID, threadID, upd.Message.MessageID)
	if err != nil {
		return err
	}

	return nil
}

func (tgh *TgHandlers) getTicketHandlers() map[string]stepHandler {
	return map[string]stepHandler{
		stepTitle:       tgh.handleTicketStepTitle,
		stepDescription: tgh.handleTicketStepDescription,
	}
}

func (tgh *TgHandlers) getTicketStepHandler(step string) stepHandler {
	handlers := tgh.getTicketHandlers()

	handler, ok := handlers[step]
	if !ok {
		return nil
	}

	return handler
}

func (tgh *TgHandlers) getTicketMessageID(state statecache.State) (int, error) {
	v, ok := state.Data[cacheKeyMessageID]
	if !ok {
		return 0, ErrNoMessageID
	}

	switch v := v.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	default:
		return 0, ErrNoMessageID
	}
}

func (tgh *TgHandlers) getChatID(upd telego.Update) (int64, error) {
	switch {
	case upd.CallbackQuery != nil && upd.CallbackQuery.Message != nil:
		return upd.CallbackQuery.Message.GetChat().ID, nil
	case upd.Message != nil:
		return upd.Message.Chat.ID, nil
	default:
		tgh.log.Warn("unsupported update type for chat ID extraction", slog.Any("update", upd))
		return 0, ErrNoChatID
	}
}

func (tgh *TgHandlers) editMessage(
	ctx context.Context,
	upd telego.Update,
	state statecache.State,
	text string,
	btns [][]engagedomain.InlineKeyboardButton,
) error {
	messageID, err := tgh.getTicketMessageID(state)
	if err != nil {
		if err = tgh.clearTicketCache(ctx, upd); err != nil {
			tgh.log.Warn("failed to clear ticket cache", slog.Any("err", err))
		}

		return err
	}

	chatID, err := tgh.getChatID(upd)
	if err != nil {
		if err = tgh.clearTicketCache(ctx, upd); err != nil {
			tgh.log.Warn("failed to clear ticket cache", slog.Any("err", err))
		}

		return err
	}

	return tgh.tg.EditMessageWithButtons(ctx, chatID, messageID, text, btns)
}

func (tgh *TgHandlers) clearTicketCache(ctx context.Context, upd telego.Update) error {
	if upd.Message != nil {
		if err := tgh.stateRepo.Delete(ctx, upd.Message.From.ID); err != nil {
			return err
		}
	} else if upd.CallbackQuery != nil {
		if err := tgh.stateRepo.Delete(ctx, upd.CallbackQuery.From.ID); err != nil {
			return err
		}
	}

	return nil
}
