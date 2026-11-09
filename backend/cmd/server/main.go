package main

import (
	"log/slog"

	"github.com/chishkin-afk/hive/backend/internal/app"
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
		slog.Any("config", di.Config()),
	)
}
