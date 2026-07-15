package bookingredis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/bookings/domain/booking"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, duration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, key ...string) *redis.IntCmd
}

type bookingRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *bookingRepository {
	return &bookingRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

// Save is a function that saves a booking
// to the cache using its ID; otherwise,
// it returns an error.
func (br *bookingRepository) Save(ctx context.Context, booking *booking.Booking) error {
	br.log.Debug("saving booking into cache",
		slog.String("booking_id", booking.ID().String()),
	)

	if err := br.save(ctx, br.cache, booking); err != nil {
		return fmt.Errorf("failed to save booking: %w",
			handleError(err),
		)
	}

	return nil
}

func (br *bookingRepository) save(ctx context.Context, cache Cache, booking *booking.Booking) error {
	bytes, err := bookingToBytes(booking)
	if err != nil {
		return err
	}

	key := getBookingKey(booking.ID())
	return cache.Set(ctx, key, bytes, br.cfg.Cache.BookingTTL).Err()
}

// GetByID is a function that returns a booking
// from the cache based on its ID,
// or an error otherwise.
func (br *bookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*booking.Booking, error) {
	br.log.Debug("getting booking by id from cache",
		slog.String("booking_id", id.String()),
	)

	booking, err := br.get(ctx, br.cache, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking by id: %w",
			handleError(err),
		)
	}

	return booking, nil
}

func (br *bookingRepository) get(ctx context.Context, cache Cache, id uuid.UUID) (*booking.Booking, error) {
	bytes, err := cache.Get(ctx, getBookingKey(id)).Bytes()
	if err != nil {
		return nil, err
	}

	return bytesToBooking(bytes)
}

// Delete is a method that removes a booking
// from the cache by its ID
// otherwise, it returns an error.
func (br *bookingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	br.log.Debug("deleting booking from cache by id",
		slog.String("booking_id", id.String()),
	)

	if err := br.delete(ctx, br.cache, id); err != nil {
		return fmt.Errorf("failed to delete booking: %w",
			handleError(err),
		)
	}

	return nil
}

func (br *bookingRepository) delete(ctx context.Context, cache Cache, id uuid.UUID) error {
	return cache.Del(ctx, getBookingKey(id)).Err()
}

func getBookingKey(id uuid.UUID) string {
	return fmt.Sprintf("booking:%s", id.String())
}

func handleError(err error) error {
	if errors.Is(err, redis.Nil) {
		return booking.ErrBookingNotFound
	}

	return err
}
