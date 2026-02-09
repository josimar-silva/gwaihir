// Package http provides HTTP delivery layer handlers and routes. //nolint:revive
package http

import (
	"github.com/gin-gonic/gin"
)

// NewRouter creates and configures the Gin router.
func NewRouter(handler *Handler) *gin.Engine {
	// Set Gin to release mode in production
	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Health and version endpoints
	router.GET("/health", handler.Health)
	router.GET("/version", handler.Version)

	// WoL endpoints
	router.POST("/wol", handler.Wake)

	// Machine management endpoints
	router.GET("/machines", handler.ListMachines)
	router.GET("/machines/:id", handler.GetMachine)

	return router
}
