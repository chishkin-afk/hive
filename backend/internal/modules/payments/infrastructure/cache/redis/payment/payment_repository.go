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

type paymentCacheRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *paymentCacheRepository {
	return &paymentCacheRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

func (pcr *paymentCacheRepository) SavePayment(ctx context.Context, payment *payment.Payment) error {
	pcr.log.Debug("converting payment into bytes",
		slog.String("payment_id", payment.ID().String()),
	)

	bytes, err := paymentToBytes(payment)
	if err != nil {
		return err
	}

	pcr.log.Debug("saving payment's bytes into cache",
		slog.String("payment_id", payment.ID().String()),
	)

	if err := pcr.cache.Set(ctx, pcr.getPaymentKey(payment.ID()), bytes, pcr.cfg.Cache.PaymentTTL).Err(); err != nil {
		if safeErr, ok := pcr.handleError(err); ok {
			return safeErr
		}

		return fmt.Errorf("failed to save payment into cache: %w", err)
	}

	return nil
}

func (pcr *paymentCacheRepository) DeletePayment(ctx context.Context, id uuid.UUID) error {
	pcr.log.Debug("deleting payment from cache",
		slog.String("payment_id", id.String()),
	)

	if err := pcr.cache.Del(ctx, pcr.getPaymentKey(id)).Err(); err != nil {
		if safeErr, ok := pcr.handleError(err); ok {
			return safeErr
		}
	}

	return nil
}

func (pcr *paymentCacheRepository) GetPayment(ctx context.Context, id uuid.UUID) (*payment.Payment, error) {
	pcr.log.Debug("get bytes of payment from cache",
		slog.String("payment_id", id.String()),
	)

	bytes, err := pcr.cache.Get(ctx, pcr.getPaymentKey(id)).Bytes()
	if err != nil {
		if safeErr, ok := pcr.handleError(err); ok {
			return nil, safeErr
		}

		return nil, fmt.Errorf("failed to get payment from cache: %w", err)
	}

	pcr.log.Debug("converting bytes into payment",
		slog.String("payment_id", id.String()),
	)

	return bytesToPayment(bytes)
}

func (pcr *paymentCacheRepository) handleError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.Canceled),
		errors.Is(err, context.DeadlineExceeded):
		return err, true
	case errors.Is(err, redis.Nil):
		return payment.ErrPaymentNotFound, true
	}

	return err, false
}

func (pcr *paymentCacheRepository) getPaymentKey(id uuid.UUID) string {
	return fmt.Sprintf("payment:%s", id.String())
}
