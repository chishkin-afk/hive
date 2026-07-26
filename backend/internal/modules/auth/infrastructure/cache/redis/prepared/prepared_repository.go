package preparedredis

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
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, duration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, key ...string) *redis.IntCmd
}

type preparedRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *preparedRepository {
	return &preparedRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

// Prepare is a method that creates (prepares) a user
// in the cache based on their email.
// Before doing so, it checks whether the user exists.
func (pr *preparedRepository) Prepare(
	ctx context.Context,
	preparingUser *user.User,
	generatedPassword string,
	code int,
) error {
	pr.log.Debug("checking for prepared existing",
		slog.String("prepared_id", preparingUser.ID().String()),
	)

	if err := pr.checkExisting(ctx, pr.cache, preparingUser.ConfirmedEmail()); err == nil {
		return user.ErrUserAlreadyExists
	}

	pr.log.Debug("setting prepared into cache",
		slog.String("prepared_id", preparingUser.ID().String()),
	)
	if err := pr.setPrepared(
		ctx,
		pr.cache,
		preparingUser,
		generatedPassword,
		code,
	); err != nil {
		return fmt.Errorf("failed to save prepared into cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (pr *preparedRepository) setPrepared(
	ctx context.Context,
	cache Cache,
	preparingUser *user.User,
	generatedPassword string,
	code int,
) error {
	bytes, err := preparedToBytes(
		preparingUser,
		generatedPassword,
		code,
	)
	if err != nil {
		return err
	}

	key := getPreparedKey(preparingUser.ConfirmedEmail())
	return cache.Set(ctx, key, bytes, pr.cfg.Cache.PreparingTTL).Err()
}

func (pr *preparedRepository) checkExisting(ctx context.Context, cache Cache, email user.Email) error {
	key := getPreparedKey(email)
	return cache.Get(ctx, key).Err()
}

// GetPrepared is a method that returns a prepared user,
// or an error otherwise.
func (pr *preparedRepository) GetPrepared(ctx context.Context, email user.Email) (*user.User, string, int, error) {
	pr.log.Debug("getting prepared from cache by id",
		slog.String("email", email.String()),
	)

	user, password, code, err := pr.getPrepared(ctx, pr.cache, email)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to get prepared from cache: %w",
			handleError(err),
		)
	}

	return user, password, code, err
}

func (pr *preparedRepository) getPrepared(ctx context.Context, cache Cache, email user.Email) (*user.User, string, int, error) {
	bytes, err := cache.Get(ctx, getPreparedKey(email)).Bytes()
	if err != nil {
		return nil, "", 0, err
	}

	record, err := bytesToPrepared(bytes)
	if err != nil {
		return nil, "", 0, err
	}

	return preparedToUser(record), record.GeneratedPassword, record.Code, nil
}

// DelPrepared is a method that removes
// a prepared user and their data
// from the cache based on their email.
func (pr *preparedRepository) DelPrepared(ctx context.Context, email user.Email) error {
	pr.log.Debug("deleting prepared from cache",
		slog.String("email", email.String()),
	)

	if err := pr.delPrepared(ctx, pr.cache, email); err != nil {
		return fmt.Errorf("failed to delete prepared from cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (pr *preparedRepository) delPrepared(ctx context.Context, cache Cache, email user.Email) error {
	return cache.Del(ctx, getPreparedKey(email)).Err()
}

// SetLastSend is a method that records
// the time of the user's last request
// in the cache.
func (pr *preparedRepository) SetLastSend(ctx context.Context, email user.Email, at time.Time) error {
	pr.log.Debug("setting last send into cache",
		slog.String("email", email.String()),
	)

	if err := pr.setLastSend(ctx, pr.cache, email, at); err != nil {
		return fmt.Errorf("failed to set last send into cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (pr *preparedRepository) setLastSend(ctx context.Context, cache Cache, email user.Email, at time.Time) error {
	bytes, err := lastSendToBytes(at)
	if err != nil {
		return err
	}

	key := getLastSendKey(email)
	return cache.Set(ctx, key, bytes, pr.cfg.Service.SendingDuration).Err()
}

// GetLastSend is a method that returns the timestamp
// of the last user code transmission,
// or an error otherwise.
func (pr *preparedRepository) GetLastSend(ctx context.Context, email user.Email) (time.Time, error) {
	pr.log.Debug("getting last send of user by email",
		slog.String("email", email.String()),
	)

	at, err := pr.getLastSend(ctx, pr.cache, email)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last send of user: %w",
			handleError(err),
		)
	}

	return at, nil
}

func (pr *preparedRepository) getLastSend(ctx context.Context, cache Cache, email user.Email) (time.Time, error) {
	bytes, err := cache.Get(ctx, getLastSendKey(email)).Bytes()
	if err != nil {
		return time.Time{}, err
	}

	return bytesToLastSendRecord(bytes)
}

func handleError(err error) error {
	if errors.Is(err, redis.Nil) {
		return user.ErrUserNotFound
	}

	return err
}

func getPreparedKey(email user.Email) string {
	hash := sha256.Sum256([]byte(email.String()))
	return fmt.Sprintf("prepared:%s", hex.EncodeToString(hash[:]))
}

func getLastSendKey(email user.Email) string {
	hash := sha256.Sum256([]byte(email.String()))
	return fmt.Sprintf("last_send:%s", hex.EncodeToString(hash[:]))
}
