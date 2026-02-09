// Package http provides HTTP delivery layer handlers and routes.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/josimar-silva/gwaihir/internal/infrastructure"
)

// NewRouter creates and configures the Gin router.
func NewRouter(handler *Handler) *gin.Engine {
	return NewRouterWithAuth(handler, "")
}

// NewRouterWithAuth creates and configures the Gin router with optional API key authentication.
func NewRouterWithAuth(handler *Handler, apiKey string) *gin.Engine {
	// Set Gin to release mode in production
	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Middleware
	router.Use(RequestIDMiddleware())
	router.Use(RequestLoggingMiddleware())

	// Create health handler for enhanced health checks
	healthHandler := NewHealthHandler(handler)

	// Health and version endpoints (no auth required)
	router.GET("/health", healthHandler.HealthCheckFull)
	router.GET("/live", healthHandler.HealthCheckLive)
	router.GET("/ready", healthHandler.HealthCheckReady)
	router.GET("/version", handler.Version)

	// Metrics endpoint (no auth required)
	router.GET("/metrics", gin.WrapH(infrastructure.MetricsHandler()))

	// Protected endpoints group (with optional API key auth)
	protected := router.Group("")
	if apiKey != "" {
		protected.Use(APIKeyAuthMiddleware(apiKey))
	}

	// WoL endpoints (protected)
	protected.POST("/wol", handler.Wake)

	// Machine management endpoints (protected)
	protected.GET("/machines", handler.ListMachines)
	protected.GET("/machines/:id", handler.GetMachine)

	return router
}
