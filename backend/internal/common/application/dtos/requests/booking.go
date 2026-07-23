package requests

import "github.com/google/uuid"

type CreateBooking struct {
	OfferID   uuid.UUID `json:"offer_id"`
	Comment   string    `json:"comment"`
	StartTime Time      `json:"start_time"`
}

type UpdateBooking struct {
	BookingID   uuid.UUID `json:"booking_id"`
	IsCancelled *bool     `json:"is_cancelled"`
	Comment     *string   `json:"comment"`
	StartTime   *Time     `json:"start_time"`
}

type Time struct {
	Date      string  `json:"date"`
	StartTime DayTime `json:"start_time"`
}
