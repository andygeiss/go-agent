# VENDOR.md

## Overview

This document catalogs the external vendor libraries used in this project, explaining their purpose, recommended usage patterns, and integration points. The goal is to help developers and AI agents understand which libraries to use for specific concerns and avoid duplicating vendor functionality.

---

## Approved Vendor Libraries

### github.com/andygeiss/cloud-native-utils

- **Purpose**: A modular Go library providing reusable utilities for cloud-native applications. In this project, we primarily use the `assert` package for testing.
- **Repository**: https://github.com/andygeiss/cloud-native-utils
- **Version**: v0.4.11

#### Key Packages Used

| Package | Description | Used In |
|---------|-------------|---------|
| `assert` | Minimal test assertion helper | All `*_test.go` files |
| `logging` | Structured JSON logging via `log/slog` | `adapters/outbound/openai_client.go`, `adapters/outbound/tool_executor.go` |
| `messaging` | Message dispatcher for event publishing | `adapters/outbound/event_publisher.go` |
| `service` | Context-aware function type `Function[IN, OUT]` | `adapters/outbound/openai_client.go`, `adapters/outbound/tool_executor.go` |
| `stability` | Resilience patterns (timeout, retry, circuit breaker) | `adapters/outbound/openai_client.go`, `adapters/outbound/tool_executor.go` |

#### When to Use

- **Testing assertions**: Use `assert.That(t, description, actual, expected)` for all test assertions
- **Structured logging**: Use `logging.NewJsonLogger()` for JSON-formatted logs with level control
- **Event publishing**: Use `messaging.Dispatcher` for publishing domain events
- **Resilience patterns**: Use `stability.Timeout`, `stability.Retry`, `stability.Breaker` for external API calls
- **Context-aware functions**: Use `service.Function[IN, OUT]` as the universal function signature

#### When NOT to Use

- Don't roll custom timeout/retry logic — use `stability` package instead
- Don't create new function signatures — align with `service.Function[IN, OUT]`
- Don't use external logging libraries (logrus, zap) — use `log/slog` via `logging.NewJsonLogger()`

#### Integration Patterns

**Testing with assert:**

```go
import (
    "testing"
    "github.com/andygeiss/cloud-native-utils/assert"
)

func Test_Example_Should_Work(t *testing.T) {
    // Arrange
    input := "test"
    
    // Act
    result := processInput(input)
    
    // Assert
    assert.That(t, "result must match expected", result, "expected")
}
```

**Resilience with stability (used in OpenAIClient):**

```go
import (
    "github.com/andygeiss/cloud-native-utils/service"
    "github.com/andygeiss/cloud-native-utils/stability"
)

// Define base function matching service.Function[IN, OUT] signature
baseFn := func(ctx context.Context, in Input) (Output, error) {
    return doWork(ctx, in)
}

// Wrap with stability patterns (innermost to outermost):
// 1. Timeout - enforce maximum execution time
// 2. Retry - handle transient failures  
// 3. Circuit Breaker - prevent cascading failures
var fn service.Function[Input, Output] = baseFn
fn = stability.Timeout(fn, 30*time.Second)
fn = stability.Retry(fn, 3, 2*time.Second)
fn = stability.Breaker(fn, 5)

result, err := fn(ctx, input)
```

**Tool execution with timeout:**

```go
import "github.com/andygeiss/cloud-native-utils/stability"

wrappedFn := stability.Timeout(toolFn, 30*time.Second)
result, err := wrappedFn(ctx, args)
```

**Structured logging (used in adapters):**

```go
import "github.com/andygeiss/cloud-native-utils/logging"

// Create a JSON logger (level controlled by LOGGING_LEVEL env var)
logger := logging.NewJsonLogger()

// Inject into adapters
client := outbound.NewOpenAIClient(baseURL, model).
    WithLogger(logger)

executor := outbound.NewToolExecutor().
    WithLogger(logger)
```

Log levels (via `LOGGING_LEVEL` environment variable):
- `DEBUG` — All logs including request/response details
- `INFO` — Default, general operational logs
- `WARN` — Warning conditions
- `ERROR` — Error conditions only

**Rate limiting with throttle (used in OpenAIClient):**

```go
import "github.com/andygeiss/cloud-native-utils/stability"

// Token bucket rate limiting:
// - maxTokens: 10 calls allowed in bucket
// - refill: 2 tokens added per period
// - period: 1 second refill interval
client := outbound.NewOpenAIClient(baseURL, model).
    WithThrottle(10, 2, time.Second)
```

When rate limit is exceeded, returns `stability.ErrorThrottleTooManyCalls`.

#### Available But Not Currently Used

The library offers many other packages that could be useful for future features:

| Package | Potential Use Case |
|---------|-------------------|
| `security` | AES encryption, password hashing if needed |
| `slices` | Generic slice utilities (Map, Filter, Unique) |
| `stability.Debounce` | Coalescing rapid successive calls |

---

## Standard Library Dependencies

The project relies heavily on Go's standard library for core functionality:

| Package | Usage |
|---------|-------|
| `context` | Request cancellation and timeouts |
| `encoding/json` | JSON serialization for LLM API |
| `log/slog` | Structured logging (via cloud-native-utils/logging) |
| `net/http` | HTTP client for OpenAI-compatible APIs |
| `time` | Timestamps on entities |
| `testing` | Test framework (with cloud-native-utils/assert) |
| `slices` | `slices.Contains()` for collection operations |

---

## Indirect Dependencies

These are transitive dependencies pulled in by `cloud-native-utils`. They are not used directly:

| Dependency | Brought In By | Purpose |
|------------|---------------|---------|
| `github.com/segmentio/kafka-go` | messaging | Kafka client (unused) |
| `github.com/coreos/go-oidc/v3` | security | OIDC authentication (unused) |
| `github.com/klauspost/compress` | efficiency | Compression (unused) |
| `golang.org/x/crypto` | security | Cryptographic functions (unused) |
| `golang.org/x/oauth2` | security | OAuth2 flows (unused) |

These do not affect the runtime unless their packages are imported.

---

## Cross-Cutting Concerns and Recommended Patterns

### Testing
- **Preferred**: `github.com/andygeiss/cloud-native-utils/assert`
- **Pattern**: `assert.That(t, "description", actual, expected)`
- **Naming**: `Test_<Type>_<Method>_<Scenario>_Should_<Expected>`

### HTTP Clients
- **Preferred**: Go standard library `net/http`
- **Pattern**: Wrap in adapter implementing port interface
- **Location**: `internal/adapters/outbound/`

### JSON Handling
- **Preferred**: Go standard library `encoding/json`
- **Pattern**: Use struct tags for field mapping

### Concurrency & Resilience
- **Preferred**: `cloud-native-utils/stability` for resilience patterns
- **Patterns**:
  - `stability.Timeout(fn, duration)` — enforce execution time limits
  - `stability.Retry(fn, attempts, delay)` — retry transient failures
  - `stability.Breaker(fn, threshold)` — circuit breaker with exponential backoff
- **Function signature**: `service.Function[IN, OUT]` for compatibility
- **Location**: Wrap external calls in `internal/adapters/outbound/`

### Event Publishing
- **Current**: `EventPublisher` adapter wrapping `cloud-native-utils/messaging`
- **Pattern**: Events are JSON-encoded and published via dispatcher
- **Testing**: Mock dispatcher for unit tests (`event_publisher_test.go`)
- **Future**: Consider full Kafka integration for production

---

## Adding New Vendor Dependencies

Before adding a new dependency:

1. **Check if cloud-native-utils provides it** - Prefer using existing transitive dependencies
2. **Check if standard library suffices** - Go stdlib is often enough
3. **Evaluate the library**:
   - Is it actively maintained?
   - Does it have a permissive license (MIT, Apache 2.0)?
   - Is it minimal and focused?
4. **Document in this file** - Add a new section following the template above
5. **Use adapter pattern** - Wrap external libraries behind interfaces in `internal/adapters/`

### Dependency Template

```markdown
### <package-path>

- **Purpose**: Brief description
- **Repository**: URL
- **Version**: vX.Y.Z

#### Key Packages Used
| Package | Description | Used In |
|---------|-------------|---------|
| `pkg` | What it does | Where used |

#### When to Use
- Specific use cases

#### When NOT to Use
- Anti-patterns or alternatives

#### Integration Pattern
```go
// Example code
```
```

---

## Vendors to Avoid

| Library | Reason | Alternative |
|---------|--------|-------------|
| `testify` | Project uses `cloud-native-utils/assert` | `assert.That()` |
| `logrus` | Project uses stdlib `log/slog` | Standard library |
| `gorilla/mux` | Overkill for current needs | `net/http` stdlib |

---

## Version Management

- Dependencies are managed via `go.mod`
- Run `go mod tidy` after adding/removing imports
- Update dependencies with `go get -u <package>@latest`
- Pin to specific versions for stability

---

## Related Documentation

- [CONTEXT.md](CONTEXT.md) - Project architecture and conventions
- [README.md](README.md) - User-facing documentation
- [go.mod](go.mod) - Dependency manifest
