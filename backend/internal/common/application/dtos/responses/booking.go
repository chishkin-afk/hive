package responses

import "github.com/google/uuid"

type Booking struct {
	ID          uuid.UUID `json:"id"`
	OfferID     uuid.UUID `json:"offer_id"`
	UserID      uuid.UUID `json:"user_id"`
	Comment     string    `json:"comment"`
	StartTime   int64     `json:"start_time"`
	IsCancelled bool      `json:"is_cancelled"`
	CreatedAt   int64     `json:"created_at"`
	UpdatedAt   int64     `json:"updated_at"`
}

type Bookings struct {
	List []*Booking `json:"list"`
}
