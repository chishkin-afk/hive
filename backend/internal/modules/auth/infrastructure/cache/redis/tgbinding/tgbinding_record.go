package tgbindingredis

import (
	"time"

	"github.com/google/uuid"
)

type tgBindingRecord struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	BotURL    string    `json:"bot_url"`
	CreatedAt time.Time `json:"created_at"`
}
