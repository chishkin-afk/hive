package offerredis

import (
	"time"

	"github.com/google/uuid"
)

type offerRecord struct {
	ID                  uuid.UUID `json:"id"`
	AuthorID            uuid.UUID `json:"author_id"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	Price               uint64    `json:"price"`
	SlotDurationMinutes int64     `json:"slot_duration_minutes"`
	SlotIntervalMinutes int64     `json:"slot_interval_minutes"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
