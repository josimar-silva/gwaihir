package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/josimar-silva/gwaihir/internal/config"
)

func TestConfig_LoadAndValidate_ExampleConfig(t *testing.T) {
	examplePath := filepath.Join("..", "configs", "gwaihir.example.yaml")

	// Load the example configuration
	cfg, err := config.LoadConfig(examplePath)
	assert.NoError(t, err, "failed to load example config")
	assert.NotNil(t, cfg)

	// Verify it loaded all sections
	assert.NotNil(t, cfg.Server)
	assert.NotNil(t, cfg.Authentication)
	assert.NotNil(t, cfg.Machines)
	assert.NotNil(t, cfg.Observability)

	// Verify server settings
	assert.Greater(t, cfg.Server.Port, 0, "port should be set")
	assert.NotEmpty(t, cfg.Server.Log.Format, "log format should be set")
	assert.NotEmpty(t, cfg.Server.Log.Level, "log level should be set")

	// Verify authentication
	assert.NotEmpty(t, cfg.Authentication.APIKey, "api_key should be set")

	// Verify machines
	assert.Greater(t, len(cfg.Machines), 0, "at least one machine should be configured")
	for _, machine := range cfg.Machines {
		assert.NotEmpty(t, machine.ID, "machine ID should not be empty")
		assert.NotEmpty(t, machine.Name, "machine name should not be empty")
		assert.NotEmpty(t, machine.MAC, "machine MAC should not be empty")
		assert.NotEmpty(t, machine.Broadcast, "machine broadcast should not be empty")
	}

	// Verify observability
	assert.NotNil(t, cfg.Observability.HealthCheck.Enabled, "health check enabled should be set")
	assert.NotNil(t, cfg.Observability.Metrics.Enabled, "metrics enabled should be set")

	// Validate the loaded configuration
	err = cfg.Validate()
	assert.NoError(t, err, "example config should validate successfully")
}

func TestConfig_LoadValidateDefaults_FullFlow(t *testing.T) {
	examplePath := filepath.Join("..", "configs", "gwaihir.example.yaml")

	// Full workflow: load → defaults → validate
	cfg, err := config.LoadConfig(examplePath)
	assert.NoError(t, err)

	// After LoadConfig, defaults should be applied
	// Observability fields should have pointer values (non-nil after defaults)
	assert.NotNil(t, cfg.Observability.HealthCheck.Enabled)
	assert.NotNil(t, cfg.Observability.Metrics.Enabled)

	// Configuration should be valid
	err = cfg.Validate()
	assert.NoError(t, err)

	// Verify all machines in example config
	assert.Len(t, cfg.Machines, 3, "example config should have 3 machines")

	// Verify first machine details
	assert.Equal(t, "saruman", cfg.Machines[0].ID)
	assert.NotEmpty(t, cfg.Machines[0].MAC)
	assert.NotEmpty(t, cfg.Machines[0].Broadcast)

	// Verify second machine details
	assert.Equal(t, "gandalf", cfg.Machines[1].ID)
	assert.NotEmpty(t, cfg.Machines[1].MAC)
	assert.NotEmpty(t, cfg.Machines[1].Broadcast)

	// Verify third machine details
	assert.Equal(t, "radagast", cfg.Machines[2].ID)
	assert.NotEmpty(t, cfg.Machines[2].MAC)
	assert.NotEmpty(t, cfg.Machines[2].Broadcast)
}

func TestConfig_Example_Valid(t *testing.T) {
	// Test that example config is a valid template users can work from
	examplePath := filepath.Join("..", "configs", "gwaihir.example.yaml")

	cfg, err := config.LoadConfig(examplePath)
	assert.NoError(t, err, "example config should load without errors")

	err = cfg.Validate()
	assert.NoError(t, err, "example config should pass validation")

	// Verify it has reasonable defaults applied
	assert.Greater(t, cfg.Server.Port, 0, "example should have a port")
	assert.Contains(t, []string{"json", "text"}, cfg.Server.Log.Format, "log format should be valid")
	assert.Contains(t, []string{"debug", "info", "warn", "error"}, cfg.Server.Log.Level, "log level should be valid")
}

func TestConfig_MachineValidation_Integration(t *testing.T) {
	examplePath := filepath.Join("..", "configs", "gwaihir.example.yaml")

	cfg, err := config.LoadConfig(examplePath)
	assert.NoError(t, err)

	// Validate that all machines in the example have valid configurations
	for i, machine := range cfg.Machines {
		// Each machine should have required fields
		assert.NotEmpty(t, machine.ID, "machine %d ID should not be empty", i)
		assert.NotEmpty(t, machine.Name, "machine %d name should not be empty", i)
		assert.NotEmpty(t, machine.MAC, "machine %d MAC should not be empty", i)
		assert.NotEmpty(t, machine.Broadcast, "machine %d broadcast should not be empty", i)

		// MAC format validation is done in Config.Validate()
		// This test just confirms all machines pass through validation
	}

	// The full config should validate without errors
	err = cfg.Validate()
	assert.NoError(t, err)
}
