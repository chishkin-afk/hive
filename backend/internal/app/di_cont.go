package app

import (
	"context"
	"log/slog"
	"os"

	redissdk "github.com/chishkin-afk/hive/backend/internal/common/cache/redis"
	"github.com/chishkin-afk/hive/backend/internal/common/config"
	pgsdk "github.com/chishkin-afk/hive/backend/internal/common/persistence/postgres"
	"github.com/chishkin-afk/hive/backend/pkg/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type DI struct {
	cfg *config.Config
	log *slog.Logger

	db    *pgxpool.Pool
	cache *redis.Client
}

// Calling config means that .env file was loaded
func (di *DI) Config() *config.Config {
	if di.cfg == nil {
		path := os.Getenv("APP_CONFIG_PATH")
		cfg, err := config.New(path)
		if err != nil {
			slog.Error("can't load config",
				slog.String("error", err.Error()),
				slog.String("path", path),
			)
			panic(err)
		}

		di.cfg = cfg
	}

	return di.cfg
}

func (di *DI) Log() *slog.Logger {
	if di.log == nil {
		di.log = log.New(di.Config().App.Env)
	}

	return di.log
}

func (di *DI) DB() *pgxpool.Pool {
	if di.db == nil {
		db, err := pgsdk.Connect(&di.Config().Persistence)
		if err != nil {
			slog.Error("can't connect db",
				slog.String("error", err.Error()),
			)
			panic(err)
		}

		if err := Add(func(ctx context.Context) error {
			done := make(chan struct{}, 1)
			go func() {
				defer close(done)
				db.Close()
			}()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-done:
				return nil
			}
		}); err != nil {
			slog.Warn("can't create fn to close db",
				slog.String("error", err.Error()),
			)
		}

		di.db = db
	}

	return di.db
}

func (di *DI) Cache() *redis.Client {
	if di.cache == nil {
		client, err := redissdk.Connect(&di.Config().Cache)
		if err != nil {
			slog.Error("can't connect cache",
				slog.String("error", err.Error()),
			)
			panic(err)
		}

		if err := Add(func(ctx context.Context) error {
			errCh := make(chan error, 1)

			go func() {
				defer close(errCh)
				if err := client.Close(); err != nil {
					errCh <- err
				}
			}()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case err := <-errCh:
				return err
			}
		}); err != nil {
			slog.Warn("can't create fn to close cache",
				slog.String("error", err.Error()),
			)
		}
	}

	return di.cache
}
