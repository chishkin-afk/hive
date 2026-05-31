package app

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	redisconnect "github.com/devathh/hive/backend/internal/common/infrastructure/cache/redis"
	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/common/infrastructure/http/server"
	"github.com/devathh/hive/backend/internal/common/infrastructure/http/server/handlers"
	postgresconnect "github.com/devathh/hive/backend/internal/common/infrastructure/persistence/postgres"
	"github.com/devathh/hive/backend/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type DIContainer struct {
	cfg *config.Config
	log *slog.Logger

	pg    *sql.DB
	redis *redis.Client

	server  *server.Server
	handler *gin.Engine
}

func (di *DIContainer) Cfg() *config.Config {
	if di.cfg == nil {
		if err := godotenv.Load(".env"); err != nil {
			slog.Warn("failed to load .env", slog.String("error", err.Error()))
		}

		cfg, err := config.New(os.Getenv("APP_CONFIG_PATH"))
		if err != nil {
			slog.Error("failed to load config", slog.String("error", err.Error()))
			os.Exit(1)
		}

		di.cfg = cfg
	}

	return di.cfg
}

func (di *DIContainer) Log() *slog.Logger {
	if di.log == nil {
		di.log = log.New(di.Cfg())
	}

	return di.log
}

func (di *DIContainer) PG() *sql.DB {
	if di.pg == nil {
		pg, err := postgresconnect.Connect(di.Cfg())
		if err != nil {
			slog.Error("failed to load postgres", slog.String("error", err.Error()))
			os.Exit(1)
		}

		if err := postgresconnect.Migrate(di.Cfg(), pg); err != nil {
			slog.Error("failed to migrate postgres", slog.String("error", err.Error()))
			os.Exit(1)
		}

		Add(dep{
			name: "postgres",
			fc: func(_ context.Context) error {
				return pg.Close()
			},
		})

		di.pg = pg
	}

	return di.pg
}

func (di *DIContainer) Redis() *redis.Client {
	if di.redis == nil {
		client, err := redisconnect.Connect(di.Cfg())
		if err != nil {
			slog.Error("failed to connect redis", slog.String("error", err.Error()))
			os.Exit(1)
		}

		Add(dep{
			name: "redis",
			fc: func(_ context.Context) error {
				return client.Close()
			},
		})

		di.redis = client
	}

	return di.redis
}

func (di *DIContainer) Server() *server.Server {
	if di.server == nil {
		di.server = server.New(di.Cfg(), di.Handler())
	}

	return di.server
}

func (di *DIContainer) Handler() *gin.Engine {
	if di.handler == nil {
		di.handler = handlers.New(di.Cfg())
	}

	return di.handler
}
