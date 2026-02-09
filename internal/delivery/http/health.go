// Package http provides HTTP delivery layer.
package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/josimar-silva/gwaihir/internal/infrastructure"
)

// HealthCheckResponse represents the health check response.
type HealthCheckResponse struct {
	Status             string            `json:"status"`
	Version            string            `json:"version"`
	BuildTime          string            `json:"build_time"`
	GitCommit          string            `json:"git_commit"`
	Timestamp          string            `json:"timestamp"`
	UptimeSeconds      int64             `json:"uptime_seconds"`
	ConfiguredMachines int               `json:"configured_machines"`
	Checks             map[string]string `json:"checks"`
}

// HealthHandler tracks the start time for uptime calculation.
type HealthHandler struct {
	handler   *Handler
	startTime time.Time
}

// NewHealthHandler creates a new health handler with the current time.
func NewHealthHandler(handler *Handler) *HealthHandler {
	return &HealthHandler{
		handler:   handler,
		startTime: time.Now(),
	}
}

// HealthCheckLive handles GET /live requests (liveness probe).
func (h *HealthHandler) HealthCheckLive(c *gin.Context) {
	requestID := GetRequestID(c)
	h.handler.logger.Debug("Liveness probe requested",
		infrastructure.String("request_id", requestID),
	)

	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// HealthCheckReady handles GET /ready requests (readiness probe).
func (h *HealthHandler) HealthCheckReady(c *gin.Context) {
	requestID := GetRequestID(c)

	// Get machine count from usecase
	machines, err := h.handler.wolUseCase.ListMachines()
	if err != nil {
		h.handler.logger.Error("Readiness check failed: unable to list machines",
			infrastructure.String("request_id", requestID),
			infrastructure.Any("error", err),
		)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "unable to load machine configuration",
		})
		return
	}

	if len(machines) == 0 {
		h.handler.logger.Warn("Readiness check: no machines configured",
			infrastructure.String("request_id", requestID),
		)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "no machines configured",
		})
		return
	}

	h.handler.logger.Debug("Readiness check passed",
		infrastructure.String("request_id", requestID),
		infrastructure.Int("machines", len(machines)),
	)

	c.JSON(http.StatusOK, gin.H{
		"status":   "ready",
		"machines": len(machines),
	})
}

// HealthCheckFull handles GET /health requests (combined health check).
func (h *HealthHandler) HealthCheckFull(c *gin.Context) {
	requestID := GetRequestID(c)

	// Get machine count
	machines, err := h.handler.wolUseCase.ListMachines()
	machineCount := 0
	machineCheckStatus := "ok"
	if err != nil {
		machineCheckStatus = "error"
	} else {
		machineCount = len(machines)
		if machineCount == 0 {
			machineCheckStatus = "warning"
		}
	}

	checks := map[string]string{
		"config_loaded": "ok",
		"machines":      machineCheckStatus,
	}

	response := HealthCheckResponse{
		Status:             "healthy",
		Version:            h.handler.version,
		BuildTime:          h.handler.buildTime,
		GitCommit:          h.handler.gitCommit,
		Timestamp:          time.Now().UTC().Format(time.RFC3339),
		UptimeSeconds:      int64(time.Since(h.startTime).Seconds()),
		ConfiguredMachines: machineCount,
		Checks:             checks,
	}

	statusCode := http.StatusOK
	if machineCheckStatus == "error" {
		response.Status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	h.handler.logger.Debug("Health check requested",
		infrastructure.String("request_id", requestID),
		infrastructure.String("status", response.Status),
	)

	c.JSON(statusCode, response)
}
