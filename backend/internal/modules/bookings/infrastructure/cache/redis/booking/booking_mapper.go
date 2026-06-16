package bookingredis

import (
	"encoding/json"

	"github.com/devathh/hive/backend/internal/modules/bookings/domain/booking"
)

func bookingToBytes(booking *booking.Booking) ([]byte, error) {
	return json.Marshal(bookingRecord{
		ID:          booking.ID(),
		OfferID:     booking.OfferID(),
		UserID:      booking.UserID(),
		Comment:     booking.Comment(),
		StartTime:   booking.StartTime(),
		IsCancelled: booking.IsCancelled(),
		CreatedAt:   booking.CreatedAt(),
		UpdatedAt:   booking.UpdatedAt(),
	})
}

func bytesToBooking(bytes []byte) (*booking.Booking, error) {
	var record bookingRecord
	if err := json.Unmarshal(bytes, &record); err != nil {
		return nil, err
	}

	return booking.From(
		record.ID,
		record.OfferID,
		record.UserID,
		record.Comment,
		record.StartTime,
		record.IsCancelled,
		record.CreatedAt,
		record.UpdatedAt,
	), nil
}
