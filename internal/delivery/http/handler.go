// Package http provides HTTP delivery layer using Gin.
package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/josimar-silva/gwaihir/internal/domain"
	"github.com/josimar-silva/gwaihir/internal/infrastructure"
	"github.com/josimar-silva/gwaihir/internal/usecase"
)

// Handler handles HTTP requests for WoL operations.
type Handler struct {
	wolUseCase *usecase.WoLUseCase
	logger     *infrastructure.Logger
	metrics    *infrastructure.Metrics
	version    string
	buildTime  string
	gitCommit  string
}

// NewHandler creates a new HTTP handler.
func NewHandler(wolUseCase *usecase.WoLUseCase, logger *infrastructure.Logger, metrics *infrastructure.Metrics, version, buildTime, gitCommit string) *Handler {
	return &Handler{
		wolUseCase: wolUseCase,
		logger:     logger,
		metrics:    metrics,
		version:    version,
		buildTime:  buildTime,
		gitCommit:  gitCommit,
	}
}

// WakeRequest represents the JSON request to wake a machine.
type WakeRequest struct {
	MachineID string `json:"machine_id" binding:"required"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a success response.
type SuccessResponse struct {
	Message string `json:"message"`
}

// VersionResponse represents version information.
type VersionResponse struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
}

// Wake handles POST /wol requests.
func (h *Handler) Wake(c *gin.Context) {
	startTime := time.Now()
	requestID := GetRequestID(c)

	var req WakeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid WoL request",
			infrastructure.String("request_id", requestID),
			infrastructure.Any("error", err),
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
		h.metrics.RequestDuration.Observe(time.Since(startTime).Seconds())
		return
	}

	if err := h.wolUseCase.SendWakePacket(req.MachineID); err != nil {
		duration := time.Since(startTime).Seconds()
		h.metrics.RequestDuration.Observe(duration)

		if errors.Is(err, domain.ErrMachineNotFound) {
			h.logger.Warn("Machine not found",
				infrastructure.String("request_id", requestID),
				infrastructure.String("machine_id", req.MachineID),
			)
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Machine not found or not allowed",
			})
			return
		}

		h.logger.Error("Failed to send WoL packet",
			infrastructure.String("request_id", requestID),
			infrastructure.String("machine_id", req.MachineID),
			infrastructure.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to send WoL packet: " + err.Error(),
		})
		return
	}

	h.logger.Info("WoL packet sent successfully",
		infrastructure.String("request_id", requestID),
		infrastructure.String("machine_id", req.MachineID),
	)
	h.metrics.RequestDuration.Observe(time.Since(startTime).Seconds())

	c.JSON(http.StatusAccepted, SuccessResponse{
		Message: "WoL packet sent successfully",
	})
}

// ListMachines handles GET /machines requests.
func (h *Handler) ListMachines(c *gin.Context) {
	startTime := time.Now()
	requestID := GetRequestID(c)

	machines, err := h.wolUseCase.ListMachines()
	duration := time.Since(startTime).Seconds()
	h.metrics.RequestDuration.Observe(duration)

	if err != nil {
		h.logger.Error("Failed to retrieve machines",
			infrastructure.String("request_id", requestID),
			infrastructure.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve machines: " + err.Error(),
		})
		return
	}

	h.metrics.MachinesListed.Inc()
	h.logger.Info("Machines list retrieved",
		infrastructure.String("request_id", requestID),
		infrastructure.Int("count", len(machines)),
	)

	c.JSON(http.StatusOK, machines)
}

// GetMachine handles GET /machines/:id requests.
func (h *Handler) GetMachine(c *gin.Context) {
	startTime := time.Now()
	requestID := GetRequestID(c)
	machineID := c.Param("id")

	machine, err := h.wolUseCase.GetMachine(machineID)
	duration := time.Since(startTime).Seconds()
	h.metrics.RequestDuration.Observe(duration)

	if err != nil {
		if errors.Is(err, domain.ErrMachineNotFound) {
			h.logger.Warn("Machine not found",
				infrastructure.String("request_id", requestID),
				infrastructure.String("machine_id", machineID),
			)
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Machine not found",
			})
			return
		}

		h.logger.Error("Failed to retrieve machine",
			infrastructure.String("request_id", requestID),
			infrastructure.String("machine_id", machineID),
			infrastructure.Any("error", err),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve machine: " + err.Error(),
		})
		return
	}

	h.metrics.MachinesRetrieved.Inc()
	h.logger.Info("Machine retrieved",
		infrastructure.String("request_id", requestID),
		infrastructure.String("machine_id", machineID),
	)

	c.JSON(http.StatusOK, machine)
}

// Health handles GET /health requests.
func (h *Handler) Health(c *gin.Context) {
	requestID := GetRequestID(c)
	h.logger.Debug("Health check requested",
		infrastructure.String("request_id", requestID),
	)
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// Version handles GET /version requests.
func (h *Handler) Version(c *gin.Context) {
	requestID := GetRequestID(c)
	h.logger.Debug("Version info requested",
		infrastructure.String("request_id", requestID),
	)
	c.JSON(http.StatusOK, VersionResponse{
		Version:   h.version,
		BuildTime: h.buildTime,
		GitCommit: h.gitCommit,
	})
}
