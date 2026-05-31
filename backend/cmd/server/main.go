package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/devathh/hive/backend/internal/app"
)

func main() {
	var di app.DIContainer
	di.Log().Info("di container was loaded")

	go func() {
		di.Log().Info("server is running", slog.String("addr", di.Cfg().Server.HTTP.Addr))
		if err := di.Server().Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			di.Log().Error("failed to run server", slog.String("error", err.Error()))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	stop()

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := di.Server().Shutdown(ctxTimeout); err != nil {
		di.Log().Error("failed to shutdown server", slog.String("error", err.Error()))
	}
	di.Log().Info("server was closed")

	if err := app.Close(ctxTimeout); err != nil {
		di.Log().Error("failed to close all the connections", slog.String("error", err.Error()))
	}
	di.Log().Info("all the connections was closed")
}
