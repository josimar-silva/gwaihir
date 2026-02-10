# Unified Configuration Architecture

**Document Version**: 1.0
**Date**: 2026-02-10
**Status**: Proposed
**ADR Type**: Architecture Decision
**Related Issue**: Configuration improvements v0.1.1

## Executive Summary

This ADR proposes consolidating Gwaihir's scattered configuration into a single unified YAML configuration file. Currently, application settings are scattered across environment variables, with machines defined separately. This change establishes a single source of truth for all configuration, improves deployment simplicity, and establishes a foundation for future features like live configuration reloading.

**Key Changes**:
- Single configuration file (`gwaihir.yaml`) containing server, authentication, machines, and observability settings
- Centralized configuration package with explicit validation at startup
- Environment variables override configuration file values (clear precedence hierarchy)
- Conditional endpoint exposure based on configuration
- Log level controls all logging behavior (no redundant log flags)

## Problem Statement

### Current Configuration Challenges

**Scattered configuration sources:**
- Application settings spread across environment variables (`GWAIHIR_CONFIG`, `PORT`, `GWAIHIR_API_KEY`, `GWAIHIR_PRODUCTION`, `LOG_JSON`, `GIN_MODE`)
- Machines defined in separate `machines.yaml` file
- No unified view of all configuration
- Inconsistent defaults handling
- Unclear configuration precedence and defaults

**Semantic ambiguity:**
- `GWAIHIR_PRODUCTION` boolean flag has unclear meaning (controls both log format and level?)
- `LOG_JSON` flag is cryptic; what does "JSON" control exactly?
- No clear relationship between environment variables and their purpose
- Difficult to understand complete configuration state

**Observability configuration issues:**
- Health check and metrics endpoints always exposed (no control)
- All endpoints logged at same level regardless of log level setting
- Multiple redundant flags controlling logging behavior

**Deployment and operations friction:**
- Single application deployment requires understanding multiple environment variables
- Configuration changes scattered across multiple places
- Configuration validation errors occur at runtime, not startup
- No fail-fast mechanism for invalid configurations
- Difficult to document configuration in a single place

**Future feature limitations:**
- Separate configuration sources complicate live reload feature
- No clear extension point for additional observability settings
- Difficult to reason about configuration lifecycle and precedence

## Proposed Solution

### Core Principles

```
Single Configuration File
    ↓
Centralized Validation (Fail-Fast at Startup)
    ↓
Clear Precedence (CLI > Env > File > Defaults)
    ↓
Semantic Configuration Names
    ↓
Conditional Feature Exposure
```

### Configuration Structure

```
gwaihir.yaml
├── server
│   ├── port
│   └── log
│       ├── format (json|text)
│       └── level (debug|info|warn|error)
├── authentication
│   └── api_key
├── machines
│   ├── [0]
│   │   ├── id
│   │   ├── name
│   │   ├── mac
│   │   └── broadcast
│   └── [1] ...
└── observability
    ├── health_check
    │   └── enabled
    └── metrics
        └── enabled
```

### Configuration Precedence Hierarchy

```
┌─────────────────────────────────────────┐
│   Precedence: Higher → Lower            │
├─────────────────────────────────────────┤
│  1. Environment Variables               │  GWAIHIR_*
│     (GWAIHIR_API_KEY, GWAIHIR_LOG_LEVEL)  Highest Priority
├─────────────────────────────────────────┤
│  2. Configuration File                  │  gwaihir.yaml
│     (All top-level sections)            │
├─────────────────────────────────────────┤
│  3. Hardcoded Defaults                  │  Port: 8080
│     (When neither env nor file specify)  │  Format: text
│                                          │  Level: info
│                                          │  Lowest Priority
└─────────────────────────────────────────┘
```

### Logging Behavior Simplification

```
Current State (Redundant)          →    Proposed State (Semantic)
├── GWAIHIR_PRODUCTION                 ├── server.log.level: debug|info|warn|error
├── LOG_JSON (true|false)              └── server.log.format: json|text
├── LogRequests flag
└── LogLevel per endpoint              Single source of truth:
                                       - Level controls request logging threshold
                                       - Format controls output style
```

### Authentication Configuration

**Simple, non-redundant approach:**
- Single `authentication.api_key` field
- Can be set in config file OR via `GWAIHIR_API_KEY` env var (env var takes precedence)
- No feature flags needed (Bearer token deferred to v0.2)
- Middleware checks X-API-Key header when key is configured

### Observability Control

```
observability:
  health_check:
    enabled: true    ← Controls /health, /live, /ready endpoint exposure
  metrics:
    enabled: true    ← Controls /metrics endpoint exposure

Logic:
- enabled: false → Endpoint not registered, not accessible, not logged
- enabled: true → Endpoint registered, uses server.log.level for logging
```

## Architectural Benefits

### 1. Single Source of Truth
- All configuration in one file eliminates synchronization issues
- Easier to review: "What's my Gwaihir configuration?" → Read one file
- Simpler to validate: One validation pass catches all issues

### 2. Fail-Fast Validation
```
Startup Sequence:
  1. Load configuration file
  2. Apply environment variable overrides
  3. Validate all fields (required, format, valid values)
  4. Validate all machines (MAC, broadcast IP)
  5. If any validation fails → Exit with clear error message
  
Current Behavior: Runtime errors discovered after startup
New Behavior: Configuration errors caught immediately
```

### 3. Clear Precedence
```
Before: Unclear which env vars override which file values
After:  Explicit hierarchy makes reasoning easier

Example:
- File has: authentication.api_key: "file-key"
- Env has:  GWAIHIR_API_KEY=env-key
- Result:   Env var wins (clear and documented)
```

### 4. Semantic Configuration Names
```
Before                          After
├── inProduction: bool    →    ├── server.log.level: string
├── LOG_JSON: bool             └── server.log.format: string
└── GIN_MODE: string
                              Intent is clear:
                              - Level controls verbosity
                              - Format controls output style
```

### 5. Reduced Deployment Complexity
```
Before:
  Volumes:
    - machines.yaml
    - application config (env vars)
    
After:
  Volumes:
    - gwaihir.yaml
    
Kubernetes:
  Before: Mount 2 ConfigMaps
  After:  Mount 1 ConfigMap
```

### 6. Future-Ready for Live Reload
```
Live Reload Implementation Simplified:
- Single file to watch with fsnotify
- Single validation pass to validate reload
- No race conditions between multiple config files
- Atomic reload: All settings change together or none change
```

## Design Decisions & Trade-offs

### Decision 1: Single File vs. Feature Flags for Auth Methods

**Considered Options**:
1. Single `api_key` field (recommended)
2. Feature flags for each auth method (`api_key.enabled`, `bearer.enabled`)

**Chosen**: Option 1 (Single field)

**Rationale**:
- Currently only support X-API-Key
- Bearer token is v0.2 feature
- Simpler config without unnecessary abstraction
- When Bearer is added (v0.2): Simply add `bearer` field alongside `api_key`
- Middleware can check both sequentially without config complexity
- YAGNI principle: Don't add flexibility until needed

**Trade-off**: If future demands multi-method support simultaneously, we refactor to feature flags (backward-compatible change)

### Decision 2: Conditional Endpoint Exposure (enabled: true|false)

**Considered Options**:
1. Always expose all endpoints, use logging to control visibility
2. Make endpoint exposure configurable (chosen)
3. Use environment variables to toggle endpoints

**Chosen**: Option 2 (Configuration-based toggling)

**Rationale**:
- Security: Don't expose endpoints you don't need
- Kubernetes: Different clusters may have different observability requirements
- Clear intent: Configuration explicitly states what's exposed
- Future extensibility: Can add more conditional endpoints

**Trade-off**: Slightly more configuration verbosity, but clearer semantics

### Decision 3: Log Level Controls All Logging (Not Per-Endpoint)

**Considered Options**:
1. Global log level + per-endpoint overrides (rejected)
2. Single log level controls everything (chosen)

**Chosen**: Option 2 (Single log level)

**Rationale**:
- Simplicity: `server.log.level: info` means "log info and above"
- Consistency: All endpoints follow same rules
- KISS principle: No redundant configuration
- Behavior: At debug level, router logs all requests; at info level, router skips health/metrics

**Trade-off**: Less granular control, but matches common logging patterns

## Configuration Validation Strategy

### Startup Validation (Fail-Fast)

```
Required Fields Validation:
├── server.port: 1-65535
├── server.log.format: must be "json" or "text"
├── server.log.level: must be "debug", "info", "warn", or "error"
├── authentication.api_key: must not be empty (required)
├── machines: list must have ≥ 1 machine
└── machines[*]:
    ├── id: must not be empty (unique)
    ├── name: must not be empty
    ├── mac: must be valid MAC format
    └── broadcast: must be valid IPv4
```

### Error Messages

```
"Configuration validation failed:
  - authentication.api_key is required (set via config or GWAIHIR_API_KEY)
  - machine[0] validation failed: invalid MAC address"
```

## Environment Variable Mapping

```yaml
Environment Variable          Configuration Path
─────────────────────────────────────────────────
GWAIHIR_PORT                 → server.port
GWAIHIR_LOG_FORMAT           → server.log.format
GWAIHIR_LOG_LEVEL            → server.log.level
GWAIHIR_API_KEY              → authentication.api_key
GWAIHIR_CONFIG               → File path to load (default: /etc/gwaihir/gwaihir.yaml)
```

## Observability Implications

### Logging Behavior

Single `server.log.level` controls all logging behavior:
- `debug`: All requests logged including health checks and metrics  
- `info` and above: Only business operations logged; infrastructure endpoints silent

This eliminates redundant configuration flags and provides semantic clarity.

### Metrics

No changes to metrics collection. Same metrics are recorded; configuration only affects when endpoints are exposed and logged.

## Future Considerations

### Live Configuration Reload (v0.2)

```
Current (v0.1.1):
  Config load → Validate → Static until restart

Future (v0.2):
  Config load → Validate
         ↓
    Watch gwaihir.yaml
         ↓
    File changed → Validate new config
         ↓
    If valid: Apply (atomic swap)
    If invalid: Log error, keep old config
```

**Benefits of single-file approach**:
- One file to watch with fsnotify
- Atomic updates (all config changes together)
- Simple validation flow
- No race conditions between multiple files

### Bearer Token Support (v0.2)

```yaml
# Current v0.1.1
authentication:
  api_key: "secret-key"

# Future v0.2
authentication:
  api_key: "secret-key"
  bearer: "token-value"

# Middleware enhanced to check both:
# 1. Check Authorization: Bearer header
# 2. Fallback to X-API-Key header
```

## Rationale Summary

| Aspect | Before (Scattered) | After (Unified) | Benefit |
|--------|----------|--------|---------|
| Config Source | Multiple env vars + machines.yaml | Single gwaihir.yaml | Single source of truth |
| Configuration Discovery | Multiple places to check | One file | Easier to understand complete config |
| Validation | Runtime errors after startup | Startup validation | Fail-fast, prevents invalid states |
| Precedence | Implicit (env > defaults?) | Explicit hierarchy | Env > File > Defaults (clear) |
| Semantic Clarity | `GWAIHIR_PRODUCTION`, `LOG_JSON` | `log.level`, `log.format` | Self-documenting configuration |
| Observability Control | All endpoints exposed, always logged | Conditional + level-based | Explicit control, reduced noise |
| Deployment | Multiple env vars to set | Single config file | Simpler deployment, fewer errors |
| Future Extensibility | Scattered concerns | Unified structure | Ready for live reload (v0.2) |

## Risk Assessment

### Low Risk
- Simple, focused architectural change
- No changes to core business logic (domain layer unchanged)
- Configuration loading is independent concern
- Clear rollback path (revert to v0.1.0)
- Configuration validation ensures fail-fast behavior prevents invalid state

## References

### Related ADRs
- [2026-02-09 Service Architecture](./2026-02-09-service-architecture.md)

### Configuration Best Practices
- 12-Factor App: Configuration (http://12factor.net/config)
- Go Best Practices: Application Configuration
- YAML: Human-Readable Data Serialization Language

### Related Files
- [README.md](../../README.md)
- [AGENTS.md](../../AGENTS.md)
- [CONTRIBUTING.md](../../CONTRIBUTING.md)

---

**Document History**:
- **2026-02-10**: Initial version (v1.0) - Unified configuration proposal
