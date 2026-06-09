package clientpg

import (
	"time"

	"github.com/google/uuid"
)

type clientRecord struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	OfferID   uuid.UUID
	Visits    uint64
	LastVisit time.Time
	CreatedAt time.Time
}
