package rateredis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/payments/domain/rate"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	allRatesKey = "allrates"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, duration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, key ...string) *redis.IntCmd
}

type rateCacheRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *rateCacheRepository {
	return &rateCacheRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

func (rcr *rateCacheRepository) SaveRate(ctx context.Context, rate *rate.Rate) error {
	rcr.log.Debug("converting single rate into bytes",
		slog.String("rate_id", rate.ID().String()),
	)

	bytes, err := rateToBytes(rate)
	if err != nil {
		return err
	}

	rcr.log.Debug("saving rate's bytes into cache",
		slog.String("rate_id", rate.ID().String()),
	)

	if err := rcr.cache.Set(ctx, rcr.getSingleKey(rate.ID()), bytes, rcr.cfg.Cache.RateTTL).Err(); err != nil {
		if safeErr, ok := rcr.handleError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to save single rate into cache: %w", err)
	}

	return nil
}

func (rcr *rateCacheRepository) GetRate(ctx context.Context, id uuid.UUID) (*rate.Rate, error) {
	rcr.log.Debug("getting single rate by id from cache",
		slog.String("rate_id", id.String()),
	)

	bytes, err := rcr.cache.Get(ctx, rcr.getSingleKey(id)).Bytes()
	if err != nil {
		if safeErr, ok := rcr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get single rate: %w", err)
	}

	rcr.log.Debug("converting bytes of rate into domain model",
		slog.String("rate_id", id.String()),
	)

	return bytesToRate(bytes)
}

func (rcr *rateCacheRepository) SaveRates(ctx context.Context, rates []*rate.Rate) error {
	rcr.log.Debug("converting a list of rates into bytes",
		slog.Int("len", len(rates)),
	)

	bytes, err := ratesToBytes(rates)
	if err != nil {
		return err
	}

	rcr.log.Debug("saving a list of rates into cache",
		slog.Int("len", len(rates)),
	)
	if err := rcr.cache.Set(ctx, allRatesKey, bytes, rcr.cfg.Cache.RateTTL).Err(); err != nil {
		if safeErr, ok := rcr.handleError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to save a list of rates into cache: %w", err)
	}

	return nil
}

func (rcr *rateCacheRepository) ListRates(ctx context.Context) ([]*rate.Rate, error) {
	rcr.log.Debug("getting a list of rates from cache")

	bytes, err := rcr.cache.Get(ctx, allRatesKey).Bytes()
	if err != nil {
		if safeErr, ok := rcr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get a list of rates from cache: %w", err)
	}

	rcr.log.Debug("converting bytes to a list of rates")

	return bytesToRates(bytes)
}

func (rcr *rateCacheRepository) handleError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, redis.Nil):
		return rate.ErrRateNotFound, true
	}

	return err, false
}

func (rcr *rateCacheRepository) getSingleKey(id uuid.UUID) string {
	return fmt.Sprintf("rate:%s", id.String())
}
