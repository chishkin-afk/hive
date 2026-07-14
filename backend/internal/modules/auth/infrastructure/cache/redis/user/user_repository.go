package userredis

import (
	"context"
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

type userRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *userRepository {
	return &userRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

// Set is a function used to save a user
// to the cache based on their ID.
func (ur *userRepository) Set(ctx context.Context, user *user.User) error {
	ur.log.Debug("setting user into cache",
		slog.String("user_id", user.ID().String()),
	)

	if err := ur.set(ctx, ur.cache, user); err != nil {
		return fmt.Errorf("failed to set user into cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (ur *userRepository) set(ctx context.Context, cache Cache, user *user.User) error {
	bytes, err := domainToBytes(user)
	if err != nil {
		return err
	}

	return cache.Set(ctx, getUserKey(user.ID()), bytes, ur.cfg.Cache.UserTTL).Err()
}

// Get is a function that returns a user
// from the cache by their ID,
// or otherwise returns an error.
func (ur *userRepository) Get(ctx context.Context, id uuid.UUID) (*user.User, error) {
	ur.log.Debug("getting user from cache by id",
		slog.String("user_id", id.String()),
	)

	user, err := ur.get(ctx, ur.cache, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from cache: %w",
			handleError(err),
		)
	}

	return user, nil
}

func (ur *userRepository) get(ctx context.Context, cache Cache, id uuid.UUID) (*user.User, error) {
	bytes, err := cache.Get(ctx, getUserKey(id)).Bytes()
	if err != nil {
		return nil, err
	}

	return bytesToDomain(bytes)
}

// Del is a function that clears a user's cache
// based on their ID; otherwise,
// it returns an error
// (except when the user does not exist).
func (ur *userRepository) Del(ctx context.Context, id uuid.UUID) error {
	ur.log.Debug("deleting user's cache by id",
		slog.String("user_id", id.String()),
	)

	if err := ur.del(ctx, ur.cache, id); err != nil {
		return fmt.Errorf("failed to delete user's cache by id: %w",
			handleError(err),
		)
	}

	return nil
}

func (ur *userRepository) del(ctx context.Context, cache Cache, id uuid.UUID) error {
	return cache.Del(ctx, getUserKey(id)).Err()
}

func handleError(err error) error {
	if errors.Is(err, redis.Nil) {
		return user.ErrUserNotFound
	}

	return err
}

func getUserKey(id uuid.UUID) string {
	return fmt.Sprintf("user:%s", id.String())
}
