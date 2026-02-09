// Package http provides HTTP delivery layer using Gin. //nolint:revive
package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/josimar-silva/gwaihir/internal/domain"
	"github.com/josimar-silva/gwaihir/internal/usecase"
)

// Handler handles HTTP requests for WoL operations.
type Handler struct {
	wolUseCase *usecase.WoLUseCase
	version    string
	buildTime  string
	gitCommit  string
}

// NewHandler creates a new HTTP handler.
func NewHandler(wolUseCase *usecase.WoLUseCase, version, buildTime, gitCommit string) *Handler {
	return &Handler{
		wolUseCase: wolUseCase,
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
	var req WakeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.wolUseCase.SendWakePacket(req.MachineID); err != nil {
		if errors.Is(err, domain.ErrMachineNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Machine not found or not allowed",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to send WoL packet: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, SuccessResponse{
		Message: "WoL packet sent successfully",
	})
}

// ListMachines handles GET /machines requests.
func (h *Handler) ListMachines(c *gin.Context) {
	machines, err := h.wolUseCase.ListMachines()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve machines: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, machines)
}

// GetMachine handles GET /machines/:id requests.
func (h *Handler) GetMachine(c *gin.Context) {
	machineID := c.Param("id")

	machine, err := h.wolUseCase.GetMachine(machineID)
	if err != nil {
		if errors.Is(err, domain.ErrMachineNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Machine not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to retrieve machine: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, machine)
}

// Health handles GET /health requests.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// Version handles GET /version requests.
func (h *Handler) Version(c *gin.Context) {
	c.JSON(http.StatusOK, VersionResponse{
		Version:   h.version,
		BuildTime: h.buildTime,
		GitCommit: h.gitCommit,
	})
}
