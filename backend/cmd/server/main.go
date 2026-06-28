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

// @title           HIVE API
// @version         1.0
// @description     HIVE API is a backend server for the hive crm that allows you to record customers, process them, and interact with offers

// @host      localhost:8090
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "<your JWT token>"

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

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
	di.PG()

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
