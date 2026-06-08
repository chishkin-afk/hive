package handlers

import (
	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config,
	authService authService,
	offerService offerService,
	mws ...gin.HandlerFunc) *gin.Engine {
	router := getRouter(cfg)
	router.Use(mws...)

	funcs := HandlerFuncs{
		authService:  authService,
		offerService: offerService,
	}

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.POST("/prepare", funcs.Prepare())
			v1.POST("/confirm", funcs.Confirm())
			v1.POST("/login", funcs.Login())
			v1.PATCH("/user", funcs.UpdateUser())
			v1.GET("/user", funcs.GetSelfUser())
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
