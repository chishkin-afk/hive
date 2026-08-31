package main

import (
	"log/slog"
	"os"

	"github.com/braunkc/hive-engage/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("failed to load env", slog.Any("err", err))
		os.Exit(1)
	}

	a := app.New()
	a.Run()
}
