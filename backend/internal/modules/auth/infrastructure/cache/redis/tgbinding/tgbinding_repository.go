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

type tgbindingCacheRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	cache Cache,
) *tgbindingCacheRepository {
	return &tgbindingCacheRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

func (tb *tgbindingCacheRepository) Save(ctx context.Context, tgBinding *tgbinding.TgBinding) error {
	tb.log.Debug("converting tg's binding model into bytes",
		slog.String("binding_id", tgBinding.ID().String()),
	)

	bytes, err := tgBindingToBytes(tgBinding)
	if err != nil {
		return fmt.Errorf("failed to convert tg's binding into bytes: %w", err)
	}

	tb.log.Debug("setting binding's bytes into cache",
		slog.String("binding_id", tgBinding.ID().String()),
	)

	if err := tb.cache.Set(ctx, tb.getBindingKey(tgBinding.ID()), bytes, tb.cfg.Cache.TgBindingTTL).Err(); err != nil {
		if safeErr, ok := tb.handleError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to set binding into cache: %w", err)
	}

	return nil
}

func (tb *tgbindingCacheRepository) GetByID(ctx context.Context, id uuid.UUID) (*tgbinding.TgBinding, error) {
	tb.log.Debug("getting bytes from cache of tg binding",
		slog.String("binding_id", id.String()),
	)

	bytes, err := tb.cache.Get(ctx, tb.getBindingKey(id)).Bytes()
	if err != nil {
		if safeErr, ok := tb.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get tgbinding cache: %w", err)
	}

	return bytesToTgBinding(bytes)
}

func (tb *tgbindingCacheRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tb.log.Debug("trying to delete binding by id",
		slog.String("binding_id", id.String()),
	)

	if err := tb.cache.Del(ctx, tb.getBindingKey(id)).Err(); err != nil {
		if safeErr, ok := tb.handleError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to delete tgbinding cache: %w", err)
	}

	return nil
}

func (tb *tgbindingCacheRepository) handleError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, redis.Nil):
		return tgbinding.ErrTgBindingNotFound, true
	}

	return err, false
}

func (tb *tgbindingCacheRepository) getBindingKey(id uuid.UUID) string {
	return fmt.Sprintf("tgbinding:%s", id.String())
}
