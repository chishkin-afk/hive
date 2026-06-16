package requests

import "github.com/google/uuid"

type CreateBooking struct {
	OfferID   uuid.UUID `json:"offer_id"`
	Comment   string    `json:"comment"`
	StartTime int64     `json:"start_time"` // unix
}

type UpdateBooking struct {
	BookingID   uuid.UUID `json:"booking_id"`
	IsCancelled *bool     `json:"is_cancelled"`
	Comment     *string   `json:"comment"`
	StartTime   *int64    `json:"start_time"` // unix
}
