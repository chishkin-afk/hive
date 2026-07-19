package handlers

import (
	"github.com/devathh/hive/frontend/internal/infrastructure/config"
	"github.com/gin-gonic/gin"
)

const (
	EnvDev   = "dev"
	EnvProd  = "prod"
	EnvLocal = "local"
)

func New(cfg *config.Config, fs frontendService, mws ...gin.HandlerFunc) *gin.Engine {
	router := getRouter(cfg.App.Env)
	router.Use(mws...)

	router.LoadHTMLGlob("internal/web/templates/**/*.html")
	router.Static("/static", "./internal/web/static")

	handlers := handlers{
		fs: fs,
	}

	router.GET("/js/config.js", handlers.Config())

	router.GET("/", handlers.Landing())
	router.GET("/auth", handlers.Auth())
	router.GET("/appointment/:offerID", handlers.Appointment())

	return router
}

func getRouter(env string) *gin.Engine {
	switch env {
	case EnvDev, EnvLocal:
		router := gin.New()
		router.Use(gin.Recovery())
		return router
	}

	return gin.Default()
}
