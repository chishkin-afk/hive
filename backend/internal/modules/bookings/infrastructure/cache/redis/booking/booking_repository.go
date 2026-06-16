package bookingredis

import (
	"context"
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
	panic("not impl")
}

func (bcr *bookingCacheRepository) GetByID(ctx context.Context, id uuid.UUID) (*booking.Booking, error) {
	panic("not impl")
}

func (bcr *bookingCacheRepository) Delete(ctx context.Context, id uuid.UUID) error {
	panic("not impl")
}

func (bcr *bookingCacheRepository) getBookingKey(id uuid.UUID) string {
	panic("not impl")
}
