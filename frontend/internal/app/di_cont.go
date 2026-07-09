package app

import (
	"log/slog"
	"os"

	"github.com/devathh/hive/frontend/internal/infrastructure/application/services"
	"github.com/devathh/hive/frontend/internal/infrastructure/config"
	"github.com/devathh/hive/frontend/internal/infrastructure/http/server"
	"github.com/devathh/hive/frontend/internal/infrastructure/http/server/handlers"
	"github.com/devathh/hive/frontend/pkg/log"
	"github.com/gin-gonic/gin"
)

type DIContainer struct {
	cfg *config.Config
	log *slog.Logger

	handler *gin.Engine
	server  *server.Server

	frontendService *services.FrontendServices
}

func (di *DIContainer) Cfg() *config.Config {
	if di.cfg == nil {
		cfg, err := config.New(os.Getenv("APP_CONFIG_PATH"))
		if err != nil {
			slog.Error("failed to load config",
				slog.String("error", err.Error()),
			)

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

func (di *DIContainer) Handler() *gin.Engine {
	if di.handler == nil {
		di.handler = handlers.New(
			di.Cfg(),
			di.FrontendService(),
		)
	}

	return di.handler
}

func (di *DIContainer) Server() *server.Server {
	if di.server == nil {
		di.server = server.New(
			di.Cfg(),
			di.Handler(),
		)
	}

	return di.server
}

func (di *DIContainer) FrontendService() *services.FrontendServices {
	if di.frontendService == nil {
		di.frontendService = services.New(
			di.Cfg(),
			di.Log(),
		)
	}

	return di.frontendService
}
