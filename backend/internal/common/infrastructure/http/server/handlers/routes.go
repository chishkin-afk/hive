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
			auth := v1.Group("/auth")
			{
				auth.POST("/prepare", funcs.Prepare())
				auth.POST("/confirm", funcs.Confirm())
				auth.POST("/login", funcs.Login())
				auth.PATCH("/user", funcs.UpdateUser())
				auth.GET("/user", funcs.GetSelfUser())
			}

			offers := v1.Group("/offers")
			{
				offers.POST("/offer", funcs.CreateOffer())
				offers.PATCH("/offer", funcs.UpdateOffer())
				offers.DELETE("/offer/:id", funcs.DeleteOffer())
				offers.GET("/offer/:id", funcs.GetOfferByID())
				offers.GET("/offers/:authorID", funcs.ListOffersByAuthorID()) // ?limit=&page=
			}
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
