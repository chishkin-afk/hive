package bookingpg

import "github.com/devathh/hive/backend/internal/modules/bookings/domain/booking"

func bookingToRecord(booking *booking.Booking) *bookingRecord {
	return &bookingRecord{
		ID:          booking.ID(),
		OfferID:     booking.OfferID(),
		UserID:      booking.UserID(),
		Comment:     booking.Comment(),
		StartTime:   booking.StartTime(),
		IsCancelled: booking.IsCancelled(),
		CreatedAt:   booking.CreatedAt(),
		UpdatedAt:   booking.UpdatedAt(),
	}
}

func recordToBooking(record *bookingRecord) *booking.Booking {
	return booking.From(
		record.ID,
		record.OfferID,
		record.UserID,
		record.Comment,
		record.StartTime,
		record.IsCancelled,
		record.CreatedAt,
		record.UpdatedAt,
	)
}
