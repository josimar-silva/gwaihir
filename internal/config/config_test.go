package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

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
	assert.True(t, cfg.Observability.HealthCheck.Enabled)
	assert.True(t, cfg.Observability.Metrics.Enabled)
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
	assert.False(t, cfg.Observability.HealthCheck.Enabled)
	assert.False(t, cfg.Observability.Metrics.Enabled)
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
	assert.NotNil(t, cfg.Observability.HealthCheck)
	assert.NotNil(t, cfg.Observability.Metrics)
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

func TestLoadConfig_ValidFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	assert.NoError(t, err)

	filename := tmpFile.Name()
	defer func() {
		_ = os.Remove(filename)
	}()

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

	_, err = tmpFile.WriteString(configContent)
	assert.NoError(t, err)
	assert.NoError(t, tmpFile.Close())

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
	tmpFile, err := os.CreateTemp("", "config-invalid-*.yaml")
	assert.NoError(t, err)

	filename := tmpFile.Name()
	defer func() {
		_ = os.Remove(filename)
	}()

	invalidYAML := `
server:
  port: 8080
  log:
    format: json
    level: info
invalid: [unclosed
`

	_, err = tmpFile.WriteString(invalidYAML)
	assert.NoError(t, err)
	assert.NoError(t, tmpFile.Close())

	cfg, err := LoadConfig(filename)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config-empty-*.yaml")
	assert.NoError(t, err)

	filename := tmpFile.Name()
	defer func() {
		_ = os.Remove(filename)
	}()
	assert.NoError(t, tmpFile.Close())

	cfg, err := LoadConfig(filename)
	// Empty file should deserialize to empty struct, not error
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
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
