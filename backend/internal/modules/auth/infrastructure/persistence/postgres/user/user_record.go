package userpg

import (
	"time"

	"github.com/google/uuid"
)

type userRecord struct {
	ID             uuid.UUID
	Firstname      string
	Lastname       string
	ConfirmedEmail string
	PhoneNumber    string
	PasswordHash   string
	PaidUntil      time.Time
	OnVacation     bool
	TgChatID       int64
	Notification   []byte
	TrialUsed      bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type notification struct {
	TelegramEnable bool `json:"telegram_enable"`
}
