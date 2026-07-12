package booking

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type BookingPersistenceRepository interface {
	Save(ctx context.Context, offerID uuid.UUID, startTime, endTime time.Time, cbFunc CreateFunc) (*Booking, error)
	ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) ([]*Booking, int64, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*Booking, int64, error)
	ListAllPerTime(ctx context.Context, offerID uuid.UUID, startTime, endTime time.Time) ([]*Booking, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Booking, error)
	Update(ctx context.Context, id uuid.UUID, updFunc UpdateFunc) (*Booking, error)
}

type BookingCacheRepository interface {
	Save(ctx context.Context, booking *Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*Booking, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

var (
	ErrBookingNotFound      = errors.New("booking not found")
	ErrBookingAlreadyExists = errors.New("booking already exists")
	ErrInvalidOfferID       = errors.New("invalid offer's id")
	ErrTooMuchBookings      = errors.New("too much bookings")
)
