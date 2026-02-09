// Package http provides HTTP delivery layer handlers and routes.
package http

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const requestIDKey = "request_id"
const contextRequestIDKey contextKey = "request_id"

// RequestIDMiddleware generates and injects a unique request ID for correlation.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set(requestIDKey, requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// RequestLoggingMiddleware logs request details with timing.
func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		requestID := getRequestID(c)

		c.Set(requestIDKey, requestID)

		c.Next()

		duration := time.Since(startTime)
		c.Set("duration", duration)
	}
}

// GetRequestID retrieves the request ID from the Gin context.
func GetRequestID(c *gin.Context) string {
	return getRequestID(c)
}

func getRequestID(c *gin.Context) string {
	if id, exists := c.Get(requestIDKey); exists {
		if str, ok := id.(string); ok {
			return str
		}
	}
	return fmt.Sprintf("unknown-%d", time.Now().UnixNano())
}

// ContextWithRequestID adds request ID to a context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, contextRequestIDKey, requestID)
}
