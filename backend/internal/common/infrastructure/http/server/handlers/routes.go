package handlers

import (
	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config) *gin.Engine {
	router := getRouter(cfg)
	funcs := HandlerFuncs{}

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			_ = v1
			_ = funcs
		}
	}

	return router
}

func getRouter(cfg *config.Config) *gin.Engine {
	if cfg.App.Env == "prod" {
		router := gin.New()
		router.Use(gin.Recovery())
		return router
	}

	return gin.New()
}
