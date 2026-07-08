package main

import (
	"log/slog"

	"github.com/devathh/hive/frontend/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		slog.Warn("failed to load .env",
			slog.String("error", err.Error()),
		)
	}
	var di app.DIContainer
	di.Log().Info("di container was loaded")

	di.Server().Start()
}
