package tghandlers

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/braunkc/hive-engage/config"
	ticketapplication "github.com/braunkc/hive-engage/internal/application/ticket"
	engagedomain "github.com/braunkc/hive-engage/internal/domain/engage"
	ticketdomain "github.com/braunkc/hive-engage/internal/domain/ticket"
	httpclient "github.com/braunkc/hive-engage/internal/infra/http/client"
	statecache "github.com/braunkc/hive-engage/internal/infra/repository/cache/state"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

var (
	ErrNilPointer = errors.New("nil pointer")
)

type TgHandlers struct {
	tg            engagedomain.EngageAPI
	client        *httpclient.HTTPClient
	ticketService *ticketapplication.TicketService
	stateRepo     *statecache.StateRepo
	cfg           config.TgConfig
	log           *slog.Logger
}

func NewTgHandlers(
	tg engagedomain.EngageAPI,
	client *httpclient.HTTPClient,
	ticketService *ticketapplication.TicketService,
	stateRepo *statecache.StateRepo,
	cfg config.TgConfig,
	log *slog.Logger) (*TgHandlers, error) {
	if ticketService == nil || stateRepo == nil {
		return nil, ErrNilPointer
	}

	return &TgHandlers{
		tg:            tg,
		client:        client,
		ticketService: ticketService,
		stateRepo:     stateRepo,
		cfg:           cfg,
		log:           log,
	}, nil
}

func (tgh *TgHandlers) SetupHandlers(bot *telego.Bot, updates <-chan telego.Update) (*th.BotHandler, error) {
	if bot == nil {
		return nil, ErrNilPointer
	}

	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		return nil, err
	}

	tgh.setupCMDHandlers(bh)
	tgh.setupCallbackQueryHandlers(bh)
	tgh.setupMsgHandlers(bh)

	return bh, nil
}

func (tgh *TgHandlers) handleError(ctx context.Context, upd telego.Update, err error) error {
	if err == nil {
		return nil
	}

	tgh.log.Debug("handling error", slog.Any("err", err))

	text := tgh.errorMap(err)

	var chatID int64
	var threadID int
	if upd.Message != nil {
		chatID = upd.Message.Chat.ID
		threadID = upd.Message.MessageThreadID
	} else if upd.CallbackQuery != nil && upd.CallbackQuery.Message != nil {
		if err = tgh.stateRepo.Delete(ctx, upd.CallbackQuery.From.ID); err != nil {
			tgh.log.Warn("failed to delete state from cache", slog.Any("err", err))
		}

		if err = tgh.changeMenuToMain(ctx, upd); err != nil {
			tgh.log.Warn("failed to change menu to main", slog.Any("err", err))
		}

		chatID = upd.CallbackQuery.Message.GetChat().ID
		threadID = upd.CallbackQuery.Message.Message().MessageThreadID
	}

	if _, err = tgh.tg.SendTempMessage(ctx, chatID, threadID, text, time.Minute); err != nil {
		return err
	}

	return nil
}

func (tgh *TgHandlers) errorMap(err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, ErrDescriptionTooLong):
		return "Описание слишком длинное. Пожалуйста, сократите его до 500 символов"
	case errors.Is(err, statecache.ErrRecordNotFound):
		return "Сессия истекла. Пожалуйста, попробуйте заново"
	case errors.Is(err, ticketdomain.ErrTicketNotFound):
		return "Тикет не найден"
	case errors.Is(err, ticketdomain.ErrTicketAlreadyExists):
		return "Вы не можете создать еще 1 тикет"
	case errors.Is(err, httpclient.ErrTgBindingsNotFound):
		return "Ошибка авторизации. Невалидный аргумент"
	default:
		tgh.log.Error("internal error", slog.Any("err", err))
		return "Произошла внутренняя ошибка. Пожалуйста, попробуйте позже"
	}
}
