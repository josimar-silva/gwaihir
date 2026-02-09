// Package http provides HTTP delivery layer.
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIKeyAuthMiddleware validates the X-API-Key header against the expected API key.
func APIKeyAuthMiddleware(expectedAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing X-API-Key header",
			})
			c.Abort()
			return
		}

		if apiKey != expectedAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
