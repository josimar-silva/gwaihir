package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads and parses the configuration from a YAML file.
// Returns a pointer to Config if successful, or an error if the file
// cannot be read or contains invalid YAML.
// Environment variables override file values with this precedence:
//   - GWAIHIR_PORT overrides server.port
//   - GWAIHIR_LOG_FORMAT overrides server.log.format
//   - GWAIHIR_LOG_LEVEL overrides server.log.level
//   - GWAIHIR_API_KEY overrides authentication.api_key
//
// #nosec G304 - path is controlled by application, not user input
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(&cfg)

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides to the configuration.
// Environment variables take precedence over file values.
func applyEnvOverrides(cfg *Config) {
	if port := os.Getenv("GWAIHIR_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}

	if format := os.Getenv("GWAIHIR_LOG_FORMAT"); format != "" {
		cfg.Server.Log.Format = format
	}

	if level := os.Getenv("GWAIHIR_LOG_LEVEL"); level != "" {
		cfg.Server.Log.Level = level
	}

	if apiKey := os.Getenv("GWAIHIR_API_KEY"); apiKey != "" {
		cfg.Authentication.APIKey = apiKey
	}
}

// Config represents the complete unified configuration for Gwaihir.
// It contains all application settings in a single structure that can be
// loaded from YAML with environment variable overrides.
type Config struct {
	Server         ServerConfig         `yaml:"server"`
	Authentication AuthenticationConfig `yaml:"authentication"`
	Machines       []MachineConfig      `yaml:"machines"`
	Observability  ObservabilityConfig  `yaml:"observability"`
}

// ServerConfig contains HTTP server configuration.
type ServerConfig struct {
	Port int       `yaml:"port"`
	Log  LogConfig `yaml:"log"`
}

// LogConfig contains logging configuration.
type LogConfig struct {
	Format string `yaml:"format"` // json or text
	Level  string `yaml:"level"`  // debug, info, warn, error
}

// AuthenticationConfig contains authentication settings.
type AuthenticationConfig struct {
	APIKey string `yaml:"api_key"`
}

// MachineConfig represents a machine that can receive WoL packets.
type MachineConfig struct {
	ID        string `yaml:"id"`
	Name      string `yaml:"name"`
	MAC       string `yaml:"mac"`
	Broadcast string `yaml:"broadcast"`
}

// ObservabilityConfig contains observability settings.
type ObservabilityConfig struct {
	HealthCheck HealthCheckConfig `yaml:"health_check"`
	Metrics     MetricsConfig     `yaml:"metrics"`
}

// HealthCheckConfig controls health check endpoint exposure.
type HealthCheckConfig struct {
	Enabled bool `yaml:"enabled"`
}

// MetricsConfig controls metrics endpoint exposure.
type MetricsConfig struct {
	Enabled bool `yaml:"enabled"`
}
