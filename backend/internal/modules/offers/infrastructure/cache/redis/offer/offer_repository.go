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

type OfferCacheRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *OfferCacheRepository {
	return &OfferCacheRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

func (ocr *OfferCacheRepository) Set(ctx context.Context, offer *offer.Offer) error {
	ocr.log.Debug("converting domain into bytes", slog.String("offer_id", offer.ID().String()))
	bytes, err := domainToBytes(offer)
	if err != nil {
		return fmt.Errorf("failed to convert domain into bytes: %w", err)
	}

	if err := ocr.cache.Set(ctx, ocr.getOfferKey(offer.ID()), bytes, ocr.cfg.Cache.OfferTTL).Err(); err != nil {
		if safeErr, ok := ocr.normalizeError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to set offer into cache: %w", err)
	}

	return nil
}

func (ocr *OfferCacheRepository) Get(ctx context.Context, id uuid.UUID) (*offer.Offer, error) {
	ocr.log.Debug("getting offer from cache by id", slog.String("offer_id", id.String()))
	bytes, err := ocr.cache.Get(ctx, ocr.getOfferKey(id)).Bytes()
	if err != nil {
		if safeErr, ok := ocr.normalizeError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get offer from cache: %w", err)
	}

	return bytesToDomain(bytes)
}

func (ocr *OfferCacheRepository) Del(ctx context.Context, id uuid.UUID) error {
	if err := ocr.cache.Del(ctx, ocr.getOfferKey(id)).Err(); err != nil {
		if safeErr, ok := ocr.normalizeError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to delete offer from cache: %w", err)
	}

	return nil
}

func (ocr *OfferCacheRepository) getOfferKey(id uuid.UUID) string {
	return fmt.Sprintf("offer:%s", id.String())
}

func (ocr *OfferCacheRepository) normalizeError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.Canceled),
		errors.Is(err, context.DeadlineExceeded):
		return err, true
	case errors.Is(err, redis.Nil):
		return offer.ErrOfferNotFound, true
	}

	return err, false
}
