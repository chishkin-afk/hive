package clientredis

import (
	"context"
	"errors"
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
	bytes, err := clientToBytes(client)
	if err != nil {
		return fmt.Errorf("failed to convert client into bytes: %w", err)
	}

	if err := ccr.cache.Set(ctx, ccr.getClientKey(client.ID()), bytes, ccr.cfg.Cache.ClientTTL).Err(); err != nil {
		if safeErr, ok := ccr.handleError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to set client into cache: %w", err)
	}

	return nil
}

func (ccr *clientCacheRepository) GetByID(ctx context.Context, id uuid.UUID) (*client.Client, error) {
	bytes, err := ccr.cache.Get(ctx, ccr.getClientKey(id)).Bytes()
	if err != nil {
		if safeErr, ok := ccr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get client from cache: %w", err)
	}

	return bytesToClient(bytes)
}

func (ccr *clientCacheRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := ccr.cache.Del(ctx, ccr.getClientKey(id)).Err(); err != nil {
		if safeErr, ok := ccr.handleError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to delete client from cache: %w", err)
	}

	return nil
}

func (ccr *clientCacheRepository) getClientKey(id uuid.UUID) string {
	return fmt.Sprintf("client:%s", id.String())
}

func (ccr *clientCacheRepository) handleError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, redis.Nil):
		return client.ErrClientNotFound, true
	}

	return err, false
}
