package routes

import "github.com/gin-gonic/gin"

func Routes(router *gin.Engine) {
	base := router.Group("/sure-mf")

	// Health check
	base.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": 200, "msg": "ping success", "service": "SureMFService"})
	})

	fundRoutes(base)

	// All user-scoped routes are under /:uid/
	user := base.Group("/:uid")
	onboardingRoutes(user)
	orderRoutes(user)
}
