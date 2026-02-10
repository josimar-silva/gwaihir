// Package http provides HTTP delivery layer handlers and routes.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/josimar-silva/gwaihir/internal/config"
	"github.com/josimar-silva/gwaihir/internal/infrastructure"
)

// NewRouter creates and configures the Gin router.
func NewRouter(handler *Handler) *gin.Engine {
	return NewRouterWithAuth(handler, "")
}

// NewRouterWithConfig creates and configures the Gin router based on config.
func NewRouterWithConfig(handler *Handler, cfg *config.Config) *gin.Engine {
	return NewRouterWithAuthAndConfig(handler, cfg.Authentication.APIKey, cfg)
}

// NewRouterWithAuth creates and configures the Gin router with optional API key authentication.
func NewRouterWithAuth(handler *Handler, apiKey string) *gin.Engine {
	return NewRouterWithAuthAndConfig(handler, apiKey, nil)
}

// NewRouterWithAuthAndConfig creates and configures the Gin router with config-based endpoint registration.
func NewRouterWithAuthAndConfig(handler *Handler, apiKey string, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(RequestIDMiddleware())
	router.Use(RequestLoggingMiddlewareWithConfig(cfg))

	healthEnabled := true
	if cfg != nil && cfg.Observability.HealthCheck.Enabled != nil {
		healthEnabled = *cfg.Observability.HealthCheck.Enabled
	}

	metricsEnabled := true
	if cfg != nil && cfg.Observability.Metrics.Enabled != nil {
		metricsEnabled = *cfg.Observability.Metrics.Enabled
	}

	healthHandler := NewHealthHandler(handler)

	if healthEnabled {
		router.GET("/health", healthHandler.HealthCheckFull)
		router.GET("/live", healthHandler.HealthCheckLive)
		router.GET("/ready", healthHandler.HealthCheckReady)
	}

	if metricsEnabled {
		router.GET("/metrics", gin.WrapH(infrastructure.MetricsHandler()))
	}

	router.GET("/version", handler.Version)

	protected := router.Group("")
	if apiKey != "" {
		protected.Use(APIKeyAuthMiddleware(apiKey))
	}

	protected.POST("/wol", handler.Wake)

	protected.GET("/machines", handler.ListMachines)
	protected.GET("/machines/:id", handler.GetMachine)

	return router
}
