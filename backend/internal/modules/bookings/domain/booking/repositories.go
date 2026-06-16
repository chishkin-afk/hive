package booking

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type BookingPersistenceRepository interface {
	Save(ctx context.Context, booking *Booking, cbFunc CreateFunc) (*Booking, error)
	ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) ([]*Booking, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*Booking, error)
	ListAllByOfferID(ctx context.Context, offerID uuid.UUID) ([]*Booking, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Booking, error)
	Update(ctx context.Context, id uuid.UUID, updFunc UpdateFunc) (*Booking, error)
}

type BookingCacheRepository interface {
	Save(ctx context.Context, booking *Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*Booking, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

var (
	ErrBookingNotFound = errors.New("booking not found")
)
