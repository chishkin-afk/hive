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

type rateRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *rateRepository {
	return &rateRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

// Save stores the provided rate in the cache.
//
// The cached data expires after the configured TTL. An error is returned if
// the rate cannot be serialized or the cache write operation fails.
func (rr *rateRepository) Save(ctx context.Context, rate *rate.Rate) error {
	rr.log.Debug("saving rate into cache",
		slog.String("rate_id", rate.ID().String()),
	)

	if err := rr.save(ctx, rr.cache, rate); err != nil {
		return fmt.Errorf("failed to save rate into cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (rr *rateRepository) save(ctx context.Context, cache Cache, rate *rate.Rate) error {
	bytes, err := rateToBytes(rate)
	if err != nil {
		return err
	}

	key := getSingleKey(rate.ID())
	return cache.Set(ctx, key, bytes, rr.cfg.Cache.RateTTL).Err()
}

// Get retrieves the cached rate identified by id.
//
// It returns the cached rate or an error if the cache lookup or
// deserialization fails.
func (rr *rateRepository) Get(ctx context.Context, id uuid.UUID) (*rate.Rate, error) {
	rr.log.Debug("getting rate by id",
		slog.String("rate_id", id.String()),
	)

	rate, err := rr.get(ctx, rr.cache, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate by id: %w",
			handleError(err),
		)
	}

	return rate, nil
}

func (rr *rateRepository) get(ctx context.Context, cache Cache, id uuid.UUID) (*rate.Rate, error) {
	bytes, err := cache.Get(ctx, getSingleKey(id)).Bytes()
	if err != nil {
		return nil, err
	}

	return bytesToRate(bytes)
}

// SaveList caches the provided rates.
//
// The cached data expires after the configured TTL. An error is returned if
// the rates cannot be encoded or the cache write operation fails.
func (rr *rateRepository) SaveList(ctx context.Context, rates []*rate.Rate) error {
	rr.log.Debug("saving rate list into cache",
		slog.Int("len_rates", len(rates)),
	)

	if err := rr.saveList(ctx, rr.cache, rates); err != nil {
		return fmt.Errorf("failed to save rate list into cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (rr *rateRepository) saveList(ctx context.Context, cache Cache, rates []*rate.Rate) error {
	bytes, err := ratesToBytes(rates)
	if err != nil {
		return err
	}

	return cache.Set(ctx, allRatesKey, bytes, rr.cfg.Cache.RateTTL).Err()
}

// ListRates retrieves the cached list of rates.
//
// It returns the cached rates or an error if the cache lookup or
// deserialization fails.
func (rr *rateRepository) ListRates(ctx context.Context) ([]*rate.Rate, error) {
	rr.log.Debug("getting rate list from cache")

	list, err := rr.listRates(ctx, rr.cache)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate list: %w",
			handleError(err),
		)
	}

	return list, nil
}

func (rr *rateRepository) listRates(ctx context.Context, cache Cache) ([]*rate.Rate, error) {
	bytes, err := cache.Get(ctx, allRatesKey).Bytes()
	if err != nil {
		return nil, err
	}

	return bytesToRates(bytes)
}

func handleError(err error) error {
	if errors.Is(err, redis.Nil) {
		return rate.ErrRateNotFound
	}

	return err
}

func getSingleKey(id uuid.UUID) string {
	return fmt.Sprintf("rate:%s", id.String())
}
