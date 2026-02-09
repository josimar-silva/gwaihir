# Gwaihir

**G**o-based **W**ake-on-LAN **A**PI **H**andler for **I**nfrastructure **R**eliability

The Lord of the Eagles, a swift and noble messenger. When commanded by Smaug, Gwaihir takes flight across the network to deliver the wake-up call (the WoL packet) to the target machine. This isolates the privileged operation into a tiny, single-purpose, and easily audited service.

## Overview

Gwaihir is a production-ready microservice responsible for sending Wake-on-LAN (WoL) packets. It's designed to work in conjunction with [Smaug](https://github.com/josimar-silva/smaug), a reverse proxy that commands Gwaihir to wake up sleeping servers on demand.

### Key Features

- **Allowlist-based Security**: Only machines explicitly configured in `machines.yaml` can receive WoL packets
- **API Key Authentication**: Optional API key protection for all endpoints (except health/metrics)
- **Clean Architecture**: Separation of concerns with domain, use case, delivery, and repository layers
- **Gin Framework**: Fast HTTP router with excellent middleware support
- **Structured Logging**: JSON-formatted logs with request correlation IDs using Go's `log/slog`
- **Prometheus Metrics**: Comprehensive metrics for monitoring and alerting
- **Production-Grade Health Checks**: Separate liveness and readiness probes for Kubernetes
- **Comprehensive Testing**: 90%+ test coverage across all layers with mocks
- **Type-Safe**: Strong validation for MAC addresses and broadcast IPs

### Architecture

```text
┌────────────────────────────────────────────────────────────────────────┐
│                             K8s Cluster                                │
│                                                                        │
│  ┌─────────────┐     ┌───────────┐      ┌───────────┐                  │
│  │  OpenWebUI  │────▶│           │ api  │           │                  │
│  └─────────────┘     │   Smaug   │─────▶│  Gwaihir  │                  │
│                      │ (unpriv.) │      │ (priv.)   │                  │
│  ┌─────────────┐     │           │      │           │                  │
│  │   Client    │────▶│           │      │           │───────▶ WoL      │
│  └─────────────┘     └───────────┘      └───────────┘                  │
│                                                │                       │
│  ┌─────────────┐                              │                        │
│  │ Prometheus  │◀─────────────────────────────┘                        │
│  └─────────────┘      /metrics                                         │
└────────────────────────────────────────────────────────────────────────┘
           │                                    │
           ▼                                    ▼
   ┌──────────────┐                      ┌──────────────┐
   │   Server 1   │                      │   Server 2   │
   │  (sleeping)  │                      │  (sleeping)  │
   └──────────────┘                      └──────────────┘
```

## Clean Architecture

The project follows Clean Architecture principles with clear separation of concerns:

```text
gwaihir/
├── cmd/gwaihir/              # Application entry point
│   ├── main.go               # Dependency injection & wiring
│   └── version.go            # Version information
├── internal/
│   ├── domain/               # Business entities & interfaces (core)
│   │   ├── machine.go        # Machine entity with validation
│   │   ├── repository.go     # Repository interface
│   │   └── errors.go         # Domain errors
│   ├── usecase/              # Business logic
│   │   └── wol_usecase.go    # WoL operations use case
│   ├── delivery/http/        # HTTP handlers (Gin)
│   │   ├── handler.go        # HTTP request handlers
│   │   ├── router.go         # Route configuration
│   │   ├── auth.go           # API key authentication
│   │   ├── middleware.go     # Request logging & correlation
│   │   └── health.go         # Health check handlers
│   ├── repository/           # Data access implementations
│   │   └── yaml_machine_repository.go
│   └── infrastructure/       # Infrastructure concerns
│       ├── logger.go         # Structured logging wrapper
│       ├── metrics.go        # Prometheus metrics
│       └── wol_packet.go     # WoL packet sender
└── configs/
    └── machines.yaml         # Allowlist configuration
```

## Prerequisites

- Go 1.22.2 or later
- [golangci-lint](https://golangci-lint.run/usage/install/) for linting
- [just](https://github.com/casey/just) for task running

## Configuration

### Machine Allowlist

Create a `machines.yaml` file with your allowed machines:

```yaml
machines:
  - id: saruman
    name: "Saruman - AI Inference Server"
    mac: "AA:BB:CC:DD:EE:FF"
    broadcast: "192.168.1.255"

  - id: morgoth
    name: "Morgoth - Transcription Server"
    mac: "11:22:33:44:55:66"
    broadcast: "192.168.1.255"
```

See `configs/machines.example.yaml` for a template.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GWAIHIR_CONFIG` | Path to machines.yaml config file | `/etc/gwaihir/machines.yaml` |
| `GWAIHIR_API_KEY` | API key for authentication (optional) | _(none)_ |
| `PORT` | HTTP server port | `8080` |
| `GIN_MODE` | Gin mode: `debug` or `release` | `release` |
| `LOG_JSON` | Enable JSON logging (for production) | `true` |

## API Endpoints

### Authentication

When `GWAIHIR_API_KEY` is set, all WoL and machine management endpoints require authentication via the `X-API-Key` header.

```bash
# With authentication
curl -X POST http://localhost:8080/wol \
  -H "X-API-Key: your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"machine_id": "saruman"}'

# Without authentication (health/metrics always accessible)
curl http://localhost:8080/health
```

### POST /wol

Send a Wake-on-LAN packet to a specified machine (must be in allowlist).

**Authentication**: Required (if API key is configured)

**Request Body:**
```json
{
  "machine_id": "saruman"
}
```

**Success Response:** `202 Accepted`
```json
{
  "message": "WoL packet sent successfully"
}
```

**Error Responses:**

- `401 Unauthorized` - Missing or invalid API key
- `404 Not Found` - Machine not in allowlist
- `500 Internal Server Error` - Failed to send WoL packet

**Example:**
```bash
curl -X POST http://localhost:8080/wol \
  -H "X-API-Key: your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"machine_id": "saruman"}'
```

### GET /machines

List all machines in the allowlist.

**Authentication**: Required (if API key is configured)

**Success Response:** `200 OK`
```json
[
  {
    "id": "saruman",
    "name": "Saruman - AI Inference Server",
    "mac": "AA:BB:CC:DD:EE:FF",
    "broadcast": "192.168.1.255"
  },
  {
    "id": "morgoth",
    "name": "Morgoth - Transcription Server",
    "mac": "11:22:33:44:55:66",
    "broadcast": "192.168.1.255"
  }
]
```

### GET /machines/:id

Get details of a specific machine.

**Authentication**: Required (if API key is configured)

**Success Response:** `200 OK`
```json
{
  "id": "saruman",
  "name": "Saruman - AI Inference Server",
  "mac": "AA:BB:CC:DD:EE:FF",
  "broadcast": "192.168.1.255"
}
```

**Error Responses:**
- `401 Unauthorized` - Missing or invalid API key
- `404 Not Found` - Machine not found

### GET /health

Combined health check endpoint (liveness + readiness).

**Authentication**: Not required

**Success Response:** `200 OK`
```json
{
  "status": "healthy",
  "version": "0.2.0",
  "machines_loaded": 2,
  "uptime_seconds": 3600,
  "checks": {
    "config_loaded": "ok",
    "machine_count": "ok"
  }
}
```

**Degraded Response:** `503 Service Unavailable`
```json
{
  "status": "unhealthy",
  "checks": {
    "config_loaded": "failed: no machines configured"
  }
}
```

### GET /live

Liveness probe (process is alive).

**Authentication**: Not required

**Response:** `200 OK` - Process is alive

### GET /ready

Readiness probe (ready to accept traffic).

**Authentication**: Not required

**Success Response:** `200 OK`
```json
{
  "status": "ready",
  "machines_loaded": 2
}
```

**Not Ready Response:** `503 Service Unavailable`

### GET /version

Version information endpoint.

**Authentication**: Not required

**Success Response:** `200 OK`
```json
{
  "version": "0.2.0",
  "build_time": "2026-02-09_14:30:00",
  "git_commit": "abc1234"
}
```

### GET /metrics

Prometheus metrics endpoint.

**Authentication**: Not required

**Response:** Prometheus text format

## Observability

### Structured Logging

All logs are emitted in JSON format (when `LOG_JSON=true`) with structured fields:

**Example Log Output:**
```json
{
  "time": "2026-02-09T15:30:45.123Z",
  "level": "INFO",
  "msg": "Sending WoL packet",
  "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "machine_id": "saruman",
  "machine_name": "Saruman - AI Inference Server",
  "mac": "AA:BB:CC:DD:EE:FF",
  "broadcast": "192.168.1.255"
}
```

**Key Log Fields:**
- `request_id`: Correlation ID for tracing requests across services
- `machine_id`: Target machine identifier
- `operation`: Operation being performed (e.g., "send_wol", "list_machines")
- `error`: Error details (when applicable)

**Log Levels:**
- `INFO`: Normal operations (WoL packet sent, machine listed)
- `WARN`: Recoverable issues (invalid request)
- `ERROR`: Failures (WoL send failed, machine not found)

### Prometheus Metrics

Gwaihir exposes comprehensive metrics for monitoring:

**Counter Metrics:**
```promql
# Total WoL packets successfully sent
gwaihir_wol_packets_sent_total

# Total WoL packet send failures
gwaihir_wol_packets_failed_total

# Total machine not found errors
gwaihir_machine_not_found_total

# Total machine list operations
gwaihir_machines_listed_total

# Total machine retrieve operations
gwaihir_machines_retrieved_total
```

**Histogram Metrics:**
```promql
# Request duration in seconds
gwaihir_request_duration_seconds_bucket{method="POST",path="/wol",status="202"}
gwaihir_request_duration_seconds_sum
gwaihir_request_duration_seconds_count
```

**Gauge Metrics:**
```promql
# Number of configured machines in allowlist
gwaihir_configured_machines_total
```

**Example Prometheus Queries:**

```promql
# WoL packet success rate over last 5m
rate(gwaihir_wol_packets_sent_total[5m])
  / (rate(gwaihir_wol_packets_sent_total[5m]) + rate(gwaihir_wol_packets_failed_total[5m]))

# 99th percentile request latency
histogram_quantile(0.99, rate(gwaihir_request_duration_seconds_bucket[5m]))

# Total machines not found errors in last hour
increase(gwaihir_machine_not_found_total[1h])
```

**Grafana Dashboard Example:**
```yaml
# Alert when WoL failure rate exceeds 10%
- alert: HighWoLFailureRate
  expr: |
    rate(gwaihir_wol_packets_failed_total[5m])
      / (rate(gwaihir_wol_packets_sent_total[5m]) + rate(gwaihir_wol_packets_failed_total[5m])) > 0.1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High WoL failure rate detected"
```

## Development

### Quick Start

```bash
# Install dependencies
just ci

# Run tests
just test

# Run linter
just lint

# Format code
just format

# Build binary
just build

# Run locally (requires config file)
export GWAIHIR_CONFIG=configs/machines.yaml
export GWAIHIR_API_KEY=dev-secret-key  # Optional for testing
export LOG_JSON=false                   # Human-readable logs for dev
just run
```

### Available Commands

Run `just` or `just --list` to see all available commands:

```bash
just              # Show all available commands
just ci           # Install dependencies
just check        # Run format check and lint
just format       # Format code
just lint         # Run linters
just test         # Run tests with coverage
just build        # Build binary
just run          # Run locally
just pre-commit   # Run all checks before committing
```

### Testing

The project has comprehensive test coverage (90%+) across all layers:

```bash
# Run all tests
go test ./...

# Run tests with coverage
just test

# View coverage report
just test && open coverage.html

# Test specific package
go test ./internal/domain
go test ./internal/usecase
go test ./internal/infrastructure
```

**Test Coverage by Package:**
- Domain: 95%
- Use Case: 100%
- Repository: 90.6%
- Infrastructure: 93.1%
- HTTP Delivery: 94%

See [CONTRIBUTING.md](CONTRIBUTING.md) for testing standards and best practices.

## Deployment

### Docker

```bash
# Build Docker image
just docker-build latest

# Push to registry
just docker-push latest

# Run with Docker
docker run -d \
  -p 8080:8080 \
  -e GWAIHIR_API_KEY=your-secret-key \
  -v $(pwd)/configs:/etc/gwaihir \
  ghcr.io/josimar-silva/gwaihir:latest
```

Or use Docker directly:

```bash
docker build -t ghcr.io/josimar-silva/gwaihir:latest .
docker push ghcr.io/josimar-silva/gwaihir:latest
```

### Kubernetes

Gwaihir runs with `hostNetwork: true` to enable broadcast packets. It should be protected by a NetworkPolicy allowing only trusted services to access it.

**Example Deployment:**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gwaihir-config
  namespace: homelab
data:
  machines.yaml: |
    machines:
      - id: saruman
        name: "Saruman - AI Inference Server"
        mac: "AA:BB:CC:DD:EE:FF"
        broadcast: "192.168.1.255"
---
apiVersion: v1
kind: Secret
metadata:
  name: gwaihir-api-key
  namespace: homelab
type: Opaque
stringData:
  api-key: "your-secret-key-here"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gwaihir
  namespace: homelab
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gwaihir
  template:
    metadata:
      labels:
        app: gwaihir
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      containers:
        - name: gwaihir
          image: ghcr.io/josimar-silva/gwaihir:latest
          env:
            - name: GWAIHIR_CONFIG
              value: /etc/gwaihir/machines.yaml
            - name: GWAIHIR_API_KEY
              valueFrom:
                secretKeyRef:
                  name: gwaihir-api-key
                  key: api-key
            - name: LOG_JSON
              value: "true"
            - name: GIN_MODE
              value: "release"
          ports:
            - containerPort: 8080
              protocol: TCP
          livenessProbe:
            httpGet:
              path: /live
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
          volumeMounts:
            - name: config
              mountPath: /etc/gwaihir
      volumes:
        - name: config
          configMap:
            name: gwaihir-config
---
apiVersion: v1
kind: Service
metadata:
  name: gwaihir
  namespace: homelab
spec:
  selector:
    app: gwaihir
  ports:
    - port: 8080
      targetPort: 8080
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: gwaihir-network-policy
  namespace: homelab
spec:
  podSelector:
    matchLabels:
      app: gwaihir
  policyTypes:
    - Ingress
  ingress:
    # Allow from Smaug proxy
    - from:
        - podSelector:
            matchLabels:
              app: smaug
      ports:
        - protocol: TCP
          port: 8080
    # Allow from Prometheus
    - from:
        - namespaceSelector:
            matchLabels:
              name: monitoring
      ports:
        - protocol: TCP
          port: 8080
```

## Security

### Authentication

- **API Key Protection**: All WoL and machine endpoints require valid API key (when configured)
- **Header-based Auth**: Uses `X-API-Key` header for authentication
- **Secure Storage**: API key should be stored in Kubernetes Secret or environment variable
- **No Key Logging**: API keys are never logged in plaintext

### Network Security

- **Allowlist-only**: Only machines explicitly configured in `machines.yaml` can receive WoL packets
- **Validation**: MAC addresses and broadcast IPs are validated on startup
- **No dynamic registration**: Machines cannot be added at runtime
- **NetworkPolicy**: Should be restricted to only allow access from trusted services
- **Timeouts**: HTTP server has proper read/write timeouts configured
- **Graceful shutdown**: Handles SIGTERM/SIGINT properly

### Best Practices

1. **Rotate API Keys Regularly**: Update `GWAIHIR_API_KEY` periodically
2. **Use Kubernetes Secrets**: Never hardcode API keys in manifests
3. **Enable NetworkPolicy**: Restrict access to only authorized pods
4. **Monitor Metrics**: Alert on unusual patterns (high failure rates, unauthorized access)
5. **Review Logs**: Periodically audit access logs for suspicious activity

## Monitoring & Alerting

### Recommended Prometheus Alerts

```yaml
groups:
  - name: gwaihir
    interval: 30s
    rules:
      - alert: GwaihirDown
        expr: up{job="gwaihir"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Gwaihir service is down"

      - alert: HighWoLFailureRate
        expr: |
          rate(gwaihir_wol_packets_failed_total[5m])
            / (rate(gwaihir_wol_packets_sent_total[5m]) + rate(gwaihir_wol_packets_failed_total[5m])) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High WoL packet failure rate (>10%)"

      - alert: NoMachinesConfigured
        expr: gwaihir_configured_machines_total == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "No machines configured in Gwaihir allowlist"

      - alert: HighRequestLatency
        expr: histogram_quantile(0.99, rate(gwaihir_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "99th percentile request latency >1s"
```

## Release Process

This project uses a SNAPSHOT-based versioning system similar to Maven:

1. Development happens on SNAPSHOT versions (e.g., `0.2.0-SNAPSHOT`)
2. To create a release, run `just pre-release`
3. This will:
   - Run all tests and checks
   - Remove the SNAPSHOT suffix
   - Commit the version bump
4. Push to trigger the CI/CD pipeline
5. The CD workflow will:
   - Create a GitHub release
   - Build and attach binaries
   - Automatically bump to the next SNAPSHOT version

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:
- Code style and standards
- Testing requirements (90%+ coverage)
- Pull request process
- Commit message conventions

## Troubleshooting

### Common Issues

**WoL packets not reaching machines:**
- Verify `hostNetwork: true` is set in Kubernetes deployment
- Check broadcast address is correct for your network
- Ensure target machine's BIOS has WoL enabled
- Confirm MAC address is correct and formatted properly

**Authentication failures:**
- Verify `X-API-Key` header is included in request
- Check `GWAIHIR_API_KEY` environment variable is set correctly
- Ensure API key matches between client and server

**Health check failures:**
- Check that machines.yaml is mounted correctly
- Verify at least one machine is configured
- Review logs for configuration errors

## License

Licensed under the MIT License. See [LICENSE](LICENSE) for details.

## Related Projects

- [Smaug](https://github.com/josimar-silva/smaug) - Config-driven reverse proxy with automatic Wake-on-LAN
- [Project Elrond](https://github.com/josimar-silva/elrond) - Homelab infrastructure orchestration

## Architecture Documentation

For detailed architectural decisions and design patterns, see:
- [Service Architecture](docs/service-architecture.md) - Clean Architecture implementation details
- [IMPROVEMENT_PLAN.md](IMPROVEMENT_PLAN.md) - Feature development roadmap and progress
