package scheduleredis

import (
	"context"
	"log/slog"
	"time"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/schedules/domain/schedule"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, duration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, key ...string) *redis.IntCmd
}

type scheduleCacheRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *scheduleCacheRepository {
	return &scheduleCacheRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

func (scr *scheduleCacheRepository) SaveSchedules(ctx context.Context, userID uuid.UUID, schedules []*schedule.Schedule) error {
	panic("not impl")
}

func (scr *scheduleCacheRepository) GetSchedules(ctx context.Context, userID uuid.UUID) ([]*schedule.Schedule, error) {
	panic("not impl")
}

func (scr *scheduleCacheRepository) DeleteSchedules(ctx context.Context, userID uuid.UUID) error {
	panic("not impl")
}
