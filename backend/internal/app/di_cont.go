package app

import (
	"log/slog"
	"os"

	"github.com/chishkin-afk/hive/backend/internal/common/config"
	"github.com/chishkin-afk/hive/backend/pkg/log"
)

type DI struct {
	cfg *config.Config
	log *slog.Logger
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
			os.Exit(1)
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
