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
		orders.POST("/purchase", controller.PlacePurchaseOrder)
		orders.POST("/sip", controller.PlaceSIPOrder)
		orders.POST("/redemption", controller.PlaceRedemptionOrder)
		orders.POST("/:id/confirm-otp", controller.ConfirmOTP)
	}
}
