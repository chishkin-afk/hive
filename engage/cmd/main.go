package main

import (
	"log/slog"

	"github.com/braunkc/hive-engage/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("failed to load env", slog.Any("err", err))
	}

	a := app.New()
	a.Run()
}
