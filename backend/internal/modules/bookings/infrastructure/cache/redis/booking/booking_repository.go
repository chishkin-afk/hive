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

type bookingCacheRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *bookingCacheRepository {
	return &bookingCacheRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

func (bcr *bookingCacheRepository) Save(ctx context.Context, booking *booking.Booking) error {
	bcr.log.Debug("converting booking into bytes", slog.String("booking_id", booking.ID().String()))
	bytes, err := bookingToBytes(booking)
	if err != nil {
		return fmt.Errorf("failed to convert booking into bytes: %w", err)
	}

	if err := bcr.cache.Set(ctx, bcr.getBookingKey(booking.ID()), bytes, bcr.cfg.Cache.BookingTTL).Err(); err != nil {
		if safeErr, ok := bcr.handleError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to save booking into cache: %w", err)
	}

	return nil
}

func (bcr *bookingCacheRepository) GetByID(ctx context.Context, id uuid.UUID) (*booking.Booking, error) {
	bcr.log.Debug("getting bytes from cache", slog.String("booking_id", id.String()))
	bytes, err := bcr.cache.Get(ctx, bcr.getBookingKey(id)).Bytes()
	if err != nil {
		if safeErr, ok := bcr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get booking from cache: %w", err)
	}

	return bytesToBooking(bytes)
}

func (bcr *bookingCacheRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := bcr.cache.Del(ctx, bcr.getBookingKey(id)).Err(); err != nil {
		if safeErr, ok := bcr.handleError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to delete booking from cache: %w", err)
	}

	return nil
}

func (bcr *bookingCacheRepository) getBookingKey(id uuid.UUID) string {
	return fmt.Sprintf("booking:%s", id.String())
}

func (bcr *bookingCacheRepository) handleError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, redis.Nil):
		return booking.ErrBookingNotFound, true
	}

	return err, false
}
