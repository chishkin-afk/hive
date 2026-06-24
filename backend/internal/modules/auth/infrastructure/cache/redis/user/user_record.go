package userredis

import (
	"time"

	"github.com/google/uuid"
)

type userRecord struct {
	ID             uuid.UUID    `json:"id"`
	Firstname      string       `json:"firstname"`
	Lastname       string       `json:"lastname"`
	ConfirmedEmail string       `json:"confirmed_email"`
	PhoneNumber    string       `json:"phone_number"`
	PasswordHash   string       `json:"password_hash,omitempty"`
	PaidUntil      time.Time    `json:"paid_until"`
	OnVacation     bool         `json:"on_vacation"`
	TgChatID       int64        `json:"tg_chat_id"`
	Notification   notification `json:"notification"`
	TrialUsed      bool         `json:"trial_used"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type notification struct {
	TelegramEnable bool `json:"telegram_enable"`
}

type preparedUserRecord struct {
	User              userRecord `json:"user_record"`
	Code              int        `json:"code"`
	GeneratedPassword string     `json:"generated_password,omitempty"`
}
