package tgbindingredis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/auth/domain/tgbinding"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, duration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, key ...string) *redis.IntCmd
}

type tgbindingRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	cache Cache,
) *tgbindingRepository {
	return &tgbindingRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

// Set establishes the structure for
// binding a Telegram account in the cache
// otherwise, it returns an error.
func (tr *tgbindingRepository) Set(ctx context.Context, tgBinding *tgbinding.TgBinding) error {
	tr.log.Debug("setting tg's binding in cache",
		slog.String("tgbinding_id", tgBinding.ID().String()),
	)

	if err := tr.set(ctx, tr.cache, tgBinding); err != nil {
		return fmt.Errorf("failed to set tg's binding cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (tr *tgbindingRepository) set(ctx context.Context, cache Cache, tgBinding *tgbinding.TgBinding) error {
	bytes, err := tgBindingToBytes(tgBinding)
	if err != nil {
		return err
	}

	key := getBindingKey(tgBinding.ID())
	return cache.Set(ctx, key, bytes, tr.cfg.Cache.TgBindingTTL).Err()
}

// Get returns the structure for
// the Telegram bind (if it exists),
// or an error otherwise.
func (tr *tgbindingRepository) Get(ctx context.Context, id uuid.UUID) (*tgbinding.TgBinding, error) {
	tr.log.Debug("getting tg's binding from cache by id",
		slog.String("tgbinding_id", id.String()),
	)

	tgBinding, err := tr.get(ctx, tr.cache, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tgbinding from cache: %w",
			handleError(err),
		)
	}

	return tgBinding, nil
}

func (tr *tgbindingRepository) get(ctx context.Context, cache Cache, id uuid.UUID) (*tgbinding.TgBinding, error) {
	bytes, err := cache.Get(ctx, getBindingKey(id)).Bytes()
	if err != nil {
		return nil, err
	}

	return bytesToTgBinding(bytes)
}

func (tr *tgbindingRepository) Del(ctx context.Context, id uuid.UUID) error {
	tr.log.Debug("deleting tg's binding from cache by id",
		slog.String("tgbinding_id", id.String()),
	)

	if err := tr.del(ctx, tr.cache, id); err != nil {
		return fmt.Errorf("failed to delete tg's binding from cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (tr *tgbindingRepository) del(ctx context.Context, cache Cache, id uuid.UUID) error {
	return cache.Del(ctx, getBindingKey(id)).Err()
}

func handleError(err error) error {
	if errors.Is(err, redis.Nil) {
		return tgbinding.ErrTgBindingNotFound
	}

	return err
}

func getBindingKey(id uuid.UUID) string {
	return fmt.Sprintf("tgbinding:%s", id.String())
}
