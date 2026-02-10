package config

import (
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
