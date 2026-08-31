package tgapi

import (
	"context"
	"errors"
	"log/slog"
	"time"

	engagedomain "github.com/braunkc/hive-engage/internal/domain/engage"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

var (
	ErrNilPointer = errors.New("nil pointer")
)

type tgAPI struct {
	bot *telego.Bot
	log *slog.Logger
}

func NewTgAPI(bot *telego.Bot, log *slog.Logger) (engagedomain.EngageAPI, error) {
	if bot == nil {
		return nil, ErrNilPointer
	}

	return &tgAPI{
		bot: bot,
		log: log,
	}, nil
}

func (ta *tgAPI) GetUser(ctx context.Context, userID int64) (engagedomain.User, error) {
	params := telego.GetChatParams{
		ChatID: tu.ID(userID),
	}
	chat, err := ta.bot.GetChat(ctx, &params)
	if err != nil {
		return engagedomain.User{}, err
	}

	return ta.mapUserToDomain(chat), nil
}

func (ta *tgAPI) SendMessage(ctx context.Context, chatID int64, threadID int, text string) (engagedomain.Message, error) {
	ta.log.Debug("sending message",
		slog.Int64("chatID", chatID),
		slog.Int("threadID", threadID),
		slog.String("text", text),
	)

	params := telego.SendMessageParams{
		ChatID:          tu.ID(chatID),
		MessageThreadID: threadID,
		Text:            text,
	}
	msg, err := ta.bot.SendMessage(ctx, &params)
	if err != nil {
		return engagedomain.Message{}, err
	}

	return ta.mapMessageToDomain(msg), nil
}

func (ta *tgAPI) SendMessageWithButtons(ctx context.Context, chatID int64, threadID int, text string, btns [][]engagedomain.InlineKeyboardButton) (engagedomain.Message, error) {
	ta.log.Debug("sending message with buttons",
		slog.Int64("chatID", chatID),
		slog.Int("threadID", threadID),
		slog.String("text", text),
	)

	inlineKeyboard := ta.buildInlineKeyboard(btns)
	params := telego.SendMessageParams{
		ChatID:          tu.ID(chatID),
		MessageThreadID: threadID,
		Text:            text,
		ReplyMarkup:     inlineKeyboard,
	}
	msg, err := ta.bot.SendMessage(ctx, &params)
	if err != nil {
		return engagedomain.Message{}, err
	}

	return ta.mapMessageToDomain(msg), nil
}

func (ta *tgAPI) SendTempMessage(ctx context.Context, chatID int64, threadID int, text string, d time.Duration) (engagedomain.Message, error) {
	ta.log.Debug("sending temp message",
		slog.Int64("chatID", chatID),
		slog.Int("threadID", threadID),
		slog.String("text", text),
	)

	msg, err := ta.SendMessage(ctx, chatID, threadID, text)
	if err != nil {
		return engagedomain.Message{}, err
	}

	go func() {
		<-time.After(d)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := ta.DeleteMessage(ctx, msg.Chat.ID, msg.MessageID); err != nil {
			ta.log.Warn("failed to delete temp message", slog.Any("err", err))
		}
	}()

	return msg, nil
}

func (ta *tgAPI) CopyMessage(ctx context.Context, chatID, fromChatID int64, threadID, messageID int) (int, error) {
	params := telego.CopyMessageParams{
		ChatID:          tu.ID(chatID),
		MessageThreadID: threadID,
		FromChatID:      tu.ID(fromChatID),
		MessageID:       messageID,
	}
	msgID, err := ta.bot.CopyMessage(ctx, &params)
	if err != nil {
		return 0, err
	}

	return msgID.MessageID, nil
}

func (ta *tgAPI) EditMessage(ctx context.Context, chatID int64, messageID int, text string) error {
	ta.log.Debug("editing message",
		slog.Int64("chatID", chatID),
		slog.Int("messageID", messageID),
		slog.String("text", text),
	)

	params := telego.EditMessageTextParams{
		ChatID:    tu.ID(chatID),
		MessageID: messageID,
		Text:      text,
	}
	if _, err := ta.bot.EditMessageText(ctx, &params); err != nil {
		return err
	}

	return nil
}

func (ta *tgAPI) EditMessageWithButtons(ctx context.Context, chatID int64, messageID int, text string, btns [][]engagedomain.InlineKeyboardButton) error {
	ta.log.Debug("editing message with buttons",
		slog.Int64("chatID", chatID),
		slog.Int("messageID", messageID),
		slog.String("text", text),
	)

	inlineKeyboard := ta.buildInlineKeyboard(btns)
	params := telego.EditMessageTextParams{
		ChatID:      tu.ID(chatID),
		MessageID:   messageID,
		Text:        text,
		ReplyMarkup: inlineKeyboard,
	}
	if _, err := ta.bot.EditMessageText(ctx, &params); err != nil {
		return err
	}

	return nil
}

func (ta *tgAPI) DeleteMessage(ctx context.Context, chatID int64, messageID int) error {
	ta.log.Debug("deleting message",
		slog.Int64("chatID", chatID),
		slog.Int("messageID", messageID),
	)

	params := tu.Delete(tu.ID(chatID), messageID)
	if err := ta.bot.DeleteMessage(ctx, params); err != nil {
		return err
	}

	return nil
}

func (ta *tgAPI) PinMessage(ctx context.Context, chatID int64, messageID int) error {
	params := telego.PinChatMessageParams{
		ChatID:    tu.ID(chatID),
		MessageID: messageID,
	}
	return ta.bot.PinChatMessage(ctx, &params)
}

func (ta *tgAPI) UnpinMessage(ctx context.Context, chatID int64, messageID int) error {
	params := telego.UnpinChatMessageParams{
		ChatID:    tu.ID(chatID),
		MessageID: messageID,
	}
	return ta.bot.UnpinChatMessage(ctx, &params)
}

func (ta *tgAPI) AnswerCallbackQuery(ctx context.Context, callbackQueryID, text string, showAlert bool, cacheTime int) error {
	ta.log.Debug("answering to callback query",
		slog.String("callbackQueryID", callbackQueryID),
		slog.String("text", text),
		slog.Int("cacheTime", cacheTime),
	)

	params := telego.AnswerCallbackQueryParams{
		CallbackQueryID: callbackQueryID,
		Text:            text,
		ShowAlert:       showAlert,
		CacheTime:       cacheTime,
	}
	return ta.bot.AnswerCallbackQuery(ctx, &params)
}

func (ta *tgAPI) SendAction(ctx context.Context, chatID int64, threadID int, action string) error {
	ta.log.Debug("sending typing",
		slog.Int64("chatID", chatID),
		slog.Int("threadID", threadID),
		slog.String("action", action),
	)

	params := telego.SendChatActionParams{
		ChatID:          tu.ID(chatID),
		MessageThreadID: threadID,
		Action:          action,
	}
	return ta.bot.SendChatAction(ctx, &params)
}

func (ta *tgAPI) CreateTopic(ctx context.Context, chatID int64, name string) (engagedomain.ForumTopic, error) {
	ta.log.Debug("creating topic",
		slog.Int64("chatID", chatID),
		slog.String("name", name),
	)

	params := telego.CreateForumTopicParams{
		ChatID: tu.ID(chatID),
		Name:   name,
	}
	forumTopic, err := ta.bot.CreateForumTopic(ctx, &params)
	if err != nil {
		return engagedomain.ForumTopic{}, err
	}

	return ta.mapForumTopicToDomain(forumTopic), nil
}

func (ta *tgAPI) CloseTopic(ctx context.Context, chatID int64, threadID int) error {
	ta.log.Debug("closing topic",
		slog.Int64("chatID", chatID),
		slog.Int("threadID", threadID),
	)

	params := telego.CloseForumTopicParams{
		ChatID:          tu.ID(chatID),
		MessageThreadID: threadID,
	}
	return ta.bot.CloseForumTopic(ctx, &params)
}

func (ta *tgAPI) EditTopic(ctx context.Context, chatID int64, threadID int, name string) error {
	ta.log.Debug("editing topic",
		slog.Int64("chatID", chatID),
		slog.Int("threadID", threadID),
		slog.String("name", name),
	)

	params := telego.EditForumTopicParams{
		ChatID:          tu.ID(chatID),
		MessageThreadID: threadID,
		Name:            name,
	}
	return ta.bot.EditForumTopic(ctx, &params)
}

func (ta *tgAPI) DeleteTopic(ctx context.Context, chatID int64, threadID int) error {
	ta.log.Debug("deleting topic",
		slog.Int64("chatID", chatID),
		slog.Int("threadID", threadID),
	)

	params := telego.DeleteForumTopicParams{
		ChatID:          tu.ID(chatID),
		MessageThreadID: threadID,
	}
	return ta.bot.DeleteForumTopic(ctx, &params)
}

func (ta *tgAPI) buildInlineKeyboard(btns [][]engagedomain.InlineKeyboardButton) *telego.InlineKeyboardMarkup {
	if len(btns) == 0 {
		return &telego.InlineKeyboardMarkup{}
	}

	rows := make([][]telego.InlineKeyboardButton, len(btns))
	for i, row := range btns {
		if len(row) == 0 {
			continue
		}

		rows[i] = make([]telego.InlineKeyboardButton, len(row))
		for j, btn := range row {
			rows[i][j] = telego.InlineKeyboardButton{
				Text:         btn.Text,
				Style:        btn.Style,
				URL:          btn.URL,
				CallbackData: btn.CallbackData,
			}
		}
	}

	return &telego.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func (ta *tgAPI) mapMessageToDomain(message *telego.Message) engagedomain.Message {
	return engagedomain.Message{
		MessageID:       message.MessageID,
		MessageThreadID: message.MessageThreadID,
		From: engagedomain.User{
			ID:        message.From.ID,
			FirstName: message.From.FirstName,
			Username:  message.From.Username,
		},
		Chat: engagedomain.Chat{
			ID:      message.Chat.ID,
			Type:    message.Chat.Type,
			Title:   message.Chat.Title,
			IsForum: message.Chat.IsForum,
		},
		Text: message.Text,
	}
}

func (ta *tgAPI) mapUserToDomain(user *telego.ChatFullInfo) engagedomain.User {
	return engagedomain.User{
		ID:        user.ID,
		FirstName: user.FirstName,
		Username:  user.Username,
	}
}

func (ta *tgAPI) mapForumTopicToDomain(forumTopic *telego.ForumTopic) engagedomain.ForumTopic {
	return engagedomain.ForumTopic{
		ThreadID: forumTopic.MessageThreadID,
		Name:     forumTopic.Name,
	}
}
