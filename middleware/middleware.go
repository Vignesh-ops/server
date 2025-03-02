package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware ensures user is authenticated
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from cookie
		userID, err := c.Cookie("user_id")
		if err != nil || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Continue to the protected route
		c.Next()
	}
}
