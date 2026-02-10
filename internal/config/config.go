package config

import (
	"fmt"
	"net"
	"os"
	"regexp"
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

	setDefaults(&cfg)

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

// setDefaults applies sensible default values for optional configuration fields.
// Defaults are only applied when values are not already set.
func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	if cfg.Server.Log.Format == "" {
		cfg.Server.Log.Format = "text"
	}

	if cfg.Server.Log.Level == "" {
		cfg.Server.Log.Level = "info"
	}

	if cfg.Observability.HealthCheck.Enabled == nil {
		trueVal := true
		cfg.Observability.HealthCheck.Enabled = &trueVal
	}

	if cfg.Observability.Metrics.Enabled == nil {
		trueVal := true
		cfg.Observability.Metrics.Enabled = &trueVal
	}
}

// Validate validates all configuration fields and returns an error if any validation fails.
// Validation checks:
// - server.port: must be in range 1-65535
// - server.log.format: must be "json" or "text"
// - server.log.level: must be "debug", "info", "warn", or "error"
// - authentication.api_key: must not be empty
// - machines: must have at least 1 machine, each must be valid (MAC, broadcast IP)
func (cfg *Config) Validate() error {
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: must be between 1 and 65535, got %d", cfg.Server.Port)
	}

	if err := validateLogFormat(cfg.Server.Log.Format); err != nil {
		return err
	}

	if err := validateLogLevel(cfg.Server.Log.Level); err != nil {
		return err
	}

	if cfg.Authentication.APIKey == "" {
		return fmt.Errorf("authentication.api_key is required and cannot be empty")
	}

	if len(cfg.Machines) == 0 {
		return fmt.Errorf("at least one machine must be configured")
	}

	for i, machine := range cfg.Machines {
		if err := validateMachine(machine); err != nil {
			return fmt.Errorf("machine %d (%s): %w", i, machine.ID, err)
		}
	}

	return nil
}

func validateLogFormat(format string) error {
	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[format] {
		return fmt.Errorf("invalid server.log.format: must be 'json' or 'text', got '%s'", format)
	}
	return nil
}

func validateLogLevel(level string) error {
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[level] {
		return fmt.Errorf("invalid server.log.level: must be 'debug', 'info', 'warn', or 'error', got '%s'", level)
	}
	return nil
}

func validateMachine(machine MachineConfig) error {
	if !isValidMAC(machine.MAC) {
		return fmt.Errorf("invalid MAC address format: '%s' (must be XX:XX:XX:XX:XX:XX)", machine.MAC)
	}

	if !isValidIP(machine.Broadcast) {
		return fmt.Errorf("invalid broadcast IP address: '%s' (must be a valid IPv4 address)", machine.Broadcast)
	}

	return nil
}

func isValidMAC(mac string) bool {
	pattern := `^([0-9A-Fa-f]{2}[:]){5}([0-9A-Fa-f]{2})$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(mac)
}

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
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
	Enabled *bool `yaml:"enabled"`
}

// MetricsConfig controls metrics endpoint exposure.
type MetricsConfig struct {
	Enabled *bool `yaml:"enabled"`
}
