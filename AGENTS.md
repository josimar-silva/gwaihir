# AGENTS.md

This document provides guidance for AI coding agents working with the Gwaihir codebase.

## Project Overview

**Gwaihir** is a production-ready Wake-on-LAN (WoL) microservice written in Go that sends WoL packets to wake up sleeping servers. It's designed to work with [Smaug](https://github.com/josimar-silva/smaug), a reverse proxy that commands Gwaihir to wake servers on demand.

**Key Characteristics:**
- Language: Go 1.23+
- Framework: Gin (HTTP router)
- Architecture: Clean Architecture with clear layer separation
- Logging: Structured JSON logging with `log/slog`
- Metrics: Prometheus metrics
- Test Coverage: 90%+ across all layers

## Architecture

Gwaihir follows **Clean Architecture** principles with strict separation of concerns:

```
internal/
├── domain/              # Business entities & interfaces (core)
│   ├── machine.go       # Machine entity with validation
│   ├── repository.go    # Repository interface
│   └── errors.go        # Domain-specific errors
├── usecase/             # Business logic layer
│   └── wol_usecase.go   # WoL operations orchestration
├── delivery/http/       # HTTP presentation layer (Gin)
│   ├── handler.go       # HTTP request handlers
│   ├── router.go        # Route configuration
│   ├── auth.go          # API key authentication
│   ├── middleware.go    # Request logging & correlation
│   └── health.go        # Health check handlers
├── repository/          # Data access implementations
│   └── yaml_machine_repository.go
└── infrastructure/      # Infrastructure concerns
    ├── logger.go        # Structured logging wrapper
    ├── metrics.go       # Prometheus metrics
    └── wol_packet.go    # WoL packet sender
```

### Dependency Rules

Follow these strict dependency rules:
- **Domain** has NO dependencies on other layers
- **Use Case** depends only on Domain
- **Repository** implements Domain interfaces
- **Infrastructure** can depend on Domain
- **Delivery** depends on Use Case and Domain

```
Delivery → Use Case → Domain ← Repository
            ↓                    ↓
      Infrastructure ← → Repository
```

## Code Conventions

### File Organization

- Each layer has its own package
- Test files use `_test.go` suffix and live alongside source files
- Mocks are generated in the same package they mock
- Integration tests live in `tests/` directory

### Naming Conventions

- **Interfaces:** Descriptive names (e.g., `MachineRepository`, `WoLUseCase`)
- **Implementations:** Include implementation detail (e.g., `InMemoryMachineRepository`)
- **Methods:** Clear action verbs (e.g., `SendWoLPacket`, `GetMachine`, `ListMachines`)
- **Variables:** Descriptive, avoid single letters except in very short scopes
- **Errors:** Use domain-specific errors defined in `domain/errors.go`

### Error Handling

```go
// Domain errors (domain/errors.go)
var (
    ErrMachineNotFound = errors.New("machine not found")
    ErrInvalidMAC      = errors.New("invalid MAC address")
)

// Always wrap errors with context
return fmt.Errorf("failed to send WoL packet: %w", err)

// Check for specific errors
if errors.Is(err, domain.ErrMachineNotFound) {
    // Handle not found
}
```

### Logging

Always use structured logging with `log/slog`:

```go
// Good: Structured logging with context
logger.Info("Sending WoL packet",
    slog.String("machine_id", machineID),
    slog.String("mac", machine.MAC),
    slog.String("broadcast", machine.Broadcast),
)

// Bad: Unstructured logging
logger.Info("Sending WoL packet to " + machineID)
```

### Metrics

Record metrics for important operations:

```go
// Counter for operations
metrics.WoLPacketsSentTotal.Inc()
metrics.WoLPacketsFailedTotal.Inc()

// Histogram for durations
timer := prometheus.NewTimer(metrics.RequestDuration.WithLabelValues(method, path, status))
defer timer.ObserveDuration()
```

## Testing Standards

### Test Coverage Requirements

- **Minimum:** 90% coverage across all packages
- **Domain:** 95%+ (business logic is critical)
- **Use Case:** 100% (orchestration must be bulletproof)
- **HTTP Handlers:** 94%+ (all endpoints and error paths)

### Test Structure

Follow the **Arrange-Act-Assert** pattern:

```go
func TestWoLUseCase_SendWoLPacket(t *testing.T) {
    // Arrange
    mockRepo := &MockMachineRepository{}
    mockSender := &MockWoLPacketSender{}
    useCase := NewWoLUseCase(mockRepo, mockSender, logger, metrics)

    // Act
    err := useCase.SendWoLPacket(ctx, "saruman")

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 1, mockSender.SendCallCount())
}
```

### Mocking

- Use interface-based mocking
- Create mock implementations in test files
- Mock interfaces, not concrete types
- Keep mocks simple and focused

### Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
func TestMachine_Validate(t *testing.T) {
    tests := []struct {
        name    string
        machine Machine
        wantErr bool
    }{
        {
            name: "valid machine",
            machine: Machine{ID: "test", MAC: "AA:BB:CC:DD:EE:FF", Broadcast: "192.168.1.255"},
            wantErr: false,
        },
        {
            name: "invalid MAC",
            machine: Machine{ID: "test", MAC: "invalid", Broadcast: "192.168.1.255"},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.machine.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Development Workflow

### Running Tests

```bash
# Run all tests with coverage
just test

# Run specific package tests
go test ./internal/domain
go test ./internal/usecase -v

# Run with race detector
go test -race ./...
```

### Code Quality

```bash
# Format code
just format

# Run linters
just lint

# Run all pre-commit checks
just pre-commit
```

### Building

```bash
# Build binary
just build

# Build Docker image
just docker-build latest

# Run locally
export GWAIHIR_CONFIG=configs/machines.yaml
just run
```

## Common Tasks

### Adding a New Machine Attribute

1. Update `Machine` struct in `internal/domain/machine.go`
2. Add validation logic in `Validate()` method
3. Update `machines.yaml` schema and example
4. Add tests in `machine_test.go`
5. Update repository to parse new field
6. Ensure 95%+ test coverage

### Adding a New HTTP Endpoint

1. Define handler method in `internal/delivery/http/handler.go`
2. Register route in `internal/delivery/http/router.go`
3. Add authentication middleware if needed
4. Create tests in `handler_test.go`
5. Document endpoint in README.md
6. Add Prometheus metrics if applicable
7. Ensure 94%+ test coverage

### Adding a New Use Case Method

1. Define method in `WoLUseCase` interface in `internal/domain/usecase.go`
2. Implement in `internal/usecase/wol_usecase.go`
3. Add structured logging
4. Record metrics
5. Handle errors appropriately
6. Create comprehensive tests
7. Ensure 100% test coverage

### Adding a New Domain Error

1. Define error in `internal/domain/errors.go`
2. Use in appropriate layers
3. Map to HTTP status code in delivery layer
4. Add tests for error handling
5. Document in code comments

## Important Files

### Entry Point
- `cmd/gwaihir/main.go` - Application bootstrap, dependency injection
- `cmd/gwaihir/version.go` - Version information

### Configuration
- `configs/machines.yaml` - Machine allowlist configuration
- `configs/machines.example.yaml` - Template for configuration

### Core Domain
- `internal/domain/machine.go` - Machine entity with validation
- `internal/domain/repository.go` - Repository interfaces
- `internal/domain/errors.go` - Domain-specific errors

### Business Logic
- `internal/usecase/wol_usecase.go` - WoL operations orchestration

### Infrastructure
- `internal/infrastructure/wol_packet.go` - WoL packet sending
- `internal/infrastructure/logger.go` - Logging utilities
- `internal/infrastructure/metrics.go` - Prometheus metrics

### HTTP Layer
- `internal/delivery/http/handler.go` - Request handlers
- `internal/delivery/http/router.go` - Route setup
- `internal/delivery/http/auth.go` - Authentication
- `internal/delivery/http/middleware.go` - Request middleware
- `internal/delivery/http/health.go` - Health checks

### Data Access
- `internal/repository/yaml_machine_repository.go` - YAML-based storage

## Configuration

Gwaihir uses a unified YAML configuration file (gwaihir.yaml) that includes all settings: server, authentication, machines, and observability.

### Configuration File Format

```yaml
server:
  port: 8080
  log:
    format: json          # json or text
    level: info           # debug, info, warn, error

authentication:
  api_key: "secret-key"  # Optional: leave empty for public endpoints

machines:
  - id: machine-id            # Required: unique identifier
    name: "Display Name"      # Required: human-readable name
    mac: "AA:BB:CC:DD:EE:FF"  # Required: MAC address (XX:XX:XX:XX:XX:XX format)
    broadcast: "192.168.1.255" # Required: broadcast IP for network

observability:
  health_check:
    enabled: true             # Enable /health, /live, /ready endpoints
  metrics:
    enabled: true             # Enable /metrics endpoint
```

### Environment Variable Overrides

Environment variables override configuration file values:

```bash
GWAIHIR_CONFIG=/etc/gwaihir/gwaihir.yaml   # Config file path (default: /etc/gwaihir/gwaihir.yaml)
GWAIHIR_PORT=8080                          # Overrides server.port
GWAIHIR_LOG_FORMAT=json                    # Overrides server.log.format (json|text)
GWAIHIR_LOG_LEVEL=info                     # Overrides server.log.level (debug|info|warn|error)
GWAIHIR_API_KEY=secret-key                 # Overrides authentication.api_key
```

## Security Considerations

### API Key Authentication

- Set `GWAIHIR_API_KEY` to enable authentication
- All WoL/machine endpoints require `X-API-Key` header
- Health/metrics endpoints are always public
- Never log API keys

### Allowlist-Based Access

- Only machines in `machines.yaml` can receive WoL packets
- No dynamic machine registration
- Validate all inputs (MAC, broadcast IP)
- Use domain validation methods

### Network Security

- Runs with `hostNetwork: true` in Kubernetes (required for broadcast)
- Should be protected by NetworkPolicy
- Only accessible from trusted services (e.g., Smaug)

## Observability

### Structured Logging

All logs include:
- `request_id`: Correlation ID for request tracing
- `machine_id`: Target machine identifier
- `operation`: Operation being performed
- Appropriate log level (INFO, WARN, ERROR)

### Prometheus Metrics

Key metrics:
- `gwaihir_wol_packets_sent_total` - Successful WoL packets
- `gwaihir_wol_packets_failed_total` - Failed WoL packets
- `gwaihir_machine_not_found_total` - Machine not found errors
- `gwaihir_request_duration_seconds` - Request latency histogram
- `gwaihir_configured_machines_total` - Number of configured machines

## Troubleshooting

### Tests Failing

1. Check test coverage: `just test`
2. Run specific failing test: `go test ./path -run TestName -v`
3. Check for race conditions: `go test -race ./...`
4. Verify mocks are properly configured

### Linter Errors

1. Run `just format` to auto-fix formatting
2. Run `just lint` to see specific issues
3. Check `golangci-lint` configuration in `.golangci.yml`
4. Fix issues one by one, maintain code quality

### Build Issues

1. Verify Go version: `go version` (need 1.23+)
2. Clean build cache: `go clean -cache`
3. Update dependencies: `go mod tidy`
4. Check for missing dependencies: `go mod verify`

## Best Practices

### DO:
- Follow Clean Architecture layers strictly
- Write tests BEFORE implementation (TDD)
- Use structured logging with context
- Record metrics for important operations
- Validate inputs at domain layer
- Use domain-specific errors
- Keep functions small and focused
- Document exported functions and types
- Use Atomic Commits

### DON'T:
- Don't violate layer dependencies
- Don't skip writing tests
- Don't use unstructured logging
- Don't hardcode configuration values
- Don't ignore errors
- Don't use global state
- Don't mix concerns across layers
- Don't commit without running `just pre-commit`

## Release Process

Gwaihir uses SNAPSHOT-based versioning:

1. Development on SNAPSHOT versions (e.g., `0.2.0-SNAPSHOT`)
2. Run `just pre-release` to prepare release
3. Push to trigger CI/CD
4. CD workflow creates release and bumps to next SNAPSHOT

## Related Documentation

- [README.md](README.md) - User-facing documentation
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [2026-02-08-service-architecture.md](docs/adrs/2026-02-08-service-architecture.md) - Architecture details

## Quick Reference

```bash
# Install dependencies
just ci

# Run all checks
just pre-commit

# Run tests
just test

# Run locally
export GWAIHIR_CONFIG=configs/machines.yaml
just run

# Build
just build

# Docker
just docker-build latest
```

## Getting Help

- Review existing tests for patterns
- Check README.md for API documentation
- Review CONTRIBUTING.md for code standards
- Examine similar existing code in the layer you're working in
