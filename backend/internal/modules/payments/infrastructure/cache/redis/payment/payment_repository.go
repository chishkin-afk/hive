package paymentredis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/payments/domain/payment"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, duration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, key ...string) *redis.IntCmd
}

type paymentRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *paymentRepository {
	return &paymentRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

// Save stores the provided payment in the cache.
//
// The cached data expires after the configured TTL. An error is returned if
// the payment cannot be serialized or the cache write operation fails.
func (pr *paymentRepository) Save(ctx context.Context, payment *payment.Payment) error {
	pr.log.Debug("saving payment into cache",
		slog.String("payment_id", payment.ID().String()),
	)

	if err := pr.save(ctx, pr.cache, payment); err != nil {
		return fmt.Errorf("failed to save payment into cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (pr *paymentRepository) save(ctx context.Context, cache Cache, payment *payment.Payment) error {
	bytes, err := paymentToBytes(payment)
	if err != nil {
		return err
	}

	key := getPaymentKey(payment.ID())
	return cache.Set(ctx, key, bytes, pr.cfg.Cache.PaymentTTL).Err()
}

// Delete removes the cached payment identified by id.
//
// It returns an error if the cache delete operation fails.
func (pr *paymentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	pr.log.Debug("deleting payment from cache",
		slog.String("payment_id", id.String()),
	)

	if err := pr.delete(ctx, pr.cache, id); err != nil {
		return fmt.Errorf("failed to delete payment from cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (pr *paymentRepository) delete(ctx context.Context, cache Cache, id uuid.UUID) error {
	return cache.Del(ctx, getPaymentKey(id)).Err()
}

// Get retrieves the cached payment identified by id.
//
// It returns the cached payment or an error if the cache lookup or

func (pr *paymentRepository) Get(ctx context.Context, id uuid.UUID) (*payment.Payment, error) {
	pr.log.Debug("getting payment from cache",
		slog.String("payment_id", id.String()),
	)

	payment, err := pr.get(ctx, pr.cache, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment from cache: %w",
			handleError(err),
		)
	}

	return payment, nil
}

func (pr *paymentRepository) get(ctx context.Context, cache Cache, id uuid.UUID) (*payment.Payment, error) {
	bytes, err := cache.Get(ctx, getPaymentKey(id)).Bytes()
	if err != nil {
		return nil, err
	}

	return bytesToPayment(bytes)
}

func handleError(err error) error {
	if errors.Is(err, redis.Nil) {
		return payment.ErrPaymentNotFound
	}

	return err
}

func getPaymentKey(id uuid.UUID) string {
	return fmt.Sprintf("payment:%s", id.String())
}
