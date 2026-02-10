<p align="center"><img src="docs/images/gwaihir-logo.png" height="300px" weight="300px" alt="Gwaihir logo"></p>

<h1 align="center">Gwaihir</h1>
<div align="center">
   <!-- MIT License -->
  <a href="./LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="mit license" />
  </a> 
  <!-- Go version -->
  <a href="https://go.dev/doc/devel/release">
    <img src="https://img.shields.io/badge/go-1.23.0+-blue" alt="go version" />
  </a>
  <!-- Version -->
  <a href="./">
    <img src="https://img.shields.io/badge/version-0.1.0-orange.svg" alt="gwaihir" />
  </a>
  <!-- Go Report Card -->
  <a href="https://goreportcard.com/report/github.com/josimar-silva/gwaihir">
    <img src="https://goreportcard.com/badge/github.com/josimar-silva/gwaihir" alt="Gwaihir go report card" />
  </a>
  <!-- OSSF Score Card -->
  <a href="https://scorecard.dev/viewer/?uri=github.com/josimar-silva/gwaihir">
    <img src="https://img.shields.io/ossf-scorecard/github.com/josimar-silva/gwaihir?label=openssf+scorecard" alt="OpenSSF Score Card">
  </a>
  <!-- Coverage -->
  <a href="https://sonarcloud.io/summary/new_code?id=josimar-silva_gwaihir">
    <img src="https://sonarcloud.io/api/project_badges/measure?project=josimar-silva_gwaihir&metric=coverage" alt="coverage" />
  </a>
  <!-- Gwaihir Health -->
  <a href="https://hello.from-gondor.com/">
    <img src="https://status.from-gondor.com/api/v1/endpoints/internal_gwaihir/health/badge.svg" alt="Gwaihir Health" />
  </a>
  <!-- Gwaihir Uptime -->
  <a href="https://hello.from-gondor.com/">
    <img src="https://status.from-gondor.com/api/v1/endpoints/internal_gwaihir/uptimes/30d/badge.svg" alt="Gwaihir Uptime" />
  </a>
  <!-- Gwaihir Response Time -->
  <a href="https://hello.from-gondor.com/">
    <img src="https://status.from-gondor.com/api/v1/endpoints/internal_gwaihir/response-times/30d/badge.svg" alt="Gwaihir Response Time" />
  </a>
  <!-- CD -->
  <a href="https://github.com/josimar-silva/gwaihir/actions/workflows/cd.yaml">
    <img src="https://github.com/josimar-silva/gwaihir/actions/workflows/cd.yaml/badge.svg" alt="continuous delivery" />
  </a>
  <!-- CI -->
  <a href="https://github.com/josimar-silva/gwaihir/actions/workflows/ci.yaml">
    <img src="https://github.com/josimar-silva/gwaihir/actions/workflows/ci.yaml/badge.svg" alt="continuous integration" />
  </a>
  <!-- Docker -->
  <a href="https://github.com/josimar-silva/gwaihir/actions/workflows/docker.yaml">
    <img src="https://github.com/josimar-silva/gwaihir/actions/workflows/docker.yaml/badge.svg" alt="docker" />
  </a>
  <!-- CodeQL Advanced -->
  <a href="https://github.com/josimar-silva/gwaihir/actions/workflows/codeql.yaml">
    <img src="https://github.com/josimar-silva/gwaihir/actions/workflows/codeql.yaml/badge.svg" alt="CodeQL" />
  </a>
</div>

<div align="center">
  <b>G</b>o-based <b>W</b>ake-on-LAN <b>A</b>P<b>I</b> <b>H</b>andler for <b>I</b>nfrastructure <b>R</b>eliability
</div>

<div align="center">
  <sub>
    The Lord of the Eagles, a swift and noble messenger. When commanded by any trusted caller, Gwaihir takes flight across the network to deliver the wake-up call (the WoL packet) to the target machine.
  </sub>
</div>

## Table of Contents

- [Overview](#overview)
  - [Key Features](#key-features)
  - [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
  - [Machine Allowlist](#machine-allowlist)
  - [Environment Variables](#environment-variables)
- [API Endpoints](#api-endpoints)
  - [Authentication](#authentication)
  - [POST /wol](#post-wol)
  - [GET /machines](#get-machines)
  - [GET /machines/:id](#get-machinesid)
  - [GET /health](#get-health)
  - [GET /live](#get-live)
  - [GET /ready](#get-ready)
  - [GET /version](#get-version)
  - [GET /metrics](#get-metrics)
- [Observability](#observability)
  - [Structured Logging](#structured-logging)
  - [Prometheus Metrics](#prometheus-metrics)
- [Development](#development)
  - [Quick Start](#quick-start-1)
  - [Available Commands](#available-commands)
  - [Testing](#testing)
- [Deployment](#deployment)
  - [Docker](#docker)
  - [Kubernetes](#kubernetes)
- [Security](#security)
  - [Authentication](#authentication-1)
  - [Network Security](#network-security)
- [Monitoring & Alerting](#monitoring--alerting)
  - [Recommended Prometheus Alerts](#recommended-prometheus-alerts)
- [Release Process](#release-process)
- [Contributing](#contributing)
- [Troubleshooting](#troubleshooting)
  - [Common Issues](#common-issues)
- [Frequently Asked Questions](#frequently-asked-questions)
  - [General Questions](#general-questions)
  - [Security Questions](#security-questions)
  - [Operational Questions](#operational-questions)
- [License](#license)
- [Related Projects](#related-projects)

## Overview

Gwaihir is a production-ready microservice that provides Wake-on-LAN (WoL) capabilities via a secure REST API. It can be integrated with various systems, such as reverse proxies like [Smaug](https://github.com/josimar-silva/smaug), automation scripts, CI/CD pipelines, or any application that needs to wake up sleeping servers on demand. This isolates the privileged operation into a tiny, single-purpose, and easily audited service.

### Key Features

- **Allowlist-based Security**: Only machines explicitly configured in `machines.yaml` can receive WoL packets
- **API Key Authentication**: API key protection for all endpoints (except health/metrics)
- **Clean Architecture**: Separation of concerns with domain, use case, delivery, and repository layers
- **Gin Framework**: Fast HTTP router with excellent middleware support
- **Structured Logging**: JSON-formatted logs with request correlation IDs using Go's `log/slog`
- **Prometheus Metrics**: Comprehensive metrics for monitoring and alerting
- **Production-Grade Health Checks**: Separate liveness and readiness probes for Kubernetes
- **Type-Safe**: Strong validation for MAC addresses and broadcast IPs

### Architecture

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                         K8s Cluster / Network                           │
│                                                                         │
│  ┌─────────────┐                                                        │
│  │   Client    │───┐                                                    │
│  │ Application │   │                                                    │
│  └─────────────┘   │                                                    │
│                    │  API Calls                                         │
│  ┌─────────────┐   │  POST /wol                                         │
│  │   Reverse   │───┤  GET /machines          ┌───────────┐              │
│  │   Proxy     │   └────────────────────────▶│           │              │
│  │   (Smaug)   │                             │  Gwaihir  │              │
│  └─────────────┘   ┌────────────────────────▶│           │              │
│                    │  X-API-Key              │ (hostNet) │──────▶ WoL   │
│  ┌─────────────┐   │  Authentication         └───────────┘              │
│  │ Automation  │───┘                                │                   │
│  │   Scripts   │                                    │                   │
│  └─────────────┘                                    │                   │
│                                                     │                   │
│  ┌─────────────┐                                    │                   │
│  │ Prometheus  │◀───────────────────────────────────┘                   │
│  └─────────────┘      /metrics                                          │
└─────────────────────────────────────────────────────────────────────────┘
           │                                    │
           ▼                                    ▼
   ┌──────────────┐                      ┌──────────────┐
   │   Server 1   │                      │   Server 2   │
   │  (sleeping)  │                      │  (sleeping)  │
   └──────────────┘                      └──────────────┘
```

**Key Design Principles:**

- **Standalone Service**: Gwaihir is a self-contained microservice that any client can integrate with via its REST API
- **Security First**: Runs with `hostNetwork: true` (required for broadcast packets) but uses allowlist-based security and API key authentication to prevent unauthorized access
- **Separation of Concerns**: Isolates privileged network operations into a single, auditable service that can be tightly controlled via NetworkPolicy
- **Observable**: Exposes Prometheus metrics and structured logs for monitoring and debugging
- **Integration Flexibility**: Can be called by reverse proxies, automation scripts, orchestration systems, or any HTTP client

**Common Integration Patterns:**

1. **Reverse Proxy Integration**: Services like [Smaug](https://github.com/josimar-silva/smaug) can call Gwaihir to wake servers before forwarding requests
2. **CI/CD Pipelines**: Automation scripts can wake test/build servers on-demand
3. **Scheduled Tasks**: Cron jobs or Kubernetes CronJobs can wake servers at specific times
4. **Custom Applications**: Any application that needs WoL capabilities can integrate via the simple REST API

## Prerequisites

- Go 1.22.2 or later
- [golangci-lint](https://golangci-lint.run/usage/install/) for linting (development only)
- [just](https://github.com/casey/just) for task running (development only)
- Docker (optional, for containerized deployment)
- Kubernetes cluster (optional, for production deployment)

## Quick Start

Get Gwaihir running in under 5 minutes:

```bash
# 1. Clone the repository
git clone https://github.com/josimar-silva/gwaihir.git
cd gwaihir

# 2. Create a minimal configuration
cat > machines.yaml << EOF
machines:
  - id: test-server
    name: "Test Server"
    mac: "AA:BB:CC:DD:EE:FF"
    broadcast: "192.168.1.255"
EOF

# 3. Run the service
export GWAIHIR_CONFIG=machines.yaml
export LOG_JSON=false
go run cmd/gwaihir/main.go cmd/gwaihir/version.go

# 4. Test the service (in another terminal)
# Check health
curl http://localhost:8080/health

# Send WoL packet
curl -X POST http://localhost:8080/wol \
  -H "Content-Type: application/json" \
  -d '{"machine_id": "test-server"}'
```

**Note**: Replace the MAC address and broadcast IP with values for your network.

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

| Variable | Type | Description | Default |
|----------|------|-------------|---------|
| `GWAIHIR_CONFIG` | string | Path to machines.yaml config file | `/etc/gwaihir/machines.yaml` |
| `GWAIHIR_API_KEY` | string | API key for authentication | _(none)_ |
| `PORT` | string | HTTP server port | `8080` |
| `GIN_MODE` | string | Gin mode: `debug` or `release` | `release` |
| `LOG_JSON` | bool | Enable JSON logging (for production) | `true` |

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

- `400 Bad Request` - Invalid request body
```json
{
  "error": "invalid request body"
}
```

- `401 Unauthorized` - Missing or invalid API key
```json
{
  "error": "unauthorized"
}
```

- `404 Not Found` - Machine not in allowlist
```json
{
  "error": "machine not found"
}
```

- `500 Internal Server Error` - Failed to send WoL packet
```json
{
  "error": "failed to send WoL packet"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/wol \
  -H "X-API-Key: your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"machine_id": "saruman"}'
```

### GET /machines

List all machines in the allowlist.

**Authentication**: Required

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

**Authentication**: Required

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
  "version": "0.1.0",
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
  "version": "0.1.0",
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
export LOG_JSON=false                  # Human-readable logs for dev
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

# View coverage report in browser
# macOS:
open coverage.html
# Linux:
xdg-open coverage.html
# Windows:
start coverage.html

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

**Available Image Tags:**
- `latest` - Latest stable release (recommended for development)
- `vX.Y.Z` - Specific version tags (recommended for production)
- `main` - Built from main branch (bleeding edge, not recommended for production)

```bash
# Build Docker image
just docker-build latest

# Push to registry
just docker-push latest

# Run with Docker (using specific version)
docker run -d \
  -p 8080:8080 \
  -e GWAIHIR_API_KEY=your-secret-key \
  -v $(pwd)/configs:/etc/gwaihir \
  ghcr.io/josimar-silva/gwaihir:v0.1.0
```

Or use Docker directly:

```bash
docker build -t ghcr.io/josimar-silva/gwaihir:latest .
docker push ghcr.io/josimar-silva/gwaihir:latest
```

**Production Recommendation**: Always use specific version tags (e.g., `v0.1.0`) instead of `latest` to ensure reproducible deployments.

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
- Check broadcast address is correct for your network (typically x.x.x.255 for /24 networks)
- Ensure target machine's BIOS has WoL enabled (often called "Wake on LAN" or "Power On By PCI-E/PCI")
- Confirm MAC address is correct and formatted properly (colon or hyphen-separated)
- Check if firewall rules are blocking UDP port 9 (WoL uses UDP broadcast on port 9)
- Verify the target machine is on the same broadcast domain/VLAN as Gwaihir

**Authentication failures:**
- Verify `X-API-Key` header is included in request
- Check `GWAIHIR_API_KEY` environment variable is set correctly
- Ensure API key matches between client and server
- API keys are case-sensitive and must match exactly

**Health check failures:**
- Check that machines.yaml is mounted correctly at `/etc/gwaihir/machines.yaml`
- Verify at least one machine is configured
- Review logs for configuration parsing errors
- Ensure YAML syntax is valid (use `yamllint` to verify)

**Port conflicts:**
- Default port 8080 may be in use; set `PORT` environment variable to use different port
- Check if another service is bound to the same port: `netstat -tuln | grep 8080`

**Kubernetes-specific issues:**
- **RBAC permissions**: Ensure service account has necessary permissions (though Gwaihir requires no special permissions)
- **hostNetwork security**: Running with `hostNetwork: true` shares host network stack; ensure NetworkPolicy restricts access
- **Pod scheduling**: If using node selectors, ensure nodes are available and labeled correctly
- **ConfigMap not mounted**: Verify ConfigMap exists in same namespace and volume mount is correct

**Performance issues:**
- Check Prometheus metrics for high latency: `gwaihir_request_duration_seconds`
- Review logs for repeated errors or timeouts
- Ensure adequate CPU/memory limits (recommended: 100m CPU, 128Mi memory)

**YAML configuration errors:**
- Validate YAML syntax: `yamllint machines.yaml`
- Check for duplicate machine IDs
- Ensure all required fields are present (id, name, mac, broadcast)
- MAC addresses must be in format: `AA:BB:CC:DD:EE:FF` or `aa:bb:cc:dd:ee:ff`

## Frequently Asked Questions

### General Questions

**Q: Why is `hostNetwork: true` required?**
A: Wake-on-LAN packets must be sent as UDP broadcasts to the network broadcast address (e.g., 192.168.1.255). Container networking typically doesn't allow broadcast packets. Using `hostNetwork: true` gives Gwaihir direct access to the host's network interface, enabling proper broadcast transmission.

**Q: Can I run multiple replicas of Gwaihir?**
A: Yes, but it's typically unnecessary. Multiple replicas will each send duplicate WoL packets when requested. For high availability, consider using a Kubernetes Deployment with `replicas: 1` and proper liveness/readiness probes rather than horizontal scaling.

**Q: Is IPv6 supported?**
A: Currently, Gwaihir supports IPv4 broadcast addresses only. IPv6 uses multicast rather than broadcast, which would require different packet structure.

**Q: How do I handle machine IP address changes?**
A: Gwaihir only needs the MAC address and broadcast address, not the machine's IP. As long as the MAC address remains the same (it's tied to the network interface hardware), IP changes don't affect WoL functionality.

**Q: Can Gwaihir wake machines across different VLANs/subnets?**
A: No, broadcast packets are confined to their local broadcast domain. To wake machines on different subnets, you need either:
- Deploy separate Gwaihir instances in each subnet
- Configure directed broadcasts at your router (security risk, not recommended)
- Use unicast WoL if your network supports it (requires machines to have static IPs)

**Q: Does Gwaihir confirm that machines actually woke up?**
A: No, Wake-on-LAN is a fire-and-forget protocol. Gwaihir sends the magic packet but has no way to verify if the target machine powered on. You'll need to implement your own health checks or monitoring for the target machines.

**Q: What happens if I send a WoL packet to an already-running machine?**
A: Nothing harmful. The machine will simply ignore the WoL packet. It's safe to send WoL packets to machines regardless of their current power state.

### Security Questions

**Q: Is it safe to run with `hostNetwork: true`?**
A: Running with `hostNetwork: true` does increase the attack surface since the pod shares the host's network namespace. Mitigate risks by:
- Using NetworkPolicy to restrict which pods can access Gwaihir
- Enabling API key authentication (`GWAIHIR_API_KEY`)
- Running with minimal privileges (Gwaihir requires no special capabilities)
- Regularly monitoring access logs and metrics

**Q: Should I use API key authentication?**
A: Yes, for production deployments. Even within a cluster, defense in depth is important. If an attacker compromises a pod that can reach Gwaihir, the API key provides an additional security layer.

**Q: Can I use OAuth/JWT instead of API keys?**
A: Currently, Gwaihir supports simple API key authentication via `X-API-Key` header. For more sophisticated authentication, consider placing an authentication proxy (like OAuth2 Proxy) in front of Gwaihir.

### Operational Questions

**Q: How do I update the machine allowlist without restarting?**
A: Currently, you must restart Gwaihir after updating `machines.yaml`. The configuration is loaded at startup. 

**Q: What are the resource requirements?**
A: Gwaihir is extremely lightweight:
- **CPU**: ~10m idle, ~50m under load (recommended: 100m limit)
- **Memory**: ~20Mi idle, ~50Mi under load (recommended: 128Mi limit)
- **Network**: Negligible (WoL packets are 102 bytes each)

**Q: How many WoL packets per second can Gwaihir handle?**
A: Gwaihir can easily handle hundreds of WoL requests per second on modest hardware. The actual limit depends on network interface capabilities. 
For reference, sending 1000 WoL packets takes less than 100ms on typical hardware.

**Q: Can I use this with Docker Compose?**
A: Yes, but you need to use `network_mode: host` in your docker-compose.yml:
```yaml
services:
  gwaihir:
    image: ghcr.io/josimar-silva/gwaihir:v0.1.0
    network_mode: host
    environment:
      - GWAIHIR_API_KEY=your-secret-key
    volumes:
      - ./machines.yaml:/etc/gwaihir/machines.yaml
```

## License

Licensed under the MIT License. See [LICENSE](LICENSE) for details.

## Related Projects

- [Smaug](https://github.com/josimar-silva/smaug) - Config-driven reverse proxy with automatic Wake-on-LAN
