package engagedomain

import (
	"context"
	"time"
)

type EngageAPI interface {
	GetUser(ctx context.Context, userID int64) (User, error)
	SendMessage(ctx context.Context, chatID int64, threadID int, text string) (Message, error)
	SendMessageWithButtons(ctx context.Context, chatID int64, threadID int, text string, btns [][]InlineKeyboardButton) (Message, error)
	SendTempMessage(ctx context.Context, chatID int64, threadID int, text string, d time.Duration) (Message, error)
	CopyMessage(ctx context.Context, chatID, fromChatID int64, threadID, messageID int) (int, error)
	EditMessage(ctx context.Context, chatID int64, messageID int, text string) error
	EditMessageWithButtons(ctx context.Context, chatID int64, messageID int, text string, btns [][]InlineKeyboardButton) error
	DeleteMessage(ctx context.Context, chatID int64, messageID int) error
	PinMessage(ctx context.Context, chatID int64, messageID int) error
	UnpinMessage(ctx context.Context, chatID int64, messageID int) error
	AnswerCallbackQuery(ctx context.Context, callbackQueryID, text string, showAlert bool, cacheTime int) error
	SendAction(ctx context.Context, chatID int64, threadID int, action string) error
	CreateTopic(ctx context.Context, chatID int64, name string) (ForumTopic, error)
	CloseTopic(ctx context.Context, chatID int64, threadID int) error
	EditTopic(ctx context.Context, chatID int64, threadID int, name string) error
	DeleteTopic(ctx context.Context, chatID int64, threadID int) error
}
