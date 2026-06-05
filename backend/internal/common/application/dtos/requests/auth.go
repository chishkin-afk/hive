package requests

import "github.com/google/uuid"

type PrepareUser struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Password  string `json:"password"` // can be empty
}

type ConfirmUser struct {
	UserID uuid.UUID `json:"user_id"`
	Code   int       `json:"code"`
}

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUser struct {
	Firstname    *string             `json:"firstname"`
	Lastname     *string             `json:"lastname"`
	Password     *string             `json:"password"`
	OnVacation   *bool               `json:"on_vacation"`
	TgChatID     *int64              `json:"tg_chat_id"`
	Notification *UpdateNotification `json:"notification"`
}

type UpdateNotification struct {
	TelegramEnable bool `json:"telegram_enable"`
}
