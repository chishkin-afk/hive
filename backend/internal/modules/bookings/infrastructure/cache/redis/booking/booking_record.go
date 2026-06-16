package bookingredis

import (
	"time"

	"github.com/google/uuid"
)

type bookingRecord struct {
	ID          uuid.UUID `json:"id"`
	OfferID     uuid.UUID `json:"offer_id"`
	UserID      uuid.UUID `json:"user_id"`
	Comment     string    `json:"comment"`
	StartTime   time.Time `json:"start_time"`
	IsCancelled bool      `json:"is_cancelled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
