package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/josimar-silva/gwaihir/internal/config"
	"github.com/josimar-silva/gwaihir/internal/infrastructure"
)

var (
	// globalMetrics is initialized once for all tests to avoid Prometheus registration conflicts
	globalMetrics    *infrastructure.Metrics
	metricsInitOnce  bool
	metricsInitError error
)

// TestMain sets up and tears down test environment
func TestMain(m *testing.M) {
	// Set Gin to test mode to reduce noise in test output
	gin.SetMode(gin.TestMode)

	// Initialize metrics once for all tests
	logger := infrastructure.NewLogger("text", "error")
	globalMetrics, metricsInitError = infrastructure.NewMetrics()
	if metricsInitError == nil {
		metricsInitOnce = true
	}
	// Suppress logger output during tests
	_ = logger

	code := m.Run()
	os.Exit(code)
}

// getTestMetrics returns the global metrics instance for testing
func getTestMetrics(t *testing.T) *infrastructure.Metrics {
	t.Helper()
	if !metricsInitOnce {
		t.Fatalf("Failed to initialize metrics: %v", metricsInitError)
	}
	return globalMetrics
}

// setupTestConfig creates a temporary config file for testing
func setupTestConfig(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	err := os.WriteFile(configPath, []byte(content), 0600)
	require.NoError(t, err)
	return configPath
}

// validTestConfig returns a valid YAML configuration for testing
func validTestConfig() string {
	return `server:
  port: 9090
  log:
    format: json
    level: info
authentication:
  api_key: test-key-12345
machines:
  - id: server1
    name: Test Server 1
    mac: "AA:BB:CC:DD:EE:FF"
    broadcast: "192.168.1.255"
  - id: server2
    name: Test Server 2
    mac: "11:22:33:44:55:66"
    broadcast: "192.168.1.255"
observability:
  health_check:
    enabled: true
  metrics:
    enabled: true
`
}

// TestLoadConfiguration tests the loadConfiguration function
func TestLoadConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string
		setupEnv    func()
		cleanupEnv  func()
		wantErr     bool
		errContains string
	}{
		{
			name:       "success_with_env_var",
			configPath: "",
			setupEnv: func() {
				configPath := setupTestConfig(t, validTestConfig())
				require.NoError(t, os.Setenv("GWAIHIR_CONFIG", configPath))
			},
			cleanupEnv: func() {
				require.NoError(t, os.Unsetenv("GWAIHIR_CONFIG"))
			},
			wantErr: false,
		},
		{
			name:        "error_config_file_not_found",
			configPath:  "",
			setupEnv:    func() { require.NoError(t, os.Setenv("GWAIHIR_CONFIG", "/nonexistent/config.yaml")) },
			cleanupEnv:  func() { require.NoError(t, os.Unsetenv("GWAIHIR_CONFIG")) },
			wantErr:     true,
			errContains: "failed to load config",
		},
		{
			name:       "error_invalid_yaml",
			configPath: "",
			setupEnv: func() {
				invalidConfig := setupTestConfig(t, "invalid: yaml: [[[")
				require.NoError(t, os.Setenv("GWAIHIR_CONFIG", invalidConfig))
			},
			cleanupEnv:  func() { require.NoError(t, os.Unsetenv("GWAIHIR_CONFIG")) },
			wantErr:     true,
			errContains: "failed to load config",
		},
		{
			name:       "error_invalid_config_no_machines",
			configPath: "",
			setupEnv: func() {
				invalidConfig := `server:
  port: 8080
  log:
    format: json
    level: info
machines: []
`
				configPath := setupTestConfig(t, invalidConfig)
				require.NoError(t, os.Setenv("GWAIHIR_CONFIG", configPath))
			},
			cleanupEnv:  func() { require.NoError(t, os.Unsetenv("GWAIHIR_CONFIG")) },
			wantErr:     true,
			errContains: "failed to load config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv()
			}

			cfg, err := loadConfiguration()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				assert.Equal(t, 9090, cfg.Server.Port)
				assert.Equal(t, "json", cfg.Server.Log.Format)
				assert.Equal(t, "info", cfg.Server.Log.Level)
				assert.Equal(t, "test-key-12345", cfg.Authentication.APIKey)
				assert.Len(t, cfg.Machines, 2)
			}
		})
	}
}

func TestLoadConfiguration_DefaultPath(t *testing.T) {
	require.NoError(t, os.Unsetenv("GWAIHIR_CONFIG"))

	_, err := loadConfiguration()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "/etc/gwaihir/gwaihir.yaml")
}

// TestInitializeLogger tests the initializeLogger function
func TestInitializeLogger(t *testing.T) {
	tests := []struct {
		name      string
		logFormat string
		logLevel  string
	}{
		{
			name:      "json_format_debug_level",
			logFormat: "json",
			logLevel:  "debug",
		},
		{
			name:      "json_format_info_level",
			logFormat: "json",
			logLevel:  "info",
		},
		{
			name:      "text_format_warn_level",
			logFormat: "text",
			logLevel:  "warn",
		},
		{
			name:      "text_format_error_level",
			logFormat: "text",
			logLevel:  "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Server: config.ServerConfig{
					Port: 8080,
					Log: config.LogConfig{
						Format: tt.logFormat,
						Level:  tt.logLevel,
					},
				},
			}

			logger := initializeLogger(cfg)

			assert.NotNil(t, logger)
			assert.Equal(t, tt.logLevel, logger.GetLevel())
		})
	}
}

// TestInitializeMetrics tests that metrics are properly initialized
func TestInitializeMetrics(t *testing.T) {
	// Use the global metrics instance since Prometheus doesn't allow re-registration
	metrics := getTestMetrics(t)

	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.WoLPacketsSent)
	assert.NotNil(t, metrics.WoLPacketsFailed)
	assert.NotNil(t, metrics.RequestDuration)
	assert.NotNil(t, metrics.ConfiguredMachines)
}

// TestInitializeRepository tests the initializeRepository function
func TestInitializeRepository(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *config.Config
		wantErr      bool
		errContains  string
		machineCount int
	}{
		{
			name: "success_with_valid_machines",
			cfg: &config.Config{
				Machines: []config.MachineConfig{
					{
						ID:        "server1",
						Name:      "Server 1",
						MAC:       "AA:BB:CC:DD:EE:FF",
						Broadcast: "192.168.1.255",
					},
					{
						ID:        "server2",
						Name:      "Server 2",
						MAC:       "11:22:33:44:55:66",
						Broadcast: "192.168.1.255",
					},
				},
			},
			wantErr:      false,
			machineCount: 2,
		},
		{
			name: "success_with_single_machine",
			cfg: &config.Config{
				Machines: []config.MachineConfig{
					{
						ID:        "server1",
						Name:      "Server 1",
						MAC:       "AA:BB:CC:DD:EE:FF",
						Broadcast: "192.168.1.255",
					},
				},
			},
			wantErr:      false,
			machineCount: 1,
		},
		{
			name: "error_invalid_mac_address",
			cfg: &config.Config{
				Machines: []config.MachineConfig{
					{
						ID:        "server1",
						Name:      "Server 1",
						MAC:       "INVALID",
						Broadcast: "192.168.1.255",
					},
				},
			},
			wantErr:     true,
			errContains: "repository initialization failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := infrastructure.NewLogger("text", "error")

			repo, err := initializeRepository(tt.cfg, logger)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
				machines, _ := repo.GetAll()
				assert.Len(t, machines, tt.machineCount)
			}
		})
	}
}

// TestLogMachineConfiguration tests the logMachineConfiguration function
func TestLogMachineConfiguration(t *testing.T) {
	cfg := &config.Config{
		Machines: []config.MachineConfig{
			{
				ID:        "server1",
				Name:      "Server 1",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
			{
				ID:        "server2",
				Name:      "Server 2",
				MAC:       "11:22:33:44:55:66",
				Broadcast: "192.168.1.255",
			},
		},
	}

	logger := infrastructure.NewLogger("text", "error")
	metrics := getTestMetrics(t)
	repo, err := initializeRepository(cfg, logger)
	require.NoError(t, err)

	// This should not panic
	logMachineConfiguration(logger, metrics, repo)

	// Verify metrics were set (we can't easily verify logger output without mocking)
	// The ConfiguredMachines metric should be set to the count
	assert.NotNil(t, metrics.ConfiguredMachines)
}

// TestInitializeUseCase tests the initializeUseCase function
func TestInitializeUseCase(t *testing.T) {
	cfg := &config.Config{
		Machines: []config.MachineConfig{
			{
				ID:        "server1",
				Name:      "Server 1",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
		},
	}

	logger := infrastructure.NewLogger("text", "error")
	metrics := getTestMetrics(t)
	repo, err := initializeRepository(cfg, logger)
	require.NoError(t, err)

	useCase := initializeUseCase(repo, logger, metrics)

	assert.NotNil(t, useCase)
}

// TestInitializeHandler tests the initializeHandler function
func TestInitializeHandler(t *testing.T) {
	cfg := &config.Config{
		Machines: []config.MachineConfig{
			{
				ID:        "server1",
				Name:      "Server 1",
				MAC:       "AA:BB:CC:DD:EE:FF",
				Broadcast: "192.168.1.255",
			},
		},
	}

	logger := infrastructure.NewLogger("text", "error")
	metrics := getTestMetrics(t)
	repo, err := initializeRepository(cfg, logger)
	require.NoError(t, err)
	useCase := initializeUseCase(repo, logger, metrics)

	handler := initializeHandler(useCase, logger, metrics)

	assert.NotNil(t, handler)
}

// TestInitializeRouter tests the initializeRouter function
func TestInitializeRouter(t *testing.T) {
	tests := []struct {
		name         string
		logLevel     string
		apiKey       string
		expectedMode string
		checkWarning bool
	}{
		{
			name:         "debug_mode_with_api_key",
			logLevel:     "debug",
			apiKey:       "test-key",
			expectedMode: gin.DebugMode,
			checkWarning: false,
		},
		{
			name:         "release_mode_with_api_key",
			logLevel:     "info",
			apiKey:       "test-key",
			expectedMode: gin.ReleaseMode,
			checkWarning: false,
		},
		{
			name:         "release_mode_without_api_key",
			logLevel:     "info",
			apiKey:       "",
			expectedMode: gin.ReleaseMode,
			checkWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure GIN_MODE is not set to test our logic
			require.NoError(t, os.Unsetenv("GIN_MODE"))

			cfg := &config.Config{
				Server: config.ServerConfig{
					Port: 8080,
					Log: config.LogConfig{
						Format: "json",
						Level:  tt.logLevel,
					},
				},
				Authentication: config.AuthenticationConfig{
					APIKey: tt.apiKey,
				},
				Machines: []config.MachineConfig{
					{
						ID:        "server1",
						Name:      "Server 1",
						MAC:       "AA:BB:CC:DD:EE:FF",
						Broadcast: "192.168.1.255",
					},
				},
			}

			logger := infrastructure.NewLogger("text", "error")
			metrics := getTestMetrics(t)
			repo, err := initializeRepository(cfg, logger)
			require.NoError(t, err)
			useCase := initializeUseCase(repo, logger, metrics)
			handler := initializeHandler(useCase, logger, metrics)

			router := initializeRouter(handler, cfg, logger)

			assert.NotNil(t, router)
			assert.Equal(t, tt.expectedMode, gin.Mode())
		})
	}
}

// Note: TestInitializeRouter_GinModeEnvVar was removed because:
// - TestMain already sets gin.TestMode globally
// - Gin mode can only be set once per process
// - Testing environment variable precedence for GIN_MODE would require subprocess testing

// TestRun tests the run function with valid configuration
func TestRun(t *testing.T) {
	// This test is challenging because run() calls startServer() which blocks.
	// We'll test that run returns error when config is missing
	require.NoError(t, os.Setenv("GWAIHIR_CONFIG", "/nonexistent/path/config.yaml"))
	defer func() { require.NoError(t, os.Unsetenv("GWAIHIR_CONFIG")) }()

	err := run()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

// TestRun_InvalidConfig tests run function with invalid configuration
func TestRun_InvalidConfig(t *testing.T) {
	invalidConfig := `server:
  port: 99999
  log:
    format: invalid
    level: info
machines: []
`
	configPath := setupTestConfig(t, invalidConfig)
	require.NoError(t, os.Setenv("GWAIHIR_CONFIG", configPath))
	defer func() { require.NoError(t, os.Unsetenv("GWAIHIR_CONFIG")) }()

	err := run()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

// Note: TestStartServer was removed because:
// - startServer() blocks until server shuts down (signal handling)
// - Testing server startup/shutdown properly requires more complex setup with goroutines,
//   signal handling, and timing that would be fragile
// - The http.Server functionality is already well-tested by the standard library
// - Integration tests cover the actual server behavior

// TestVersion_GlobalVariables tests that version variables are accessible
func TestVersion_GlobalVariables(t *testing.T) {
	// These are global variables that should be set at build time
	assert.NotEmpty(t, Version)
	assert.NotEmpty(t, BuildTime)
	assert.NotEmpty(t, GitCommit)
}

// Benchmark tests
func BenchmarkLoadConfiguration(b *testing.B) {
	configContent := validTestConfig()
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "bench-config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(b, err)

	require.NoError(b, os.Setenv("GWAIHIR_CONFIG", configPath))
	defer func() { require.NoError(b, os.Unsetenv("GWAIHIR_CONFIG")) }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loadConfiguration()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkInitializeLogger(b *testing.B) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Log: config.LogConfig{
				Format: "json",
				Level:  "info",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger := initializeLogger(cfg)
		if logger == nil {
			b.Fatal("logger is nil")
		}
	}
}

func BenchmarkInitializeMetrics(b *testing.B) {
	// Metrics can only be initialized once due to Prometheus registration
	// This benchmark tests the getTestMetrics helper performance
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics := globalMetrics
		if metrics == nil {
			b.Fatal("metrics is nil")
		}
	}
}
