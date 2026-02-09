package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/josimar-silva/gwaihir/internal/domain"
)

func TestHealthCheckLive(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	healthHandler := NewHealthHandler(handler)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/live", healthHandler.HealthCheckLive)

	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if resp["status"] != "alive" {
		t.Errorf("Expected status 'alive', got %s", resp["status"])
	}
}

func TestHealthCheckReady_Success(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	healthHandler := NewHealthHandler(handler)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/ready", healthHandler.HealthCheckReady)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if resp["status"] != "ready" {
		t.Errorf("Expected status 'ready', got %v", resp["status"])
	}
}

func TestHealthCheckReady_NoMachines(t *testing.T) {
	handler, _, _ := newHandlerForTesting(map[string]*domain.Machine{})
	healthHandler := NewHealthHandler(handler)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/ready", healthHandler.HealthCheckReady)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if resp["status"] != "not ready" {
		t.Errorf("Expected status 'not ready', got %v", resp["status"])
	}
}

func TestHealthCheckFull_Success(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	healthHandler := NewHealthHandler(handler)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/health", healthHandler.HealthCheckFull)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp HealthCheckResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if resp.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got %s", resp.Status)
	}

	if resp.Version == "" {
		t.Error("Expected non-empty version")
	}

	if resp.BuildTime == "" {
		t.Error("Expected non-empty build_time")
	}

	if resp.GitCommit == "" {
		t.Error("Expected non-empty git_commit")
	}

	if resp.ConfiguredMachines != 2 {
		t.Errorf("Expected 2 configured machines, got %d", resp.ConfiguredMachines)
	}

	if resp.UptimeSeconds < 0 {
		t.Errorf("Expected non-negative uptime, got %d", resp.UptimeSeconds)
	}

	if resp.Checks["config_loaded"] != "ok" {
		t.Errorf("Expected config_loaded check to be 'ok', got %s", resp.Checks["config_loaded"])
	}

	if resp.Checks["machines"] != "ok" {
		t.Errorf("Expected machines check to be 'ok', got %s", resp.Checks["machines"])
	}
}

func TestHealthCheckFull_NoMachines(t *testing.T) {
	handler, _, _ := newHandlerForTesting(map[string]*domain.Machine{})
	healthHandler := NewHealthHandler(handler)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/health", healthHandler.HealthCheckFull)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp HealthCheckResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if resp.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got %s", resp.Status)
	}

	if resp.ConfiguredMachines != 0 {
		t.Errorf("Expected 0 configured machines, got %d", resp.ConfiguredMachines)
	}

	if resp.Checks["machines"] != "warning" {
		t.Errorf("Expected machines check to be 'warning', got %s", resp.Checks["machines"])
	}
}

func TestRouterHealthEndpoints(t *testing.T) {
	handler, _, _ := newHandlerForTesting(nil)
	router := NewRouterWithAuth(handler, "")

	testCases := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{"health endpoint", "/health", http.StatusOK},
		{"live endpoint", "/live", http.StatusOK},
		{"ready endpoint", "/ready", http.StatusOK},
		{"version endpoint", "/version", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}
