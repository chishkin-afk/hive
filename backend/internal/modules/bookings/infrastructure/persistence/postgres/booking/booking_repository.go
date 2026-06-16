package bookingpg

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/devathh/hive/backend/internal/modules/bookings/domain/booking"
	"github.com/google/uuid"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type bookingPersistenceRepository struct {
	log *slog.Logger
	db  DB
}

func New(log *slog.Logger, db DB) *bookingPersistenceRepository {
	return &bookingPersistenceRepository{
		log: log,
		db:  db,
	}
}

// Save is a function that first checks the number of user's bookings for this coach,
// then requests the coach's offer ID, and calls the callback function with these arguments.
// If the callback function does not return an error, it saves the user.
// Everything happens within a transaction
func (bpr *bookingPersistenceRepository) Save(ctx context.Context, booking *booking.Booking, cbFunc booking.CreateFunc) (*booking.Booking, error) {
	panic("not impl")
}

// ListByOfferID returns a list of all bookings for a specific offer.
// It is important to check whether the user is the author of the offer before doing so
func (bpr *bookingPersistenceRepository) ListByOfferID(ctx context.Context, offerID uuid.UUID, page, limit uint32) ([]*booking.Booking, error) {
	panic("not impl")
}

// ListByUserID is a function that returns a list of all bookings for a specific user.
// The UserID should be taken from the token
func (bpr *bookingPersistenceRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, limit uint32) ([]*booking.Booking, error) {
	panic("not impl")
}

// ListAllByOfferID returns all future bookings for the offer
func (bpr *bookingPersistenceRepository) ListAllByOfferID(ctx context.Context, offerID uuid.UUID) ([]*booking.Booking, error) {
	panic("not impl")
}

// GetByID returns a single booking by its ID.
// It is important to check whether this record belongs to a userID
// or whether the offer owns a userID
func (bpr *bookingPersistenceRepository) GetByID(ctx context.Context, id uuid.UUID) (*booking.Booking, error) {
	panic("not impl")
}

// Update is a function that updates a record (for example, the isCancelled field).
// It receives a booking's ID and calls updFunc, then saves the changes
func (bpr *bookingPersistenceRepository) Update(ctx context.Context, id uuid.UUID, updFunc booking.UpdateFunc) (*booking.Booking, error) {
	panic("not impl")
}
