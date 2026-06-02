package responses

import (
	"github.com/google/uuid"
)

type Token struct {
	Token string `json:"token"`
}

type User struct {
	ID           uuid.UUID    `json:"id"`
	Firstname    string       `json:"firstname"`
	Lastname     string       `json:"lastname"`
	Email        string       `json:"email"`
	PaidUntil    int64        `json:"paid_until"`
	OnVacation   bool         `json:"on_vacation"`
	TgChatID     int64        `json:"tg_chat_id"`
	Notification Notification `json:"notification"`
	CreatedAt    int64        `json:"created_at"`
	UpdatedAt    int64        `json:"updated_at"`
}

type Notification struct {
	TelegramEnable bool `json:"telegram_enable"`
}
