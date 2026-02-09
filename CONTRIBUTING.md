# Contributing to Gwaihir

Thank you for considering contributing to Gwaihir! This document outlines our development process, coding standards, and testing requirements.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Standards](#code-standards)
- [Testing Requirements](#testing-requirements)
- [Commit Messages](#commit-messages)
- [Pull Request Process](#pull-request-process)

## Code of Conduct

This project adheres to professional engineering standards:

- **Be respectful**: Treat all contributors with respect
- **Be constructive**: Provide actionable feedback in code reviews
- **Be collaborative**: Help others learn and improve
- **Be thorough**: Write comprehensive tests and documentation

## Getting Started

### Prerequisites

- Go 1.22.2 or later
- [golangci-lint](https://golangci-lint.run/usage/install/)
- [just](https://github.com/casey/just)
- Git

### Development Setup

```bash
# Clone the repository
git clone https://github.com/josimar-silva/gwaihir.git
cd gwaihir

# Install dependencies
just ci

# Run tests to verify setup
just test

# Run linter
just lint
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Changes

Follow the [Code Standards](#code-standards) outlined below.

### 3. Run Pre-commit Checks

Before committing, ensure all checks pass:

```bash
just pre-commit
```

This runs:
- Code formatting (`gofmt`, `goimports`)
- Linting (`golangci-lint`)
- All tests with coverage

### 4. Commit Changes

Follow the [Commit Message](#commit-messages) conventions.

### 5. Push and Create PR

```bash
git push origin your-branch-name
```

Then create a Pull Request on GitHub.

## Code Standards

### Clean Architecture Principles

Gwaihir follows Clean Architecture with these layers:

1. **Domain** (`internal/domain`): Core business entities and interfaces
   - No external dependencies
   - Pure Go code, no frameworks
   - Contains interfaces that outer layers implement

2. **Use Case** (`internal/usecase`): Business logic
   - Orchestrates domain entities
   - Depends only on domain interfaces
   - No HTTP/DB/infrastructure concerns

3. **Delivery** (`internal/delivery/http`): HTTP handlers
   - Translates HTTP requests to use case calls
   - Handles authentication and validation
   - Returns appropriate HTTP responses

4. **Repository** (`internal/repository`): Data access
   - Implements domain repository interfaces
   - Handles data persistence (YAML, DB, etc.)

5. **Infrastructure** (`internal/infrastructure`): Infrastructure concerns
   - Logging, metrics, external services
   - Implements domain interfaces for infrastructure

### Code Style

**Follow Go idioms:**

```go
// Good: Exported function with clear doc comment
// SendWakePacket sends a WoL packet to the specified machine.
// It validates that the machine is in the allowlist before sending.
func (uc *WoLUseCase) SendWakePacket(machineID string) error {
    // Implementation
}

// Bad: Redundant comment
// SendWakePacket sends wake packet
func (uc *WoLUseCase) SendWakePacket(machineID string) error {
    // Implementation
}
```

**Error handling:**

```go
// Good: Wrap errors with context
if err := uc.packetSender.SendMagicPacket(machine.MAC, machine.Broadcast); err != nil {
    return fmt.Errorf("failed to send WoL packet: %w", err)
}

// Bad: Swallow errors or return without context
if err := uc.packetSender.SendMagicPacket(machine.MAC, machine.Broadcast); err != nil {
    return err
}
```

**Keep functions small:**

- Max 50 lines per function
- Single responsibility principle
- Extract complex logic into helper functions

**Use meaningful names:**

```go
// Good
func (m *Machine) NormalizeMAC() string

// Bad
func (m *Machine) Normalize() string  // What are we normalizing?
func (m *Machine) DoStuff() string    // Too vague
```

### Dependency Injection

Always use dependency injection for testability:

```go
// Good: Dependencies injected
func NewWoLUseCase(
    machineRepo domain.MachineRepository,
    packetSender domain.WoLPacketSender,
    logger *infrastructure.Logger,
    metrics *infrastructure.Metrics,
) *WoLUseCase {
    return &WoLUseCase{
        machineRepo:  machineRepo,
        packetSender: packetSender,
        logger:       logger,
        metrics:      metrics,
    }
}

// Bad: Hidden dependencies (global variables, singletons)
func NewWoLUseCase() *WoLUseCase {
    return &WoLUseCase{
        logger: Logger.Instance(), // Singleton
    }
}
```

## Testing Requirements

### Coverage Targets

**Minimum Requirements:**
- Overall project coverage: **80%**
- New code coverage: **90%+**
- Critical paths (WoL sending, auth): **100%**

**Current Coverage:**
- Domain: 95%
- Use Case: 100%
- Repository: 90.6%
- Infrastructure: 93.1%
- HTTP Delivery: 94%

### Test Organization

**File naming:**
- Test files: `*_test.go` (same package)
- Mock implementations: In test file or `internal/mocks/`

**Test function naming:**

```go
// Pattern: Test<Function>_<Scenario>
func TestSendWakePacket_Success(t *testing.T)
func TestSendWakePacket_MachineNotFound(t *testing.T)
func TestSendWakePacket_NetworkError(t *testing.T)
```

### Test Structure

Use **Arrange-Act-Assert** pattern:

```go
func TestSendWakePacket_Success(t *testing.T) {
    // Arrange: Set up test dependencies
    machines := map[string]*domain.Machine{
        "saruman": {
            ID:        "saruman",
            MAC:       "AA:BB:CC:DD:EE:FF",
            Broadcast: "192.168.1.255",
        },
    }
    repo := newMockMachineRepository(machines)
    sender := newMockWoLPacketSender()
    logger := newTestLogger()
    metrics := newTestMetrics()
    useCase := NewWoLUseCase(repo, sender, logger, metrics)

    // Act: Execute the function being tested
    err := useCase.SendWakePacket("saruman")

    // Assert: Verify the expected outcome
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if sender.callCount != 1 {
        t.Errorf("Expected 1 packet sent, got %d", sender.callCount)
    }
}
```

### Table-Driven Tests

For testing multiple scenarios:

```go
func TestMachine_Validate(t *testing.T) {
    tests := []struct {
        name    string
        machine Machine
        wantErr bool
    }{
        {
            name: "valid machine",
            machine: Machine{
                ID:        "server1",
                Name:      "Test Server",
                MAC:       "AA:BB:CC:DD:EE:FF",
                Broadcast: "192.168.1.255",
            },
            wantErr: false,
        },
        {
            name: "invalid MAC",
            machine: Machine{
                ID:        "server1",
                Name:      "Test Server",
                MAC:       "invalid",
                Broadcast: "192.168.1.255",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.machine.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Mock Implementations

**Guidelines:**
- Keep mocks simple and focused
- Allow configurable behavior (errors, return values)
- Track invocations for verification

**Example:**

```go
type mockWoLPacketSender struct {
    sendPackets []sentPacket
    sendError   error
    callCount   int
}

func (m *mockWoLPacketSender) SendMagicPacket(mac, broadcast string) error {
    m.callCount++
    m.sendPackets = append(m.sendPackets, sentPacket{mac: mac, broadcast: broadcast})

    if m.sendError != nil {
        return m.sendError
    }
    return nil
}
```

### What to Test

**Required test coverage:**

1. **Domain Layer**
   - Entity validation (all validation rules)
   - Business logic functions
   - Edge cases (empty values, invalid formats)

2. **Use Case Layer**
   - Happy path (successful operations)
   - Error paths (not found, send failures)
   - Repository errors
   - Infrastructure failures

3. **HTTP Layer**
   - All HTTP status codes (200, 202, 400, 401, 404, 500, 503)
   - Request validation
   - Authentication (valid key, missing key, invalid key)
   - Response formats

4. **Repository Layer**
   - YAML parsing (valid, invalid, malformed)
   - Cache behavior
   - Duplicate ID detection

5. **Infrastructure Layer**
   - WoL packet structure (102 bytes, correct format)
   - Network errors
   - Metrics increment on operations
   - Log formatting

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
just test

# Run specific package
go test ./internal/usecase

# Run specific test
go test -run TestSendWakePacket_Success ./internal/usecase

# Run with race detector
go test -race ./...

# Verbose output
go test -v ./...
```

### Code Coverage Report

After running tests, view the coverage report:

```bash
# Generate and open HTML coverage report
just test && open coverage.html

# View coverage by function
go tool cover -func=coverage.out
```

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Build process, tooling changes
- `ci`: CI/CD changes

### Examples

**Feature:**
```
feat(usecase): add WoL packet sending

Implement actual WoL magic packet sending via UDP broadcast.
Includes validation for MAC addresses and broadcast IPs.

Closes #42
```

**Bug fix:**
```
fix(auth): handle missing API key gracefully

Return 401 Unauthorized instead of 500 when API key is missing.

Fixes #56
```

**Refactoring:**
```
refactor(handler): extract validation logic

Extract common validation logic into reusable helpers to reduce
duplication across HTTP handlers.
```

**Documentation:**
```
docs(readme): add Prometheus metrics examples

Add example queries and alert rules for monitoring Gwaihir
via Prometheus.
```

## Pull Request Process

### Before Opening a PR

1. **Run pre-commit checks:**
   ```bash
   just pre-commit
   ```

2. **Ensure tests pass:**
   ```bash
   just test
   ```

3. **Verify coverage meets requirements:**
   ```bash
   # Coverage should be >80% overall, >90% for new code
   go tool cover -func=coverage.out
   ```

4. **Update documentation:**
   - Update README.md if adding features
   - Update API examples if changing endpoints
   - Add/update code comments for complex logic

### PR Description Template

```markdown
## Description

Brief description of changes and motivation.

## Type of Change

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated (if applicable)
- [ ] Manual testing performed
- [ ] Coverage requirements met (>90% for new code)

## Checklist

- [ ] Code follows the project's style guidelines
- [ ] Self-review performed
- [ ] Comments added for complex/non-obvious code
- [ ] Documentation updated
- [ ] No new warnings from linter
- [ ] Tests pass locally
- [ ] Coverage meets requirements

## Related Issues

Closes #<issue_number>
```

### Code Review

**As a reviewer:**
- Check for adherence to Clean Architecture
- Verify test coverage and quality
- Look for potential bugs or edge cases
- Ensure documentation is updated
- Provide constructive feedback

**As an author:**
- Respond to all comments
- Make requested changes promptly
- Explain technical decisions when needed
- Be open to alternative approaches

### Merging

PRs are merged when:
- All CI checks pass
- At least one approval from a maintainer
- All review comments are resolved
- Coverage requirements are met
- Documentation is updated

## Questions?

If you have questions or need clarification:
- Open an issue for discussion
- Check existing issues and PRs
- Review the architecture documentation in `docs/`

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
