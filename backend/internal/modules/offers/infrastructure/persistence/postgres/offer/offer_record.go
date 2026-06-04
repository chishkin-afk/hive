package offerpg

import (
	"time"

	"github.com/google/uuid"
)

type offerRecord struct {
	ID                  uuid.UUID
	AuthorID            uuid.UUID
	Title               string
	Description         string
	Price               uint64
	SlotDurationMinutes int64
	SlotIntervalMinutes int64
	IsActive            bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
