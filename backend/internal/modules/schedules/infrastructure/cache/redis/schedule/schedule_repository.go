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

type scheduleRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *scheduleRepository {
	return &scheduleRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

// Save stores the provided schedules in the cache for the specified user.
//
// The cached data expires after the configured TTL. An error is returned if
// the schedules cannot be serialized or the cache operation fails.
func (sr *scheduleRepository) Save(ctx context.Context, userID uuid.UUID, schedules []*schedule.Schedule) error {
	sr.log.Debug("saving schedules into cache",
		slog.String("user_id", userID.String()),
	)

	if err := sr.save(ctx, sr.cache, userID, schedules); err != nil {
		return fmt.Errorf("failed to save schedules into cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (sr *scheduleRepository) save(ctx context.Context, cache Cache, userID uuid.UUID, schedules []*schedule.Schedule) error {
	bytes, err := schedulesToBytes(schedules)
	if err != nil {
		return err
	}

	key := getSchedulesKey(userID)
	return cache.Set(ctx, key, bytes, sr.cfg.Cache.SchedulesTTL).Err()
}

// Get retrieves the cached schedules for the specified user.
//
// It returns the cached schedules or an error if the cache lookup or
// deserialization fails.
func (sr *scheduleRepository) Get(ctx context.Context, userID uuid.UUID) ([]*schedule.Schedule, error) {
	sr.log.Debug("getting schedules by user's id from cache",
		slog.String("user_id", userID.String()),
	)

	schedules, err := sr.get(ctx, sr.cache, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules from cache: %w",
			handleError(err),
		)
	}

	return schedules, nil
}

func (sr *scheduleRepository) get(ctx context.Context, cache Cache, userID uuid.UUID) ([]*schedule.Schedule, error) {
	bytes, err := cache.Get(ctx, getSchedulesKey(userID)).Bytes()
	if err != nil {
		return nil, err
	}

	return bytesToSchedules(bytes)
}

// Delete removes the cached schedules for the specified user.
//
// It returns an error if the cache delete operation fails.
func (sr *scheduleRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	sr.log.Debug("deleting schedules by user's id from cache",
		slog.String("user_id", userID.String()),
	)

	if err := sr.delete(ctx, sr.cache, userID); err != nil {
		return fmt.Errorf("failed to delete schedules from cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (sr *scheduleRepository) delete(ctx context.Context, cache Cache, userID uuid.UUID) error {
	return cache.Del(ctx, getSchedulesKey(userID)).Err()
}

func getSchedulesKey(userID uuid.UUID) string {
	return fmt.Sprintf("schedules:%s", userID.String())
}

func handleError(err error) error {
	if errors.Is(err, redis.Nil) {
		return schedule.ErrScheduleNotFound
	}

	return err
}
