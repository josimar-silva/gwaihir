package config

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
