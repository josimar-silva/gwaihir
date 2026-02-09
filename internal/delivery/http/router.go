// Package http provides HTTP delivery layer handlers and routes.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/josimar-silva/gwaihir/internal/infrastructure"
)

// NewRouter creates and configures the Gin router.
func NewRouter(handler *Handler) *gin.Engine {
	// Set Gin to release mode in production
	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Middleware
	router.Use(RequestIDMiddleware())
	router.Use(RequestLoggingMiddleware())

	// Health and version endpoints (no auth required)
	router.GET("/health", handler.Health)
	router.GET("/version", handler.Version)

	// Metrics endpoint (no auth required)
	router.GET("/metrics", gin.WrapH(infrastructure.MetricsHandler()))

	// WoL endpoints
	router.POST("/wol", handler.Wake)

	// Machine management endpoints
	router.GET("/machines", handler.ListMachines)
	router.GET("/machines/:id", handler.GetMachine)

	return router
}
