package clientredis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/clients/domain/client"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, duration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, key ...string) *redis.IntCmd
}

type clientCacheRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *clientCacheRepository {
	return &clientCacheRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

func (ccr *clientCacheRepository) Save(ctx context.Context, client *client.Client) error {
	panic("not impl")
}

func (ccr *clientCacheRepository) GetByID(ctx context.Context, id uuid.UUID) (*client.Client, error) {
	panic("not impl")
}

func (ccr *clientCacheRepository) Delete(ctx context.Context, id uuid.UUID) error {
	panic("not impl")
}

func (ccr *clientCacheRepository) getClientKey(id uuid.UUID) string {
	return fmt.Sprintf("client:%s", id.String())
}
