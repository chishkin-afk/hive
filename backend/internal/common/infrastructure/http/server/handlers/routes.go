package handlers

import (
	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/gin-gonic/gin"
)

func New(cfg *config.Config,
	authService authService,
	offerService offerService,
	clientService clientService,
	scheduleService scheduleService,
	bookingService bookingService,
	paymentService paymentService,
	mws ...gin.HandlerFunc) *gin.Engine {
	router := getRouter(cfg)
	router.Use(mws...)

	funcs := HandlerFuncs{
		authService:     authService,
		offerService:    offerService,
		clientService:   clientService,
		scheduleService: scheduleService,
		bookingService:  bookingService,
		paymentService:  paymentService,
	}

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			// TODO: add method to fill tg chat's id
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

			clients := v1.Group("/clients")
			{
				clients.POST("/client", funcs.UpsertClient())
				clients.GET("/client/:id", funcs.GetClientByID())
				clients.GET("/clients/offer/:id", funcs.ListClientsByOfferID()) // ?limit=&page=
				clients.GET("/clients/user", funcs.ListClientsByUserID())
			}

			schedules := v1.Group("/schedules")
			{
				schedules.POST("/schedule", funcs.CreateSchedule())
				schedules.PATCH("/schedule", funcs.UpdateSchedule())
				schedules.GET("/schedule/:userID/d/:dayOfWeek", funcs.GetScheduleByUserID())
				schedules.GET("/schedules/user/:userID", funcs.GetSchedulesByUserID())
				schedules.DELETE("/schedule/:id", funcs.DeleteSchedule())
			}

			bookings := v1.Group("/bookings")
			{
				bookings.POST("/booking", funcs.CreateBooking())
				bookings.PATCH("/booking", funcs.UpdateBooking())
				bookings.GET("/booking/:id", funcs.GetBookingByID())
				bookings.GET("/bookings", funcs.ListBookings())       // ?limit=&page=&offer_id=?
				bookings.GET("/slots/:offerID", funcs.GetFreeSlots()) // ?date=2026-06-18
			}

			payments := v1.Group("/payments")
			{
				payments.POST("/payment", funcs.CreatePayment())
				payments.GET("/payments", funcs.ListPayments()) // ?page=&limit=
				payments.GET("/payment/:id", funcs.GetPaymentByID())
				payments.GET("/rates", funcs.ListAllRates())
				payments.GET("/rate/:id", funcs.GetRateByID())
				payments.POST("/yookassa", funcs.YookassaAction())
				payments.PATCH("/take-trial", funcs.TakeTrial())
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
