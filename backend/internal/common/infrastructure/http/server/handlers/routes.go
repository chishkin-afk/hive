package handlers

import (
	_ "github.com/devathh/hive/backend/docs"
	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New(cfg *config.Config,
	authService authService,
	bookingService bookingService,
	offerService offerService,
	clientService clientService,
	scheduleService scheduleService,
	paymentService paymentService,
	mws ...gin.HandlerFunc) *gin.Engine {
	router := getRouter(cfg)
	router.Use(mws...)

	funcs := HandlerFuncs{
		cfg:             cfg,
		authService:     authService,
		bookingService:  bookingService,
		offerService:    offerService,
		clientService:   clientService,
		scheduleService: scheduleService,
		paymentService:  paymentService,
	}

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			auth := v1.Group("/auth")
			{
				auth.POST("/prepare", funcs.Prepare())
				auth.POST("/confirm", funcs.Confirm())
				auth.POST("/resend", funcs.ResendCode())
				auth.POST("/login", funcs.Login())
				auth.PATCH("/user", funcs.UpdateUser())
				auth.GET("/user", funcs.GetSelfUser())
				auth.GET("/tgbinding", funcs.GetTgBindingURL())
				auth.POST("/tgbinding", funcs.BindTG())
				// auth.POST("/lookup", funcs.Lookup())
				auth.POST("/forget-password", funcs.ForgetPassword())
				auth.PATCH("/password", funcs.ChangePassword())
				auth.PATCH("/unbind-tg", funcs.UnbindTg())
			}

			offers := v1.Group("/offers")
			{
				offers.POST("/offer", funcs.CreateOffer())
				offers.PATCH("/offer", funcs.UpdateOffer())
				offers.DELETE("/offer/:id", funcs.DeleteOffer())
				offers.GET("/offer/:id", funcs.GetOfferByID())
				offers.GET("/user/:authorID", funcs.ListOffersByAuthorID()) // ?limit=&page=
			}

			clients := v1.Group("/clients")
			{
				clients.POST("/client", funcs.UpsertClient())
				clients.GET("/client/:id", funcs.GetClientByID())
				clients.GET("/offer/:id", funcs.ListClientsByOfferID()) // ?limit=&page=
				clients.GET("/user", funcs.ListClientsByUserID())       // ?limit=&page=
			}

			schedules := v1.Group("/schedules")
			{
				schedules.POST("/schedule", funcs.CreateSchedule())
				schedules.PATCH("/schedule", funcs.UpdateSchedule())
				schedules.GET("/schedule/:userID/d/:dayOfWeek", funcs.GetScheduleByUserID())
				schedules.GET("/user/:userID", funcs.GetSchedulesByUserID())
				schedules.DELETE("/schedule/:id", funcs.DeleteSchedule())
			}

			bookings := v1.Group("/bookings")
			{
				bookings.POST("/booking", funcs.CreateBooking())
				bookings.PATCH("/booking", funcs.UpdateBooking())
				bookings.GET("/booking/:id", funcs.GetBookingByID())
				bookings.GET("/bookings", funcs.ListBookings())       // ?limit=&page=&offer_id=?
				bookings.GET("/slots/:offerID", funcs.GetFreeSlots()) // ?date=2026-06-18
				bookings.GET("/days/:offerID", funcs.GetAllDays())    // ?from=2026-06-18&to=2026-06-19
			}

			payments := v1.Group("/payments")
			{
				payments.POST("/payment", funcs.CreatePayment())
				payments.GET("/user", funcs.ListPayments()) // ?page=&limit=
				payments.GET("/payment/:id", funcs.GetPaymentByID())
				payments.GET("/rates", funcs.ListAllRates())
				payments.GET("/rate/:id", funcs.GetRateByID())
				payments.POST("/yookassa", funcs.YookassaAction())
				payments.PATCH("/take-trial", funcs.TakeTrial())
			}

			v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
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

	return gin.Default()
}
