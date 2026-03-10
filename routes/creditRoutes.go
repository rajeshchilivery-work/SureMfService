package routes

import (
	"SureMFService/controller"
	"SureMFService/middleware"

	"github.com/gin-gonic/gin"
)

func creditRoutes(rg *gin.RouterGroup) {
	credit := rg.Group("/credit")
	credit.Use(middleware.AuthMiddleware())
	{
		credit.GET("/emi-roi-delta", controller.GetEMIROIDelta)
	}
}
