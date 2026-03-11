package routes

import (
	"SureMFService/controller"

	"github.com/gin-gonic/gin"
)

func fundRoutes(rg *gin.RouterGroup) {
	funds := rg.Group("/funds")
	{
		funds.GET("", controller.ListFunds)
		funds.GET("/:isin", controller.GetFund)
	}
}
