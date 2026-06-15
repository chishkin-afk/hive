package scheduleredis

import (
	"context"
	"errors"
	"fmt"
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
	scr.log.Debug("converting domain schedules into bytes", slog.String("user_id", userID.String()))
	bytes, err := schedulesToBytes(schedules)
	if err != nil {
		return fmt.Errorf("failed to convert domain into bytes: %w", err)
	}

	scr.log.Debug("setting the schedules into cache", slog.String("user_id", userID.String()))
	if err := scr.cache.Set(ctx, scr.getSchedulesKey(userID), bytes, scr.cfg.Cache.SchedulesTTL).Err(); err != nil {
		if safeErr, ok := scr.handleError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to save schedules into cache: %w", err)
	}

	return nil
}

func (scr *scheduleCacheRepository) GetSchedules(ctx context.Context, userID uuid.UUID) ([]*schedule.Schedule, error) {
	scr.log.Debug("getting the schedules from cache by user's id", slog.String("user_id", userID.String()))
	bytes, err := scr.cache.Get(ctx, scr.getSchedulesKey(userID)).Bytes()
	if err != nil {
		if safeErr, ok := scr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get schedules: %w", err)
	}

	return bytesToSchedules(bytes)
}

func (scr *scheduleCacheRepository) DeleteSchedules(ctx context.Context, userID uuid.UUID) error {
	if err := scr.cache.Del(ctx, scr.getSchedulesKey(userID)).Err(); err != nil {
		if safeErr, ok := scr.handleError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to delete schedules: %w", err)
	}

	return nil
}

func (scr *scheduleCacheRepository) getSchedulesKey(userID uuid.UUID) string {
	return fmt.Sprintf("schedules:%s", userID.String())
}

func (scr *scheduleCacheRepository) handleError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, redis.Nil):
		return schedule.ErrScheduleNotFound, true
	}

	return err, false
}
