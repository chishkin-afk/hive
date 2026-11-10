package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/chishkin-afk/hive/backend/internal/app"
	pgsdk "github.com/chishkin-afk/hive/backend/internal/common/persistence/postgres"
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

	if err := pgsdk.Migrate(&di.Config().Persistence); err != nil {
		di.Log().Error("can't up migrations",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	di.Log().Info("all migrations have been applied")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.CloseAll(ctx); err != nil {
		di.Log().Error("some functions weren't closed",
			slog.String("error", err.Error()),
		)
	}
}
