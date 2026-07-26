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

type clientRepository struct {
	cfg   *config.Config
	log   *slog.Logger
	cache Cache
}

func New(cfg *config.Config, log *slog.Logger, cache Cache) *clientRepository {
	return &clientRepository{
		cfg:   cfg,
		log:   log,
		cache: cache,
	}
}

// Save is a method that saves
// the client to the cache as bytes
// otherwise, an error occurs.
func (cr *clientRepository) Save(ctx context.Context, client *client.Client) error {
	cr.log.Debug("saving client into cache",
		slog.String("client_id", client.ID().String()),
	)

	if err := cr.save(ctx, cr.cache, client); err != nil {
		return fmt.Errorf("failed to save client: %w",
			handleError(err),
		)
	}

	return nil
}

func (cr *clientRepository) save(ctx context.Context, cache Cache, client *client.Client) error {
	bytes, err := clientToBytes(client)
	if err != nil {
		return err
	}

	key := getClientKey(client.ID())
	return cache.Set(ctx, key, bytes, cr.cfg.Cache.ClientTTL).Err()
}

// GetByID is a method that returns
// a client from the cache,
// or otherwise an error.
func (cr *clientRepository) GetByID(ctx context.Context, id uuid.UUID) (*client.Client, error) {
	cr.log.Debug("getting client from cache",
		slog.String("client_id", id.String()),
	)

	client, err := cr.get(ctx, cr.cache, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get client from cache: %w",
			handleError(err),
		)
	}

	return client, nil
}

func (cr *clientRepository) get(ctx context.Context, cache Cache, id uuid.UUID) (*client.Client, error) {
	bytes, err := cache.Get(ctx, getClientKey(id)).Bytes()
	if err != nil {
		return nil, err
	}

	return bytesToClient(bytes)
}

// Delete`is a method that removes
// a client from the cache
// otherwise, it returns an error.
func (cr *clientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	cr.log.Debug("deleting client from cache",
		slog.String("client_id", id.String()),
	)

	if err := cr.delete(ctx, cr.cache, id); err != nil {
		return fmt.Errorf("failed to delete client from cache: %w",
			handleError(err),
		)
	}

	return nil
}

func (cr *clientRepository) delete(ctx context.Context, cache Cache, id uuid.UUID) error {
	return cache.Del(ctx, getClientKey(id)).Err()
}

func getClientKey(id uuid.UUID) string {
	return fmt.Sprintf("client:%s", id.String())
}

func handleError(err error) error {
	if errors.Is(err, redis.Nil) {
		return client.ErrClientNotFound
	}

	return err
}
