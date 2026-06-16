package bookingpg

import (
	"time"

	"github.com/google/uuid"
)

type bookingRecord struct {
	ID          uuid.UUID
	OfferID     uuid.UUID
	UserID      uuid.UUID
	Comment     string
	StartTime   time.Time
	IsCancelled bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
