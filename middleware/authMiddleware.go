package middleware

import (
	"SureMFService/database/firebase"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": 401, "msg": "missing authorization token"})
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := firebase.AuthClient.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": 401, "msg": "invalid or expired token"})
			return
		}

		c.Set("uid", token.UID)
		c.Next()
	}
}
