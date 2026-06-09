package clientredis

import (
	"time"

	"github.com/google/uuid"
)

type clientRecord struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	OfferID   uuid.UUID `json:"offer_id"`
	Visits    uint64    `json:"visits"`
	LastVisit time.Time `json:"last_visit"`
	CreatedAt time.Time `json:"created_at"`
}
