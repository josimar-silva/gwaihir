package http

import (
	"net/http"
	"testing"

	"github.com/josimar-silva/gwaihir/internal/config"
	"github.com/josimar-silva/gwaihir/internal/infrastructure"
	"github.com/josimar-silva/gwaihir/internal/repository"
	"github.com/josimar-silva/gwaihir/internal/usecase"
)

// Test 4.1.1: Router accepts config
func TestNewRouterWithConfig_AcceptsConfig(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Machines: []config.MachineConfig{
			{
				ID:        "test",
				Name:      "Test",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
		},
		Observability: config.ObservabilityConfig{
			HealthCheck: config.HealthCheckConfig{Enabled: ptrBool(true)},
			Metrics:     config.MetricsConfig{Enabled: ptrBool(true)},
		},
	}

	logger := infrastructure.NewLogger("text", "debug")
	metrics, _ := infrastructure.NewMetrics()
	repo, _ := repository.NewInMemoryMachineRepository(cfg)
	packetSender := repository.NewWoLPacketSender()
	useCase := usecase.NewWoLUseCase(repo, packetSender, logger, metrics)
	handler := NewHandler(useCase, logger, metrics, "0.1.0", "2026-02-10", "abc")

	// Act
	router := NewRouterWithConfig(handler, cfg)

	// Assert
	if router == nil {
		t.Fatal("Expected non-nil router")
	}
}

// Test 4.1.2/4.1.3: Health endpoints based on config
func TestNewRouterWithConfig_HealthEndpoints(t *testing.T) {
	tests := []struct {
		name          string
		healthEnabled *bool
		expectedCount int
	}{
		{
			name:          "enabled",
			healthEnabled: ptrBool(true),
			expectedCount: 3,
		},
		{
			name:          "disabled",
			healthEnabled: ptrBool(false),
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Machines: []config.MachineConfig{
					{
						ID:        "test",
						Name:      "Test",
						MAC:       "AA:BB:CC:DD:EE:FF",
						Broadcast: "192.168.1.255",
					},
				},
				Observability: config.ObservabilityConfig{
					HealthCheck: config.HealthCheckConfig{Enabled: tt.healthEnabled},
					Metrics:     config.MetricsConfig{Enabled: ptrBool(true)},
				},
			}

			logger := infrastructure.NewLogger("text", "debug")
			metrics, _ := infrastructure.NewMetrics()
			repo, _ := repository.NewInMemoryMachineRepository(cfg)
			packetSender := repository.NewWoLPacketSender()
			useCase := usecase.NewWoLUseCase(repo, packetSender, logger, metrics)
			handler := NewHandler(useCase, logger, metrics, "0.1.0", "2026-02-10", "abc")

			router := NewRouterWithConfig(handler, cfg)

			route := router.Routes()
			healthRoutes := 0
			for _, r := range route {
				if (r.Path == "/health" || r.Path == "/live" || r.Path == "/ready") && r.Method == http.MethodGet {
					healthRoutes++
				}
			}

			if healthRoutes != tt.expectedCount {
				t.Errorf("Expected %d health routes, got %d", tt.expectedCount, healthRoutes)
			}
		})
	}
}

// Test 4.1.4: Metrics endpoint registered when enabled
func TestNewRouterWithConfig_MetricsEndpointEnabled(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Machines: []config.MachineConfig{
			{
				ID:        "test",
				Name:      "Test",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
		},
		Observability: config.ObservabilityConfig{
			HealthCheck: config.HealthCheckConfig{Enabled: ptrBool(true)},
			Metrics:     config.MetricsConfig{Enabled: ptrBool(true)},
		},
	}

	logger := infrastructure.NewLogger("text", "debug")
	metrics, _ := infrastructure.NewMetrics()
	repo, _ := repository.NewInMemoryMachineRepository(cfg)
	packetSender := repository.NewWoLPacketSender()
	useCase := usecase.NewWoLUseCase(repo, packetSender, logger, metrics)
	handler := NewHandler(useCase, logger, metrics, "0.1.0", "2026-02-10", "abc")

	router := NewRouterWithConfig(handler, cfg)

	// Metrics endpoint should be registered
	route := router.Routes()
	metricsFound := false
	for _, r := range route {
		if r.Path == "/metrics" && r.Method == http.MethodGet {
			metricsFound = true
			break
		}
	}

	if !metricsFound {
		t.Error("Expected metrics endpoint to be registered")
	}
}

// Test 4.1.5: Protected endpoints always registered
func TestNewRouterWithConfig_ProtectedEndpointsAlwaysRegistered(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Machines: []config.MachineConfig{
			{
				ID:        "test",
				Name:      "Test",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
		},
		Authentication: config.AuthenticationConfig{APIKey: "test-key"},
		Observability: config.ObservabilityConfig{
			HealthCheck: config.HealthCheckConfig{Enabled: ptrBool(false)},
			Metrics:     config.MetricsConfig{Enabled: ptrBool(false)},
		},
	}

	logger := infrastructure.NewLogger("text", "debug")
	metrics, _ := infrastructure.NewMetrics()
	repo, _ := repository.NewInMemoryMachineRepository(cfg)
	packetSender := repository.NewWoLPacketSender()
	useCase := usecase.NewWoLUseCase(repo, packetSender, logger, metrics)
	handler := NewHandler(useCase, logger, metrics, "0.1.0", "2026-02-10", "abc")

	router := NewRouterWithConfig(handler, cfg)

	// Protected endpoints should always be registered
	route := router.Routes()
	protectedEndpoints := 0
	for _, r := range route {
		if (r.Path == "/wol" || r.Path == "/machines" || r.Path == "/machines/:id") && r.Method != "OPTIONS" {
			protectedEndpoints++
		}
	}

	if protectedEndpoints != 3 {
		t.Errorf("Expected 3 protected endpoints, got %d", protectedEndpoints)
	}
}

// Helper function
func ptrBool(b bool) *bool {
	return &b
}
