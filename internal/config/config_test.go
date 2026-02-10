package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const basicConfigContent = `
server:
  port: 8080
  log:
    format: text
    level: info
authentication:
  api_key: "key"
machines:
  - id: m1
    name: "M"
    mac: "00:11:22:33:44:55"
    broadcast: "192.168.1.255"
observability:
  health_check:
    enabled: true
  metrics:
    enabled: true
`

const fileKeyConfigContent = `
server:
  port: 8080
  log:
    format: text
    level: info
authentication:
  api_key: "file-key"
machines:
  - id: m1
    name: "M"
    mac: "00:11:22:33:44:55"
    broadcast: "192.168.1.255"
observability:
  health_check:
    enabled: true
  metrics:
    enabled: true
`

func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	assert.NoError(t, err)

	filename := tmpFile.Name()
	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	assert.NoError(t, tmpFile.Close())

	t.Cleanup(func() {
		_ = os.Remove(filename)
	})

	return filename
}

func TestConfig_UnmarshalYAML_ValidConfig(t *testing.T) {
	yamlData := `
server:
  port: 8080
  log:
    format: json
    level: info
authentication:
  api_key: "test-key"
machines:
  - id: machine1
    name: "Machine 1"
    mac: "00:11:22:33:44:55"
    broadcast: "192.168.1.255"
observability:
  health_check:
    enabled: true
  metrics:
    enabled: true
`

	var cfg Config
	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	assert.NoError(t, err)

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "json", cfg.Server.Log.Format)
	assert.Equal(t, "info", cfg.Server.Log.Level)
	assert.Equal(t, "test-key", cfg.Authentication.APIKey)
	assert.Len(t, cfg.Machines, 1)
	assert.Equal(t, "machine1", cfg.Machines[0].ID)
	assert.Equal(t, "00:11:22:33:44:55", cfg.Machines[0].MAC)
	assert.Equal(t, "192.168.1.255", cfg.Machines[0].Broadcast)
	assert.Equal(t, boolPtr(true), cfg.Observability.HealthCheck.Enabled)
	assert.Equal(t, boolPtr(true), cfg.Observability.Metrics.Enabled)
}

func TestConfig_UnmarshalYAML_AllFields(t *testing.T) {
	yamlData := `
server:
  port: 9090
  log:
    format: text
    level: debug
authentication:
  api_key: "secret"
machines:
  - id: server1
    name: "Production Server"
    mac: "AA:BB:CC:DD:EE:FF"
    broadcast: "10.0.0.255"
  - id: server2
    name: "Backup Server"
    mac: "11:22:33:44:55:66"
    broadcast: "10.0.0.255"
observability:
  health_check:
    enabled: false
  metrics:
    enabled: false
`

	var cfg Config
	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	assert.NoError(t, err)

	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "text", cfg.Server.Log.Format)
	assert.Equal(t, "debug", cfg.Server.Log.Level)
	assert.Equal(t, "secret", cfg.Authentication.APIKey)
	assert.Len(t, cfg.Machines, 2)
	assert.Equal(t, boolPtr(false), cfg.Observability.HealthCheck.Enabled)
	assert.Equal(t, boolPtr(false), cfg.Observability.Metrics.Enabled)
}

func TestConfig_UnmarshalYAML_MinimalConfig(t *testing.T) {
	yamlData := `
server:
  port: 8080
  log:
    format: text
    level: info
authentication:
  api_key: "key"
machines:
  - id: m1
    name: "Machine"
    mac: "00:11:22:33:44:55"
    broadcast: "192.168.1.255"
observability:
  health_check:
    enabled: true
  metrics:
    enabled: true
`

	var cfg Config
	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	assert.NoError(t, err)

	// Verify structure is intact
	assert.NotNil(t, cfg.Server)
	assert.NotNil(t, cfg.Server.Log)
	assert.NotNil(t, cfg.Authentication)
	assert.NotNil(t, cfg.Machines)
	assert.NotNil(t, cfg.Observability)
	// Note: with pointer types, these will be non-nil after YAML parsing
	// The defaults are applied later in LoadConfig
}

func TestConfig_UnmarshalYAML_EmptyMachinesList(t *testing.T) {
	yamlData := `
server:
  port: 8080
  log:
    format: text
    level: info
authentication:
  api_key: "key"
machines: []
observability:
  health_check:
    enabled: true
  metrics:
    enabled: true
`

	var cfg Config
	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	assert.NoError(t, err)
	assert.Len(t, cfg.Machines, 0)
}

func TestConfig_UnmarshalYAML_NoAPIKey(t *testing.T) {
	yamlData := `
server:
  port: 8080
  log:
    format: text
    level: info
authentication: {}
machines:
  - id: m1
    name: "Machine"
    mac: "00:11:22:33:44:55"
    broadcast: "192.168.1.255"
observability:
  health_check:
    enabled: true
  metrics:
    enabled: true
`

	var cfg Config
	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "", cfg.Authentication.APIKey)
}

// Helper to compare bool pointers
func boolPtr(v bool) *bool {
	return &v
}

func TestLoadConfig_ValidFile(t *testing.T) {
	configContent := `
server:
  port: 8080
  log:
    format: json
    level: info
authentication:
  api_key: "test-key"
machines:
  - id: machine1
    name: "Machine 1"
    mac: "00:11:22:33:44:55"
    broadcast: "192.168.1.255"
observability:
  health_check:
    enabled: true
  metrics:
    enabled: true
`

	filename := createTempConfigFile(t, configContent)

	cfg, err := LoadConfig(filename)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "test-key", cfg.Authentication.APIKey)
	assert.Len(t, cfg.Machines, 1)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	nonExistentPath := "/tmp/non-existent-gwaihir-config-12345.yaml"

	cfg, err := LoadConfig(nonExistentPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.True(t, errors.Is(err, os.ErrNotExist) || os.IsNotExist(err))
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	invalidYAML := `
server:
  port: 8080
  log:
    format: json
    level: info
invalid: [unclosed
`

	filename := createTempConfigFile(t, invalidYAML)

	cfg, err := LoadConfig(filename)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	filename := createTempConfigFile(t, "")

	cfg, err := LoadConfig(filename)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "invalid configuration")
}

func TestLoadConfig_ExampleConfigFile(t *testing.T) {
	examplePath := filepath.Join("..", "..", "configs", "gwaihir.example.yaml")
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		t.Skip("Example config file not found, skipping integration test")
	}

	cfg, err := LoadConfig(examplePath)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Greater(t, cfg.Server.Port, 0)
	assert.NotEmpty(t, cfg.Server.Log.Format)
	assert.NotEmpty(t, cfg.Server.Log.Level)
	assert.Len(t, cfg.Machines, 3)
}

func TestLoadConfig_PortEnvOverride(t *testing.T) {
	filename := createTempConfigFile(t, basicConfigContent)

	t.Setenv("GWAIHIR_PORT", "9090")

	cfg, err := LoadConfig(filename)
	assert.NoError(t, err)
	assert.Equal(t, 9090, cfg.Server.Port)
}

func TestLoadConfig_InvalidPortEnvOverride(t *testing.T) {
	filename := createTempConfigFile(t, basicConfigContent)

	t.Setenv("GWAIHIR_PORT", "not-a-number")

	cfg, err := LoadConfig(filename)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "GWAIHIR_PORT")
	assert.Contains(t, err.Error(), "not-a-number")
}

func TestLoadConfig_LogLevelEnvOverride(t *testing.T) {
	filename := createTempConfigFile(t, basicConfigContent)

	t.Setenv("GWAIHIR_LOG_LEVEL", "debug")

	cfg, err := LoadConfig(filename)
	assert.NoError(t, err)
	assert.Equal(t, "debug", cfg.Server.Log.Level)
}

func TestLoadConfig_LogFormatEnvOverride(t *testing.T) {
	filename := createTempConfigFile(t, basicConfigContent)

	t.Setenv("GWAIHIR_LOG_FORMAT", "json")

	cfg, err := LoadConfig(filename)
	assert.NoError(t, err)
	assert.Equal(t, "json", cfg.Server.Log.Format)
}

func TestLoadConfig_APIKeyEnvOverride(t *testing.T) {
	filename := createTempConfigFile(t, fileKeyConfigContent)

	t.Setenv("GWAIHIR_API_KEY", "env-key")

	cfg, err := LoadConfig(filename)
	assert.NoError(t, err)
	assert.Equal(t, "env-key", cfg.Authentication.APIKey)
}

func TestLoadConfig_PrecedenceEnvOverFile(t *testing.T) {
	filename := createTempConfigFile(t, fileKeyConfigContent)

	// Set multiple env vars
	t.Setenv("GWAIHIR_PORT", "9090")
	t.Setenv("GWAIHIR_LOG_LEVEL", "debug")
	t.Setenv("GWAIHIR_LOG_FORMAT", "json")
	t.Setenv("GWAIHIR_API_KEY", "env-key")

	cfg, err := LoadConfig(filename)
	assert.NoError(t, err)

	// All env vars should override file values
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Server.Log.Level)
	assert.Equal(t, "json", cfg.Server.Log.Format)
	assert.Equal(t, "env-key", cfg.Authentication.APIKey)
}

func TestLoadConfig_NoEnvOverrideWhenUnset(t *testing.T) {
	filename := createTempConfigFile(t, fileKeyConfigContent)

	// Ensure env vars are NOT set
	t.Setenv("GWAIHIR_PORT", "")
	t.Setenv("GWAIHIR_LOG_LEVEL", "")
	t.Setenv("GWAIHIR_LOG_FORMAT", "")
	t.Setenv("GWAIHIR_API_KEY", "")

	cfg, err := LoadConfig(filename)
	assert.NoError(t, err)

	// File values should be used
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "info", cfg.Server.Log.Level)
	assert.Equal(t, "text", cfg.Server.Log.Format)
	assert.Equal(t, "file-key", cfg.Authentication.APIKey)
}

func TestLoadConfig_DefaultPortWhenMissing(t *testing.T) {
	minimalConfig := `
server:
  log:
    format: text
    level: info
authentication:
  api_key: "key"
machines:
  - id: m1
    name: "M"
    mac: "00:11:22:33:44:55"
    broadcast: "192.168.1.255"
observability:
  health_check:
    enabled: true
  metrics:
    enabled: true
`

	filename := createTempConfigFile(t, minimalConfig)

	cfg, err := LoadConfig(filename)
	assert.NoError(t, err)
	assert.Equal(t, 8080, cfg.Server.Port)
}

func TestLoadConfig_DefaultLogFormatWhenMissing(t *testing.T) {
	minimalConfig := `
server:
  port: 8080
  log:
    level: info
authentication:
  api_key: "key"
machines:
  - id: m1
    name: "M"
    mac: "00:11:22:33:44:55"
    broadcast: "192.168.1.255"
observability:
  health_check:
    enabled: true
  metrics:
    enabled: true
`

	filename := createTempConfigFile(t, minimalConfig)

	cfg, err := LoadConfig(filename)
	assert.NoError(t, err)
	assert.Equal(t, "text", cfg.Server.Log.Format)
}

func TestLoadConfig_DefaultLogLevelWhenMissing(t *testing.T) {
	minimalConfig := `
server:
  port: 8080
  log:
    format: text
authentication:
  api_key: "key"
machines:
  - id: m1
    name: "M"
    mac: "00:11:22:33:44:55"
    broadcast: "192.168.1.255"
observability:
  health_check:
    enabled: true
  metrics:
    enabled: true
`

	filename := createTempConfigFile(t, minimalConfig)

	cfg, err := LoadConfig(filename)
	assert.NoError(t, err)
	assert.Equal(t, "info", cfg.Server.Log.Level)
}

func TestLoadConfig_DefaultObservabilityEnabled(t *testing.T) {
	minimalConfig := `
server:
  port: 8080
  log:
    format: text
    level: info
authentication:
  api_key: "key"
machines:
  - id: m1
    name: "M"
    mac: "00:11:22:33:44:55"
    broadcast: "192.168.1.255"
observability: {}
`

	filename := createTempConfigFile(t, minimalConfig)

	cfg, err := LoadConfig(filename)
	assert.NoError(t, err)
	assert.Equal(t, boolPtr(true), cfg.Observability.HealthCheck.Enabled)
	assert.Equal(t, boolPtr(true), cfg.Observability.Metrics.Enabled)
}

func TestLoadConfig_DefaultsNotOverridingFileValues(t *testing.T) {
	configContent := `
server:
  port: 9090
  log:
    format: json
    level: warn
authentication:
  api_key: "key"
machines:
  - id: m1
    name: "M"
    mac: "00:11:22:33:44:55"
    broadcast: "192.168.1.255"
observability:
  health_check:
    enabled: false
  metrics:
    enabled: false
`

	filename := createTempConfigFile(t, configContent)

	cfg, err := LoadConfig(filename)
	assert.NoError(t, err)

	// File values should be respected, not defaults
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "json", cfg.Server.Log.Format)
	assert.Equal(t, "warn", cfg.Server.Log.Level)
	assert.Equal(t, boolPtr(false), cfg.Observability.HealthCheck.Enabled)
	assert.Equal(t, boolPtr(false), cfg.Observability.Metrics.Enabled)
}

func TestLoadConfig_DefaultsNotOverridingEnvValues(t *testing.T) {
	minimalConfig := `
server:
  log:
    format: text
    level: info
authentication:
  api_key: "key"
machines:
  - id: m1
    name: "M"
    mac: "00:11:22:33:44:55"
    broadcast: "192.168.1.255"
observability:
  health_check:
    enabled: true
  metrics:
    enabled: true
`

	filename := createTempConfigFile(t, minimalConfig)

	// Set env vars that should override defaults
	t.Setenv("GWAIHIR_PORT", "9999")
	t.Setenv("GWAIHIR_LOG_FORMAT", "json")
	t.Setenv("GWAIHIR_LOG_LEVEL", "debug")

	cfg, err := LoadConfig(filename)
	assert.NoError(t, err)

	// Env values should take precedence
	assert.Equal(t, 9999, cfg.Server.Port)
	assert.Equal(t, "json", cfg.Server.Log.Format)
	assert.Equal(t, "debug", cfg.Server.Log.Level)
}

// Validation tests

func TestConfig_Validate_ValidConfig(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Log: LogConfig{
				Format: "json",
				Level:  "info",
			},
		},
		Authentication: AuthenticationConfig{
			APIKey: "test-key",
		},
		Machines: []MachineConfig{
			{
				ID:        "m1",
				Name:      "Machine 1",
				MAC:       "00:11:22:33:44:55",
				Broadcast: "192.168.1.255",
			},
		},
		Observability: ObservabilityConfig{
			HealthCheck: HealthCheckConfig{Enabled: boolPtr(true)},
			Metrics:     MetricsConfig{Enabled: boolPtr(true)},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_PortOutOfRange_Low(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 0,
			Log: LogConfig{
				Format: "text",
				Level:  "info",
			},
		},
		Authentication: AuthenticationConfig{APIKey: "key"},
		Machines: []MachineConfig{
			{ID: "m1", Name: "M", MAC: "00:11:22:33:44:55", Broadcast: "192.168.1.255"},
		},
		Observability: ObservabilityConfig{
			HealthCheck: HealthCheckConfig{Enabled: boolPtr(true)},
			Metrics:     MetricsConfig{Enabled: boolPtr(true)},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port")
}

func TestConfig_Validate_PortOutOfRange_High(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 65536,
			Log: LogConfig{
				Format: "text",
				Level:  "info",
			},
		},
		Authentication: AuthenticationConfig{APIKey: "key"},
		Machines: []MachineConfig{
			{ID: "m1", Name: "M", MAC: "00:11:22:33:44:55", Broadcast: "192.168.1.255"},
		},
		Observability: ObservabilityConfig{
			HealthCheck: HealthCheckConfig{Enabled: boolPtr(true)},
			Metrics:     MetricsConfig{Enabled: boolPtr(true)},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port")
}

func TestConfig_Validate_InvalidLogFormat(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Log: LogConfig{
				Format: "invalid",
				Level:  "info",
			},
		},
		Authentication: AuthenticationConfig{APIKey: "key"},
		Machines: []MachineConfig{
			{ID: "m1", Name: "M", MAC: "00:11:22:33:44:55", Broadcast: "192.168.1.255"},
		},
		Observability: ObservabilityConfig{
			HealthCheck: HealthCheckConfig{Enabled: boolPtr(true)},
			Metrics:     MetricsConfig{Enabled: boolPtr(true)},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "format")
}

func TestConfig_Validate_InvalidLogLevel(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Log: LogConfig{
				Format: "json",
				Level:  "invalid",
			},
		},
		Authentication: AuthenticationConfig{APIKey: "key"},
		Machines: []MachineConfig{
			{ID: "m1", Name: "M", MAC: "00:11:22:33:44:55", Broadcast: "192.168.1.255"},
		},
		Observability: ObservabilityConfig{
			HealthCheck: HealthCheckConfig{Enabled: boolPtr(true)},
			Metrics:     MetricsConfig{Enabled: boolPtr(true)},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "level")
}

func TestConfig_Validate_APIKeyOptional(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Log: LogConfig{
				Format: "text",
				Level:  "info",
			},
		},
		Authentication: AuthenticationConfig{APIKey: ""},
		Machines: []MachineConfig{
			{ID: "m1", Name: "M", MAC: "00:11:22:33:44:55", Broadcast: "192.168.1.255"},
		},
		Observability: ObservabilityConfig{
			HealthCheck: HealthCheckConfig{Enabled: boolPtr(true)},
			Metrics:     MetricsConfig{Enabled: boolPtr(true)},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_MachinesRequired(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Log: LogConfig{
				Format: "text",
				Level:  "info",
			},
		},
		Authentication: AuthenticationConfig{APIKey: "key"},
		Machines:       []MachineConfig{},
		Observability: ObservabilityConfig{
			HealthCheck: HealthCheckConfig{Enabled: boolPtr(true)},
			Metrics:     MetricsConfig{Enabled: boolPtr(true)},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "machine")
}

func TestConfig_Validate_InvalidMachineMAC(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Log: LogConfig{
				Format: "text",
				Level:  "info",
			},
		},
		Authentication: AuthenticationConfig{APIKey: "key"},
		Machines: []MachineConfig{
			{ID: "m1", Name: "M", MAC: "invalid-mac", Broadcast: "192.168.1.255"},
		},
		Observability: ObservabilityConfig{
			HealthCheck: HealthCheckConfig{Enabled: boolPtr(true)},
			Metrics:     MetricsConfig{Enabled: boolPtr(true)},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MAC")
}

func TestConfig_Validate_InvalidMachineBroadcast(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Log: LogConfig{
				Format: "text",
				Level:  "info",
			},
		},
		Authentication: AuthenticationConfig{APIKey: "key"},
		Machines: []MachineConfig{
			{ID: "m1", Name: "M", MAC: "00:11:22:33:44:55", Broadcast: "not-an-ip"},
		},
		Observability: ObservabilityConfig{
			HealthCheck: HealthCheckConfig{Enabled: boolPtr(true)},
			Metrics:     MetricsConfig{Enabled: boolPtr(true)},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "broadcast")
}

func TestConfig_Validate_AllLogLevelsValid(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		cfg := &Config{
			Server: ServerConfig{
				Port: 8080,
				Log: LogConfig{
					Format: "text",
					Level:  level,
				},
			},
			Authentication: AuthenticationConfig{APIKey: "key"},
			Machines: []MachineConfig{
				{ID: "m1", Name: "M", MAC: "00:11:22:33:44:55", Broadcast: "192.168.1.255"},
			},
			Observability: ObservabilityConfig{
				HealthCheck: HealthCheckConfig{Enabled: boolPtr(true)},
				Metrics:     MetricsConfig{Enabled: boolPtr(true)},
			},
		}

		err := cfg.Validate()
		assert.NoError(t, err, "level %s should be valid", level)
	}
}

func TestConfig_Validate_AllLogFormatsValid(t *testing.T) {
	formats := []string{"json", "text"}

	for _, format := range formats {
		cfg := &Config{
			Server: ServerConfig{
				Port: 8080,
				Log: LogConfig{
					Format: format,
					Level:  "info",
				},
			},
			Authentication: AuthenticationConfig{APIKey: "key"},
			Machines: []MachineConfig{
				{ID: "m1", Name: "M", MAC: "00:11:22:33:44:55", Broadcast: "192.168.1.255"},
			},
			Observability: ObservabilityConfig{
				HealthCheck: HealthCheckConfig{Enabled: boolPtr(true)},
				Metrics:     MetricsConfig{Enabled: boolPtr(true)},
			},
		}

		err := cfg.Validate()
		assert.NoError(t, err, "format %s should be valid", format)
	}
}

func TestConfig_Validate_MultipleMachines(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Log: LogConfig{
				Format: "text",
				Level:  "info",
			},
		},
		Authentication: AuthenticationConfig{APIKey: "key"},
		Machines: []MachineConfig{
			{ID: "m1", Name: "M1", MAC: "00:11:22:33:44:55", Broadcast: "192.168.1.255"},
			{ID: "m2", Name: "M2", MAC: "AA:BB:CC:DD:EE:FF", Broadcast: "192.168.1.255"},
			{ID: "m3", Name: "M3", MAC: "11:22:33:44:55:66", Broadcast: "10.0.0.255"},
		},
		Observability: ObservabilityConfig{
			HealthCheck: HealthCheckConfig{Enabled: boolPtr(true)},
			Metrics:     MetricsConfig{Enabled: boolPtr(true)},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}
