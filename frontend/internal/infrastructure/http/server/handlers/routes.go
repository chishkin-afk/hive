package handlers

import (
	"github.com/devathh/hive/frontend/internal/infrastructure/config"
	"github.com/devathh/hive/frontend/internal/infrastructure/http/server/middleware"
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

	// --- Публичные страницы (без авторизации) ---
	router.GET("/", handlers.Landing())
	router.GET("/auth", middleware.RedirectIfAuthenticated(), handlers.Auth())
	router.GET("/appointment/:offerID", handlers.Appointment())
	router.GET("/reset", handlers.Reset())
	router.GET("/terms", handlers.Terms())
	router.GET("/privacy", handlers.Privacy())
	router.GET("/offer", handlers.Offer())

	// --- Защищённые страницы (требуется токен) ---
	protected := router.Group("", middleware.AuthRequired())
	{
		protected.GET("/services", handlers.Services())
		protected.GET("/profile", handlers.Profile())
		protected.GET("/topup", handlers.Topup())
		protected.GET("/bookings", handlers.Bookings())
	}

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
