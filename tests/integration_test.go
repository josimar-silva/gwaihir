//go:build integration
// +build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	httpdelivery "github.com/josimar-silva/gwaihir/internal/delivery/http"
	"github.com/josimar-silva/gwaihir/internal/infrastructure"
	"github.com/josimar-silva/gwaihir/internal/repository"
	"github.com/josimar-silva/gwaihir/internal/usecase"
)

// getConfigPath returns the path to the machines.yaml config file for integration tests.
// It checks (in order):
// 1. tests/testdata/machines.yaml (test-specific config, preferred)
// 2. GWAIHIR_CONFIG environment variable (for CI/CD override)
// 3. configs/machines.yaml (fallback to main config)
func getConfigPath(t *testing.T) string {
	projectRoot := findProjectRoot(t)

	// Priority 1: Test-specific config in testdata
	testConfigPath := filepath.Join(projectRoot, "tests", "testdata", "machines.yaml")
	if _, err := os.Stat(testConfigPath); err == nil {
		return testConfigPath
	}

	// Priority 2: Environment variable (allows CI/CD override)
	if configPath := os.Getenv("GWAIHIR_CONFIG"); configPath != "" {
		return configPath
	}

	// Priority 3: Main config as fallback
	mainConfigPath := filepath.Join(projectRoot, "configs", "machines.yaml")
	if _, err := os.Stat(mainConfigPath); err == nil {
		return mainConfigPath
	}

	t.Fatalf("Could not find machines.yaml. Checked:\n  - %s (test config)\n  - GWAIHIR_CONFIG env var\n  - %s (main config)", testConfigPath, mainConfigPath)
	return ""
}

// findProjectRoot walks up the directory tree until it finds go.mod.
func findProjectRoot(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd
		}

		parent := filepath.Dir(cwd)
		if parent == cwd {
			t.Fatalf("Could not find project root (go.mod). Current directory: %s", cwd)
		}
		cwd = parent
	}
}

// startServer starts a test server and returns the base URL and cleanup function.
func startServer(t *testing.T, port string, configPath string) (string, func()) {
	logger := infrastructure.NewLogger(false)

	machineRepo, err := repository.NewYAMLMachineRepository(configPath)
	if err != nil {
		t.Fatalf("Failed to initialize machine repository: %v", err)
	}

	packetSender := repository.NewWoLPacketSender()

	// Use a custom metrics structure to avoid conflicts
	metrics := &infrastructure.Metrics{
		WoLPacketsSent: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_wol_packets_sent_total",
			Help: "Total number of WoL packets successfully sent",
		}),
		WoLPacketsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_wol_packets_failed_total",
			Help: "Total number of WoL packet send failures",
		}),
		MachineNotFound: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_machine_not_found_total",
			Help: "Total number of machine not found errors",
		}),
		MachinesListed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_machines_listed_total",
			Help: "Total number of times machines list was requested",
		}),
		MachinesRetrieved: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_machines_retrieved_total",
			Help: "Total number of times a machine was retrieved by ID",
		}),
		RequestDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "gwaihir_request_duration_seconds",
			Help:    "Request latency in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		ConfiguredMachines: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gwaihir_configured_machines_total",
			Help: "Total number of configured machines in allowlist",
		}),
	}

	wolUseCase := usecase.NewWoLUseCase(machineRepo, packetSender, logger, metrics)
	handler := httpdelivery.NewHandler(wolUseCase, logger, metrics, "0.2.0", "2026-02-09", "abc123")

	router := httpdelivery.NewRouter(handler)
	server := &http.Server{
		Addr:              net.JoinHostPort("localhost", port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			t.Logf("Server shutdown error: %v", err)
		}
	}

	return fmt.Sprintf("http://localhost:%s", port), cleanup
}

// TestIntegration_WoLEndpoint tests the full WoL endpoint with a live API.
func TestIntegration_WoLEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	configPath := getConfigPath(t)
	baseURL, cleanup := startServer(t, "8080", configPath)
	defer cleanup()

	testCases := []struct {
		name           string
		machineID      string
		expectedStatus int
	}{
		{
			name:           "Valid machine ID",
			machineID:      "saruman",
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "Invalid machine ID",
			machineID:      "nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload := map[string]string{"machine_id": tc.machineID}
			body, _ := json.Marshal(payload)

			resp, err := http.Post(
				baseURL+"/wol",
				"application/json",
				bytes.NewReader(body),
			)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				respBody, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Response: %s", tc.expectedStatus, resp.StatusCode, string(respBody))
			}
		})
	}
}

// TestIntegration_ListMachinesEndpoint tests the machines list endpoint.
func TestIntegration_ListMachinesEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	configPath := getConfigPath(t)
	baseURL, cleanup := startServer(t, "8081", configPath)
	defer cleanup()

	resp, err := http.Get(baseURL + "/machines")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var machines []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&machines); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(machines) == 0 {
		t.Error("Expected machines to be returned")
	}
}

// TestIntegration_HealthEndpoints tests all health check endpoints.
func TestIntegration_HealthEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	configPath := getConfigPath(t)
	baseURL, cleanup := startServer(t, "8082", configPath)
	defer cleanup()

	testCases := []struct {
		endpoint       string
		expectedStatus int
	}{
		{"/health", http.StatusOK},
		{"/live", http.StatusOK},
		{"/ready", http.StatusOK},
		{"/version", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Endpoint_%s", tc.endpoint), func(t *testing.T) {
			resp, err := http.Get(baseURL + tc.endpoint)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			var response map[string]interface{}
			if err := json.Unmarshal(body, &response); err != nil {
				t.Errorf("Failed to decode response: %v", err)
			}
		})
	}
}

// TestIntegration_MetricsEndpoint tests the Prometheus metrics endpoint.
func TestIntegration_MetricsEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	configPath := getConfigPath(t)
	baseURL, cleanup := startServer(t, "8083", configPath)
	defer cleanup()

	resp, err := http.Get(baseURL + "/metrics")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Verify it's valid Prometheus format (contains HELP or TYPE)
	if !bytes.Contains([]byte(bodyStr), []byte("# HELP")) && !bytes.Contains([]byte(bodyStr), []byte("# TYPE")) {
		t.Error("Expected Prometheus format (HELP or TYPE comments) not found")
	}

	// Verify it returns valid content
	if len(bodyStr) == 0 {
		t.Error("Expected non-empty metrics response")
	}
}

// TestIntegration_APIKeyAuthentication tests API key authentication.
func TestIntegration_APIKeyAuthentication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := infrastructure.NewLogger(false)

	configPath := getConfigPath(t)
	machineRepo, err := repository.NewYAMLMachineRepository(configPath)
	if err != nil {
		t.Fatalf("Failed to initialize machine repository: %v", err)
	}

	apiKey := "test-api-key"
	packetSender := repository.NewWoLPacketSender()

	// Create isolated metrics
	metrics := &infrastructure.Metrics{
		WoLPacketsSent: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_wol_packets_sent_total",
			Help: "Total number of WoL packets successfully sent",
		}),
		WoLPacketsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_wol_packets_failed_total",
			Help: "Total number of WoL packet send failures",
		}),
		MachineNotFound: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_machine_not_found_total",
			Help: "Total number of machine not found errors",
		}),
		MachinesListed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_machines_listed_total",
			Help: "Total number of times machines list was requested",
		}),
		MachinesRetrieved: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_machines_retrieved_total",
			Help: "Total number of times a machine was retrieved by ID",
		}),
		RequestDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "gwaihir_request_duration_seconds",
			Help:    "Request latency in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		ConfiguredMachines: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gwaihir_configured_machines_total",
			Help: "Total number of configured machines in allowlist",
		}),
	}

	wolUseCase := usecase.NewWoLUseCase(machineRepo, packetSender, logger, metrics)
	handler := httpdelivery.NewHandler(wolUseCase, logger, metrics, "0.2.0", "2026-02-09", "abc123")

	router := httpdelivery.NewRouterWithAuth(handler, apiKey)
	server := &http.Server{
		Addr:              net.JoinHostPort("localhost", "8084"),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			t.Logf("Server shutdown error: %v", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	baseURL := "http://localhost:8084"
	testCases := []struct {
		name           string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "Valid API key",
			apiKey:         apiKey,
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "Invalid API key",
			apiKey:         "wrong-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing API key",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload := map[string]string{"machine_id": "saruman"}
			body, _ := json.Marshal(payload)

			req, _ := http.NewRequest(
				http.MethodPost,
				baseURL+"/wol",
				bytes.NewReader(body),
			)
			req.Header.Set("Content-Type", "application/json")

			if tc.apiKey != "" {
				req.Header.Set("X-API-Key", tc.apiKey)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
		})
	}
}
