package offerredis

import (
	"context"
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
	panic("not impl")
}

func (ocr *OfferCacheRepository) Get(ctx context.Context, id uuid.UUID) (*offer.Offer, error) {
	panic("not impl")
}

func (ocr *OfferCacheRepository) Del(ctx context.Context, id uuid.UUID) error {
	panic("not impl")
}

func (ocr *OfferCacheRepository) getOfferKey(id uuid.UUID) string {
	return fmt.Sprintf("offer:%s", id.String())
}
