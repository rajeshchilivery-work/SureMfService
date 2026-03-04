package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.Param("uid")
		if uid == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": 401, "msg": "missing uid"})
			return
		}
		c.Set("uid", uid)
		c.Next()
	}
}
