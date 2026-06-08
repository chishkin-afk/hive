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
	"github.com/devathh/hive/backend/internal/common/infrastructure/http/server/middlewares"
	"github.com/devathh/hive/backend/internal/common/infrastructure/jwt"
	postgresconnect "github.com/devathh/hive/backend/internal/common/infrastructure/persistence/postgres"
	smtpinfra "github.com/devathh/hive/backend/internal/common/infrastructure/smtp"
	authservices "github.com/devathh/hive/backend/internal/modules/auth/application/services"
	userredis "github.com/devathh/hive/backend/internal/modules/auth/infrastructure/cache/redis/user"
	userpg "github.com/devathh/hive/backend/internal/modules/auth/infrastructure/persistence/postgres/user"
	offerservices "github.com/devathh/hive/backend/internal/modules/offers/application/services"
	offerredis "github.com/devathh/hive/backend/internal/modules/offers/infrastructure/cache/redis/offer"
	offerpg "github.com/devathh/hive/backend/internal/modules/offers/infrastructure/persistence/postgres/offer"
	"github.com/devathh/hive/backend/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type DIContainer struct {
	cfg *config.Config
	log *slog.Logger

	pg         *sql.DB
	redis      *redis.Client
	jwtManager *jwt.JWTManager

	server  *server.Server
	handler *gin.Engine

	authService  *authservices.AuthService
	offerService *offerservices.OfferService

	smtpClient *smtpinfra.SMTPClient
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
		di.handler = handlers.New(
			di.Cfg(),
			di.AuthService(),
			di.OfferService(),
			middlewares.NewAuthMiddleware(di.JWT(), map[string]bool{
				"GET:/api/v1/auth/user":           true,
				"PATCH:/api/v1/auth/user":         true,
				"POST:/api/v1/offers/offer":       true,
				"PATCH:/api/v1/offers/offer":      true,
				"DELETE:/api/v1/offers/offer/:id": true,
			}),
		)
	}

	return di.handler
}

func (di *DIContainer) JWT() *jwt.JWTManager {
	if di.jwtManager == nil {
		jwt, err := jwt.New(di.Cfg())
		if err != nil {
			slog.Error("failed to create jwt manager", slog.String("error", err.Error()))
			os.Exit(1)
		}

		di.jwtManager = jwt
	}

	return di.jwtManager
}

func (di *DIContainer) SMTP() *smtpinfra.SMTPClient {
	if di.smtpClient == nil {
		di.smtpClient = smtpinfra.New(di.Cfg())
	}

	return di.smtpClient
}

func (di *DIContainer) AuthService() *authservices.AuthService {
	if di.authService == nil {
		di.authService = authservices.New(
			di.Cfg(),
			di.Log(),
			userpg.New(
				di.Log(),
				di.PG(),
			),
			userredis.New(
				di.Cfg(),
				di.Log(),
				di.Redis(),
			),
			di.JWT(),
			di.SMTP(),
		)
	}

	return di.authService
}

func (di *DIContainer) OfferService() *offerservices.OfferService {
	if di.offerService == nil {
		di.offerService = offerservices.New(
			di.Cfg(),
			di.Log(),
			offerpg.New(di.Log(), di.PG()),
			offerredis.New(di.Cfg(), di.Log(), di.Redis()),
		)
	}

	return di.offerService
}
