package app

import (
	"log/slog"
	"os"

	"github.com/devathh/hive/frontend/internal/infrastructure/config"
	"github.com/devathh/hive/frontend/pkg/log"
)

type DIContainer struct {
	cfg *config.Config
	log *slog.Logger
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
