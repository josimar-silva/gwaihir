# Gwaihir Service Architecture

**Document Version**: 1.0
**Date**: 2026-02-09
**Status**: Current

## Executive Summary

Gwaihir is a production-ready microservice for sending Wake-on-LAN (WoL) packets to sleeping machines in a homelab environment. 
Built with Go and following Clean Architecture principles, it provides a secure, observable, and maintainable solution for managing machine wake-up operations.

**Key Characteristics:**
- **Architecture**: Clean Architecture (Hexagonal/Ports & Adapters)
- **Language**: Go 1.22+
- **Framework**: Gin (HTTP routing)
- **Security**: API key authentication, allowlist-based access control
- **Observability**: Structured logging (slog), Prometheus metrics
- **Deployment**: Docker container, Kubernetes-ready

## Architecture Overview

### Clean Architecture Layers

Gwaihir implements Clean Architecture with dependency inversion, ensuring:
- **Independence from frameworks**: Core business logic has no framework dependencies
- **Testability**: All layers can be tested in isolation
- **Independence from UI/DB**: Core logic is independent of delivery mechanism
- **Independence from external agencies**: Business rules don't depend on infrastructure

```text
┌──────────────────────────────────────────────────────────────────┐
│                         Gwaihir Service                          │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │                      Domain Layer (Core)                   │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────────┐     │  │
│  │  │   Machine    │  │ Repository   │  │WoLPacketSender│     │  │
│  │  │   Entity     │  │  Interface   │  │  Interface    │     │  │
│  │  └──────────────┘  └──────────────┘  └───────────────┘     │  │
│  │  Business Rules • Entities • Interfaces                    │  │
│  └────────────────────────────────────────────────────────────┘  │
│                             ▲                                    │
│                             │                                    │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │                     Use Case Layer                         │  │
│  │  ┌──────────────────────────────────────────────────────┐  │  │
│  │  │            WoLUseCase (Business Logic)               │  │  │
│  │  │  • SendWakePacket(machineID)                         │  │  │
│  │  │  • ListMachines()                                    │  │  │
│  │  │  • GetMachine(machineID)                             │  │  │
│  │  └──────────────────────────────────────────────────────┘  │  │
│  │  Application-specific business rules                       │  │
│  └────────────────────────────────────────────────────────────┘  │
│                             ▲                                    │
│                             │                                    │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │                   Adapter Layers                           │  │
│  │                                                            │  │
│  │  Delivery (HTTP)        Repository           Infrastructure│  │
│  │  ┌─────────────┐  ┌───────────────┐  ┌─────────────────┐   │  │
│  │  │  Handlers   │  │YAMLMachineRepo│  │ WoLPacketSender │   │  │
│  │  │  Router     │  │               │  │ Logger          │   │  │
│  │  │  Middleware │  │               │  │ Metrics         │   │  │
│  │  └─────────────┘  └───────────────┘  └─────────────────┘   │  │
│  │  Adapts external interfaces to use case layer              │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. Domain Layer (`internal/domain/`)

**Purpose**: Contains core business entities and rules.

**Components**:
- **Entities**: `Machine` - Represents a wakeable machine
- **Interfaces**: `MachineRepository`, `WoLPacketSender`
- **Value Objects**: Domain errors
- **Business Rules**: MAC validation, broadcast IP validation

**Key Characteristics**:
- No external dependencies
- Pure Go code
- Contains interfaces that outer layers implement
- Highest level of abstraction

**Example**:
```go
// Machine entity with validation
type Machine struct {
    ID        string
    Name      string
    MAC       string
    Broadcast string
}

func (m *Machine) Validate() error {
    // Business rule: MAC must be valid format
    if err := ValidateMAC(m.MAC); err != nil {
        return err
    }
    // Business rule: Broadcast must be valid IPv4
    if err := ValidateBroadcast(m.Broadcast); err != nil {
        return err
    }
    return nil
}
```

### 2. Use Case Layer (`internal/usecase/`)

**Purpose**: Orchestrates business logic by coordinating domain entities.

**Components**:
- **WoLUseCase**: Main business logic for WoL operations

**Responsibilities**:
- Retrieve machines from repository
- Send WoL packets via packet sender interface
- Log operations
- Record metrics
- Handle errors and wrap with context

**Key Characteristics**:
- Depends only on domain interfaces
- Framework-agnostic
- 100% test coverage

**Example**:
```go
func (uc *WoLUseCase) SendWakePacket(machineID string) error {
    // Get machine from repository (domain interface)
    machine, err := uc.machineRepo.GetByID(machineID)
    if err != nil {
        uc.metrics.MachineNotFound.Inc()
        return fmt.Errorf("failed to get machine: %w", err)
    }

    // Log structured event
    uc.logger.Info("Sending WoL packet",
        infrastructure.String("machine_id", machine.ID))

    // Send packet via infrastructure interface
    if err := uc.packetSender.SendMagicPacket(machine.MAC, machine.Broadcast); err != nil {
        uc.metrics.WoLPacketsFailed.Inc()
        return fmt.Errorf("failed to send WoL packet: %w", err)
    }

    uc.metrics.WoLPacketsSent.Inc()
    return nil
}
```

### 3. Delivery Layer (`internal/delivery/http/`)

**Purpose**: Adapts HTTP requests to use case calls.

**Components**:
- **Handler**: HTTP request handlers
- **Router**: Route configuration
- **Middleware**: Authentication, logging, request ID
- **Health**: Health check endpoints

**Responsibilities**:
- Parse HTTP requests
- Validate input
- Call use case methods
- Format HTTP responses
- Handle authentication

**Key Characteristics**:
- Translates between HTTP and use case layer
- No business logic
- Returns appropriate HTTP status codes

**Example**:
```go
func (h *Handler) Wake(c *gin.Context) {
    var req WakeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Error: "Invalid request: " + err.Error(),
        })
        return
    }

    // Call use case (business logic)
    if err := h.wolUseCase.SendWakePacket(req.MachineID); err != nil {
        if errors.Is(err, domain.ErrMachineNotFound) {
            c.JSON(http.StatusNotFound, ErrorResponse{
                Error: "Machine not found or not allowed",
            })
            return
        }
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error: "Failed to send WoL packet",
        })
        return
    }

    c.JSON(http.StatusAccepted, SuccessResponse{
        Message: "WoL packet sent successfully",
    })
}
```

### 4. Repository Layer (`internal/repository/`)

**Purpose**: Implements data access.

**Components**:
- **YAMLMachineRepository**: Loads machines from YAML file

**Responsibilities**:
- Load configuration from YAML
- Cache machines in memory
- Validate machine data on load
- Detect duplicate IDs

**Key Characteristics**:
- Implements `domain.MachineRepository` interface
- Thread-safe (uses RWMutex)
- Immutable after initialization

**Example**:
```go
type YAMLMachineRepository struct {
    machines map[string]*domain.Machine
    mu       sync.RWMutex
}

func (r *YAMLMachineRepository) GetByID(id string) (*domain.Machine, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    machine, exists := r.machines[id]
    if !exists {
        return nil, domain.ErrMachineNotFound
    }
    return machine, nil
}
```

### 5. Infrastructure Layer (`internal/infrastructure/`)

**Purpose**: Implements infrastructure concerns.

**Components**:
- **WoLPacketSender**: Sends UDP broadcast packets
- **Logger**: Structured logging wrapper (slog)
- **Metrics**: Prometheus metrics

**Responsibilities**:
- Send actual WoL magic packets
- Emit structured logs
- Record Prometheus metrics
- Handle network errors

**Key Characteristics**:
- Implements domain interfaces
- Contains external dependencies (net, slog, prometheus)
- Handles low-level details

**Example**:
```go
func (s *WoLPacketSender) SendMagicPacket(mac, broadcast string) error {
    // Create magic packet (6 bytes of 0xFF + 16 repetitions of MAC)
    packet := make([]byte, 102)
    for i := 0; i < 6; i++ {
        packet[i] = 0xFF
    }

    macBytes := parseMACAddress(mac)
    for i := 1; i <= 16; i++ {
        copy(packet[i*6:(i+1)*6], macBytes)
    }

    // Send via UDP broadcast
    conn, err := s.dialer.Dial("udp", broadcast+":9")
    if err != nil {
        return fmt.Errorf("failed to dial broadcast address: %w", err)
    }
    defer conn.Close()

    if _, err := conn.Write(packet); err != nil {
        return fmt.Errorf("failed to send packet: %w", err)
    }

    return nil
}
```

## Dependency Flow

```text
main.go (wiring)
    │
    ├──> Repository (YAML)          ──┐
    ├──> Infrastructure (WoL, Logs) ──┼──> Use Case ──> Handler ──> Router
    └──> Metrics                    ──┘
```

**Dependency Rule**: Dependencies point inward. Inner layers don't know about outer layers.

- **main.go**: Wires everything together (dependency injection)
- **Handler**: Depends on `UseCase` interface
- **UseCase**: Depends on `MachineRepository` and `WoLPacketSender` interfaces
- **Repository/Infrastructure**: Implement domain interfaces

## Data Flow

### Sending a WoL Packet

```text
1. HTTP Request
   POST /wol
   {
     "machine_id": "saruman"
   }

2. Middleware Layer
   ├─> Request ID generated (UUID)
   ├─> API key validated (if configured)
   └─> Request logged with correlation ID

3. Handler Layer
   ├─> Parse JSON request
   ├─> Validate request structure
   └─> Call use case: SendWakePacket("saruman")

4. Use Case Layer
   ├─> Get machine from repository
   │   └─> YAMLMachineRepository.GetByID("saruman")
   ├─> Log: "Sending WoL packet" (with request_id, machine_id)
   ├─> Send packet via infrastructure
   │   └─> WoLPacketSender.SendMagicPacket("AA:BB:CC:DD:EE:FF", "192.168.1.255")
   ├─> Record metric: gwaihir_wol_packets_sent_total++
   └─> Return success

5. Infrastructure Layer
   ├─> Create 102-byte magic packet
   ├─> Open UDP connection to broadcast:9
   ├─> Send packet
   └─> Close connection

6. Handler Response
   └─> Return 202 Accepted
       {
         "message": "WoL packet sent successfully"
       }

7. Middleware Logging
   └─> Log request completion (status, duration, request_id)
```

## Security Architecture

### Authentication

**API Key Authentication**:
- Optional (controlled by `GWAIHIR_API_KEY` environment variable)
- Header-based: `X-API-Key: <key>`
- Middleware validates before reaching handlers
- Public endpoints: `/health`, `/version`, `/metrics`, `/live`, `/ready`
- Protected endpoints: `/wol`, `/machines`, `/machines/:id`

**Flow**:
```text
Request → APIKeyAuthMiddleware → Handler
          │
          ├─ No API key configured → Allow
          ├─ API key present & valid → Allow
          └─ API key missing/invalid → 401 Unauthorized
```

### Allowlist Security

**Machine Allowlist**:
- Only machines in `machines.yaml` can receive WoL packets
- Validated on startup (MAC format, broadcast IP format)
- Immutable at runtime (no dynamic registration)
- Duplicate IDs rejected

**Validation**:
```text
Startup → Load YAML → Validate Each Machine → Cache in Memory
          │            │
          │            ├─ Valid MAC? (XX:XX:XX:XX:XX:XX)
          │            ├─ Valid IPv4? (192.168.1.255)
          │            ├─ Unique ID?
          │            └─ All fields present?
          │
          Fail → Exit with error
          Success → Start HTTP server
```

## Observability Architecture

### Structured Logging

**Implementation**: Go's `log/slog` package

**Log Format** (JSON in production):
```json
{
  "time": "2026-02-09T15:30:45.123Z",
  "level": "INFO",
  "msg": "Sending WoL packet",
  "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "machine_id": "saruman",
  "mac": "AA:BB:CC:DD:EE:FF"
}
```

**Log Levels**:
- `INFO`: Normal operations (WoL sent, machines listed)
- `WARN`: Recoverable issues (invalid request format)
- `ERROR`: Failures (WoL send failed, machine not found)

**Request Correlation**:
- Every request gets a unique UUID (`request_id`)
- Propagated through all log statements
- Enables tracing across services

### Prometheus Metrics

**Metrics Exposed**:

```promql
# Counters
gwaihir_wol_packets_sent_total
gwaihir_wol_packets_failed_total
gwaihir_machine_not_found_total
gwaihir_machines_listed_total
gwaihir_machines_retrieved_total

# Histograms
gwaihir_request_duration_seconds{method, path, status}

# Gauges
gwaihir_configured_machines_total
```

**Instrumentation Points**:
- Use case layer: Records success/failure metrics
- Middleware: Records request duration
- Repository: Sets gauge on machine count

### Health Checks

**Three Endpoints**:

1. `/health` - Combined liveness + readiness
   - Returns 200 if healthy, 503 if unhealthy
   - Checks: Config loaded, machine count > 0

2. `/live` - Liveness probe
   - Returns 200 if process is alive
   - Used by Kubernetes liveness probe

3. `/ready` - Readiness probe
   - Returns 200 if ready to accept traffic
   - Checks configuration validity
   - Used by Kubernetes readiness probe

## Testing Strategy

### Test Coverage Targets

- **Overall**: 80%+ (Current: 90%+)
- **Domain**: 90%+ (Current: 95%)
- **Use Case**: 100% (Current: 100%)
- **Repository**: 85%+ (Current: 90.6%)
- **Infrastructure**: 85%+ (Current: 93.1%)
- **HTTP Delivery**: 85%+ (Current: 94%)

### Test Pyramid

```text
              ┌──────┐
              │ E2E  │  (Manual: Deploy and test in K8s)
             ┌┴──────┴┐
             │Integration│ (Minimal: Config → Use Case → Repository)
            ┌┴──────────┴┐
            │   Unit      │ (Extensive: All layers, all paths)
           ┌┴─────────────┴┐
           └───────────────┘
```

### Testing Approach

**Unit Tests**:
- Test each layer in isolation
- Use mocks for dependencies
- Table-driven tests for multiple scenarios
- Test both happy path and error paths

**Mock Strategy**:
- Domain interfaces are mocked in outer layer tests
- Simple mock implementations in test files
- Configurable behavior (errors, return values)
- Track invocations for verification

**Example Test**:
```go
func TestSendWakePacket_Success(t *testing.T) {
    // Arrange: Set up mocks
    repo := newMockMachineRepository(machines)
    sender := newMockWoLPacketSender()
    logger := newTestLogger()
    metrics := newTestMetrics()
    useCase := NewWoLUseCase(repo, sender, logger, metrics)

    // Act: Execute operation
    err := useCase.SendWakePacket("saruman")

    // Assert: Verify outcome
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if sender.callCount != 1 {
        t.Errorf("Expected 1 packet sent, got %d", sender.callCount)
    }
}
```

## Deployment Architecture

### Container Image

**Multi-stage Build**:
```dockerfile
# Stage 1: Build
FROM golang:1.22.2-alpine AS builder
COPY . .
RUN go build -o /gwaihir ./cmd/gwaihir

# Stage 2: Runtime (minimal)
FROM scratch
COPY --from=builder /gwaihir /gwaihir
ENTRYPOINT ["/gwaihir"]
```

**Image Size**: ~10-15 MB (scratch base image)

### Kubernetes Deployment

**Key Characteristics**:
- `hostNetwork: true` - Required for UDP broadcast packets
- `dnsPolicy: ClusterFirstWithHostNet` - Enables service discovery
- Liveness/Readiness probes configured
- Secrets for API key
- ConfigMap for machines.yaml
- NetworkPolicy to restrict access

**Resource Requirements**:
- CPU: 50m (request), 200m (limit)
- Memory: 64Mi (request), 128Mi (limit)

## Performance Characteristics

### Latency

**Expected Response Times**:
- `/health`, `/version`: <10ms
- `/wol`: <50ms (WoL packet send)
- `/machines`: <10ms (in-memory lookup)
- `/metrics`: <20ms

### Throughput

**Capacity**:
- ~1000 requests/second (single instance)
- Bottleneck: UDP packet sending (50ms per packet)
- Can be scaled horizontally if needed

### Resource Usage

**Memory**: ~20-30 MB (idle), ~50 MB (under load)
**CPU**: <1% (idle), ~5-10% (under load)
**Network**: Minimal (102-byte packets)

## Design Decisions

### Why Clean Architecture?

**Rationale**:
- **Testability**: 90%+ coverage achieved through isolation
- **Maintainability**: Clear separation of concerns
- **Flexibility**: Easy to swap implementations (YAML → DB)
- **Independence**: Core logic has no framework dependencies

### Why Go?

**Rationale**:
- **Networking**: Native UDP support, excellent concurrency
- **Performance**: Fast startup, low memory footprint
- **Ecosystem**: Prometheus, Kubernetes native integration
- **Simplicity**: Clean syntax, easy to maintain

### Why Gin?

**Rationale**:
- **Performance**: Fastest Go HTTP router (benchmarks)
- **Middleware**: Built-in support for authentication, logging
- **Community**: Large ecosystem, well-maintained

### Why YAML for Configuration?

**Rationale**:
- **Human-readable**: Easy to edit and review
- **Kubernetes-native**: ConfigMaps use YAML
- **Validation**: Parsed and validated on startup
- **Simplicity**: No database required for small machine lists

## Future Considerations

### Potential Enhancements

1. **Dynamic Configuration Reload**
   - Watch machines.yaml for changes
   - Hot-reload without restart
   - Preserves metrics and uptime

2. **Database Backend**
   - PostgreSQL/MySQL repository implementation
   - For larger machine lists (100+)
   - Implements same `MachineRepository` interface

3. **Multi-tenancy**
   - Support multiple API keys
   - Machine isolation per tenant
   - Per-tenant metrics

4. **Advanced Observability**
   - OpenTelemetry tracing
   - Distributed tracing across Smaug → Gwaihir
   - Jaeger/Zipkin integration

5. **Rate Limiting**
   - Prevent abuse (too many WoL packets)
   - Per-machine rate limits
   - Per-client rate limits

### Non-Goals

- **Machine Discovery**: Gwaihir doesn't discover machines automatically
- **WoL Status Verification**: Doesn't verify if machine actually woke up
- **Multi-protocol Support**: Only supports standard WoL (UDP port 9)
- **Stateful Operations**: Gwaihir is stateless (no persistent state)

## References

### Clean Architecture
- Robert C. Martin, "Clean Architecture: A Craftsman's Guide to Software Structure and Design"
- Hexagonal Architecture (Ports & Adapters pattern)

### Go Best Practices
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Related Documentation
- [README.md](../../README.md) - User-facing documentation
- [CONTRIBUTING.md](../../CONTRIBUTING.md) - Development guidelines
---

**Document History**:
- **2026-02-09**: Initial version (v1.0) - Clean Architecture implementation documented
