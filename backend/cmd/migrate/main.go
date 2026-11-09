package main

import (
	"log/slog"
	"os"

	"github.com/chishkin-afk/hive/backend/internal/app"
	"github.com/chishkin-afk/hive/backend/internal/common/persistence/postgres"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		slog.Warn("can't load .env",
			slog.String("error", err.Error()),
		)
	}

	var di app.DI
	di.Log().Info("di was loaded",
		slog.String("name", di.Config().App.Name),
		slog.String("version", di.Config().App.Version),
		slog.String("env_mode", di.Config().App.Env),
	)

	if err := postgres.Migrate(&di.Config().Persistence); err != nil {
		di.Log().Error("can't up migrations",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	di.Log().Info("all migrations have been applied")
}
