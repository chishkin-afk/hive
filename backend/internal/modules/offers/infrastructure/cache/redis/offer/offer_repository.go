package offerredis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/offers/domain/offer"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, duration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, key ...string) *redis.IntCmd
}

type OfferRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *OfferRepository {
	return &OfferRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

// Set is a method that stores an offer
// in the cache by its ID
// otherwise, it returns an error.
func (or *OfferRepository) Set(ctx context.Context, offer *offer.Offer) error {
	or.log.Debug("setting offer into cache",
		slog.String("offer_id", offer.ID().String()),
	)

	if err := or.set(ctx, or.cache, offer); err != nil {
		return fmt.Errorf("failed to set offer into cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (or *OfferRepository) set(ctx context.Context, cache Cache, offer *offer.Offer) error {
	bytes, err := domainToBytes(offer)
	if err != nil {
		return err
	}

	key := getOfferKey(offer.ID())
	return cache.Set(ctx, key, bytes, or.cfg.Cache.OfferTTL).Err()
}

// Get is a method that returns an offer
// from the cache by its ID,
// or otherwise returns an error.
func (or *OfferRepository) Get(ctx context.Context, id uuid.UUID) (*offer.Offer, error) {
	or.log.Debug("getting offer by its id",
		slog.String("offer_id", id.String()),
	)

	offer, err := or.get(ctx, or.cache, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get offer from cache: %w",
			handleError(err),
		)
	}

	return offer, nil
}

func (or *OfferRepository) get(ctx context.Context, cache Cache, id uuid.UUID) (*offer.Offer, error) {
	bytes, err := cache.Get(ctx, getOfferKey(id)).Bytes()
	if err != nil {
		return nil, err
	}

	return bytesToDomain(bytes)
}

// Del is a method that removes an offer
// from the cache by its ID,
// otherwise, it returns an error.
func (or *OfferRepository) Del(ctx context.Context, id uuid.UUID) error {
	or.log.Debug("deleting offer by id",
		slog.String("offer_id", id.String()),
	)

	if err := or.del(ctx, or.cache, id); err != nil {
		return fmt.Errorf("failed to delete offer from cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (or *OfferRepository) del(ctx context.Context, cache Cache, id uuid.UUID) error {
	return cache.Del(ctx, getOfferKey(id)).Err()
}

func getOfferKey(id uuid.UUID) string {
	return fmt.Sprintf("offer:%s", id.String())
}

func handleError(err error) error {
	if errors.Is(err, redis.Nil) {
		return offer.ErrOfferNotFound
	}

	return err
}
