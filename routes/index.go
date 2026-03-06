package routes

import (
	"SureMFService/controller"
	"SureMFService/middleware"

	"github.com/gin-gonic/gin"
)

func Routes(router *gin.Engine) {
	base := router.Group("/sure-mf")

	// Health check
	base.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": 200, "msg": "ping success", "service": "SureMFService"})
	})

	fundRoutes(base)

	// Callbacks (no auth — called by FP via browser redirect or webhook)
	callbacks := base.Group("/callbacks")
	callbacks.GET("/payment", controller.PaymentCallback)
	callbacks.POST("/payment", controller.PaymentCallback)
	callbacks.GET("/mandate", controller.MandateCallback)
	callbacks.POST("/mandate", controller.MandateCallback)

	// All user-scoped routes are under /:uid/
	user := base.Group("/:uid")
	onboardingRoutes(user)
	orderRoutes(user)

	// Holdings (legacy OMS API)
	holdings := user.Group("/holdings")
	holdings.Use(middleware.AuthMiddleware())
	holdings.GET("", controller.GetHoldings)

	// Portfolio (v2 folios API)
	portfolio := user.Group("/portfolio")
	portfolio.Use(middleware.AuthMiddleware())
	portfolio.GET("", controller.GetPortfolio)
	portfolio.GET("/:id", controller.GetFolioDetail)

	// Mandates
	mandates := user.Group("/mandates")
	mandates.Use(middleware.AuthMiddleware())
	mandates.POST("", controller.CreateMandate)
	mandates.POST("/authorize", controller.AuthorizeMandate)
	mandates.GET("", controller.ListMandates)
	mandates.GET("/:id", controller.GetMandateStatus)
	mandates.POST("/:id/cancel", controller.CancelMandate)
}
