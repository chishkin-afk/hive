package userredis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/auth/domain/user"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, duration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, key ...string) *redis.IntCmd
}

type userCacheRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *userCacheRepository {
	return &userCacheRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

func (ucr *userCacheRepository) Set(ctx context.Context, user *user.User) error {
	ucr.log.Debug("start to convert domain into bytes", slog.String("user_id", user.ID().String()))
	bytes, err := domainToRecord(user)
	if err != nil {
		return fmt.Errorf("failed to convert domain into bytes: %w", err)
	}

	ucr.log.Debug("setting the user's cache by id", slog.String("user_id", user.ID().String()))
	if err := ucr.cache.Set(ctx, ucr.getUserKey(user.ID()), bytes, ucr.cfg.Cache.UserTTL).Err(); err != nil {
		if safeErr, ok := ucr.normalizeError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to set user's cache: %w", err)
	}

	return nil
}

func (ucr *userCacheRepository) Get(ctx context.Context, id uuid.UUID) (*user.User, error) {
	ucr.log.Debug("get the user's cache by id", slog.String("user_id", id.String()))
	bytes, err := ucr.cache.Get(ctx, ucr.getUserKey(id)).Bytes()
	if err != nil {
		if safeErr, ok := ucr.normalizeError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get user's cache: %w", err)
	}

	return recordToDomain(bytes)
}

func (ucr *userCacheRepository) Del(ctx context.Context, id uuid.UUID) error {
	ucr.log.Debug("deleting the user's cache", slog.String("user_id", id.String()))
	if err := ucr.cache.Del(ctx, ucr.getUserKey(id)).Err(); err != nil {
		if safeErr, ok := ucr.normalizeError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to delete user's cache: %w", err)
	}

	return nil
}

func (ucr *userCacheRepository) Prepare(ctx context.Context, user *user.User, generatedPassword string, code int) error {
	ucr.log.Debug("converting domain into prepared", slog.Int("code", code))
	preparedRecord, err := domainToPrepared(user, generatedPassword)
	if err != nil {
		return fmt.Errorf("failed to convert domain into prepared: %w", err)
	}

	ucr.log.Debug("setting the prepared record into cache", slog.Int("code", code))
	if err := ucr.cache.Set(ctx, ucr.getPreparedKey(code, user.ID()), preparedRecord, ucr.cfg.Cache.PreparingTTL).Err(); err != nil {
		if safeErr, ok := ucr.normalizeError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to set prepared record: %w", err)
	}

	return nil
}

func (ucr *userCacheRepository) GetPrepared(ctx context.Context, code int, id uuid.UUID) (*user.User, string, error) {
	ucr.log.Debug("get the prepared user from cache", slog.Int("code", code))
	bytes, err := ucr.cache.Get(ctx, ucr.getPreparedKey(code, id)).Bytes()
	if err != nil {
		if safeErr, ok := ucr.normalizeError(err); ok {
			return nil, "", safeErr
		}

		return nil, "", fmt.Errorf("failed to get prepared record: %w", err)
	}

	return preparedToDomain(bytes)
}

func (ucr *userCacheRepository) normalizeError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, redis.Nil):
		return user.ErrUserNotFound, true
	}

	return err, false
}

func (ucr *userCacheRepository) getPreparedKey(code int, id uuid.UUID) string {
	hash := sha256.Sum256(fmt.Appendf(nil, "%d:%s", code, id.String()))
	return fmt.Sprintf("code:%s", hex.EncodeToString(hash[:]))
}

func (ucr *userCacheRepository) getUserKey(id uuid.UUID) string {
	return fmt.Sprintf("user:%s", id.String())
}
