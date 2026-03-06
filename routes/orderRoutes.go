package routes

import (
	"SureMFService/controller"
	"SureMFService/middleware"

	"github.com/gin-gonic/gin"
)

func orderRoutes(rg *gin.RouterGroup) {
	orders := rg.Group("/orders")
	orders.Use(middleware.AuthMiddleware())
	{
		orders.GET("", controller.GetOrders)

		// Lumpsum purchase
		orders.POST("/purchase", controller.PlacePurchaseOrder)
		orders.POST("/:id/confirm-otp", controller.ConfirmOTP)
		orders.PATCH("/:id/consent", controller.UpdateConsent)
		orders.POST("/:id/payment", controller.CreatePayment)
		orders.PATCH("/:id/confirm", controller.ConfirmPurchaseState)
		orders.GET("/:id/status", controller.GetPurchaseStatus)

		// SIP lifecycle
		orders.POST("/sip", controller.PlaceSIPOrder)
		orders.GET("/sips", controller.ListSIPs)
		orders.GET("/sips/:id", controller.GetSIPDetail)
		orders.PATCH("/sips/:id/confirm", controller.ConfirmSIP)
		orders.GET("/sips/:id/installments", controller.GetSIPInstallments)
		orders.POST("/sips/:id/cancel", controller.CancelSIP)

		// Redemption lifecycle
		orders.POST("/redemption", controller.PlaceRedemptionOrder)
		orders.GET("/redemptions", controller.ListRedemptions)
		orders.GET("/redemptions/:id", controller.GetRedemptionDetail)
		orders.PATCH("/redemptions/:id/confirm", controller.ConfirmRedemption)
	}
}
