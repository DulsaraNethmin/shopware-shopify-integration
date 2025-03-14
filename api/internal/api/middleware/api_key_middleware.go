package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// In your auth.go middleware
func APIKeyMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the API Key from the request header
		requestAPIKey := c.GetHeader("X-API-Key")

		// Check if the API Key is valid
		if requestAPIKey == "" || requestAPIKey != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			return
		}

		// API Key is valid, continue
		c.Next()
	}
}
