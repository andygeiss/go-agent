# VENDOR.md

## Overview

This document describes the external vendor libraries used in go-agent and provides guidance on when and how to use them. The primary vendor is `cloud-native-utils`, a companion library providing reusable patterns for cloud-native Go applications.

**Guiding principle:** Prefer these approved vendors over custom implementations. If a vendor already provides functionality you need, use it.

---

## Approved Vendor Libraries

### cloud-native-utils

- **Purpose**: Comprehensive utility library for cloud-native Go applications providing resilience patterns, messaging, resource access, testing, and functional utilities.
- **Repository**: [github.com/andygeiss/cloud-native-utils](https://github.com/andygeiss/cloud-native-utils)
- **Version**: v0.4.12+ (see `go.mod`)

This is the primary vendor dependency. It provides cross-cutting concerns that should **always** be used instead of rolling custom implementations.

---

## Package Reference

### stability — Resilience Patterns

**Import**: `github.com/andygeiss/cloud-native-utils/stability`

**Purpose**: Wraps functions with resilience patterns to handle failures gracefully.

**Key Functions**:

| Function | Description |
|----------|-------------|
| `Timeout(fn, duration)` | Cancels execution if it exceeds the given duration |
| `Retry(fn, attempts, delay)` | Retries failed calls with fixed delay between attempts |
| `Breaker(fn, threshold)` | Opens circuit after consecutive failures, preventing cascading failures |
| `Throttle(fn, tokens, refill, period)` | Rate limits calls using token bucket algorithm |
| `Debounce(fn, period)` | Coalesces rapid successive calls into a single execution |

**When to use**:
- Any external API call (LLM, HTTP, database)
- Operations that may fail transiently
- Rate-limited services
- Protecting downstream systems from overload

**Integration pattern**: Adapters layer (`internal/adapters/outbound/`)

**Example** (from `openai_client.go`):
```go
import "github.com/andygeiss/cloud-native-utils/stability"

// Wrap with stability patterns (innermost to outermost)
var fn service.Function[Input, Output] = baseFn
fn = stability.Timeout(fn, 120*time.Second)
fn = stability.Retry(fn, 3, 2*time.Second)
fn = stability.Breaker(fn, 5)
```

**Cautions**:
- Order matters: apply innermost first (timeout → retry → breaker)
- Retry should wrap timeout so each attempt has its own timeout
- Circuit breaker should be outermost to prevent retrying when circuit is open

---

### messaging — Event Dispatching

**Import**: `github.com/andygeiss/cloud-native-utils/messaging`

**Purpose**: Decoupled event publishing and subscription.

**Key Types**:

| Type | Description |
|------|-------------|
| `Dispatcher` | Interface for publishing messages |
| `Message` | Topic + payload container |
| `NewExternalDispatcher()` | Creates dispatcher for external event systems |
| `NewMessage(topic, payload)` | Creates a message with topic and byte payload |

**When to use**:
- Publishing domain events
- Decoupling components via events
- Integration with external event systems (Kafka, etc.)

**Integration pattern**: Adapters layer (`internal/adapters/outbound/event_publisher.go`)

**Example**:
```go
import "github.com/andygeiss/cloud-native-utils/messaging"

dispatcher := messaging.NewExternalDispatcher()
msg := messaging.NewMessage("agent.task.completed", jsonPayload)
dispatcher.Publish(ctx, msg)
```

---

### event — Event Interface

**Import**: `github.com/andygeiss/cloud-native-utils/event`

**Purpose**: Defines the standard event interface for domain events.

**Key Types**:

| Type | Description |
|------|-------------|
| `Event` | Interface requiring `Topic() string` method |

**When to use**:
- Defining domain events
- Any event type that needs to be published

**Integration pattern**: Domain layer (`internal/domain/agent/events.go`)

**Example**:
```go
import "github.com/andygeiss/cloud-native-utils/event"

type EventTaskCompleted struct {
    TaskID string `json:"task_id"`
    Output string `json:"output"`
}

func (e EventTaskCompleted) Topic() string {
    return "agent.task.completed"
}
```

---

### resource — Generic Storage Access

**Import**: `github.com/andygeiss/cloud-native-utils/resource`

**Purpose**: Generic CRUD interface with multiple backend implementations.

**Key Types**:

| Type | Description |
|------|-------------|
| `Access[K, V]` | Generic interface for key-value storage |
| `NewInMemoryAccess[K, V]()` | In-memory storage (testing, ephemeral data) |
| `NewJsonFileAccess[K, V](path)` | JSON file persistence |
| `NewYamlFileAccess[K, V](path)` | YAML file persistence |

**Key Methods** on `Access[K, V]`:
- `Create(ctx, key, value)` — Create new record
- `Read(ctx, key)` — Read single record
- `ReadAll(ctx)` — Read all records
- `Update(ctx, key, value)` — Update existing record
- `Delete(ctx, key)` — Delete record

**When to use**:
- Persisting domain entities
- Testing with in-memory backends
- Simple file-based storage (conversations, memory notes)

**Integration pattern**: Adapters layer (`internal/adapters/outbound/memory_store.go`, `conversation_store.go`)

**Example**:
```go
import "github.com/andygeiss/cloud-native-utils/resource"

// Production: file-based
access := resource.NewJsonFileAccess[string, MyEntity]("data.json")

// Testing: in-memory
access := resource.NewInMemoryAccess[string, MyEntity]()

// Use generically
err := access.Create(ctx, "key-1", entity)
entity, err := access.Read(ctx, "key-1")
```

**Cautions**:
- `ErrorResourceAlreadyExists` is returned as string, check `err.Error()`
- File-based access creates file on first write
- No built-in concurrency protection for file access

---

### security — Encryption Utilities

**Import**: `github.com/andygeiss/cloud-native-utils/security`

**Purpose**: AES-GCM encryption for data at rest.

**Key Functions**:

| Function | Description |
|----------|-------------|
| `Encrypt(plaintext, key)` | Encrypts data with AES-GCM |
| `Decrypt(ciphertext, key)` | Decrypts AES-GCM encrypted data |
| `GenerateKey()` | Generates a random 32-byte key |
| `Getenv(varName)` | Reads key from environment variable |

**When to use**:
- Encrypting sensitive data at rest
- Storing conversation history securely
- Any data that should not be readable in storage

**Integration pattern**: Adapters layer (`internal/adapters/outbound/encrypted_conversation_store.go`)

**Example**:
```go
import "github.com/andygeiss/cloud-native-utils/security"

// Generate or load key
key := security.GenerateKey() // or security.Getenv("ENCRYPTION_KEY")

// Encrypt
ciphertext := security.Encrypt(plaintext, key)

// Decrypt
plaintext, err := security.Decrypt(ciphertext, key)
```

**Cautions**:
- Key must be exactly 32 bytes (AES-256)
- Store keys securely (environment variables, secrets manager)
- Each encryption generates a unique nonce (ciphertext is non-deterministic)

---

### slices — Functional Slice Utilities

**Import**: `github.com/andygeiss/cloud-native-utils/slices`

**Purpose**: Functional programming utilities for slices (filter, map, etc.).

**Key Functions**:

| Function | Description |
|----------|-------------|
| `Filter(slice, predicate)` | Returns elements matching predicate |
| `Map(slice, transform)` | Transforms each element |
| `Contains(slice, element)` | Checks if element exists |
| `Sort(slice, less)` | Sorts slice with custom comparator |

**When to use**:
- Filtering collections (completed tasks, matching notes)
- Transforming data between layers
- Any slice manipulation

**Integration pattern**: Domain layer (anywhere slices are processed)

**Example** (from `agent.go`):
```go
import "github.com/andygeiss/cloud-native-utils/slices"

func (a *Agent) CompletedTaskCount() int {
    return len(slices.Filter(a.Tasks, func(t *Task) bool {
        return t.Status == TaskStatusCompleted
    }))
}
```

---

### efficiency — Parallel Processing

**Import**: `github.com/andygeiss/cloud-native-utils/efficiency`

**Purpose**: Worker pool and channel-based parallel processing.

**Key Functions**:

| Function | Description |
|----------|-------------|
| `Generate(items...)` | Creates a channel from items |
| `Process(inCh, fn)` | Processes items in parallel, returns output and error channels |

**When to use**:
- Parallel tool execution
- Batch processing
- I/O-bound operations that benefit from concurrency

**Integration pattern**: Domain layer (`internal/domain/agent/task_service.go`)

**Example**:
```go
import "github.com/andygeiss/cloud-native-utils/efficiency"

// Generate input channel
inCh := efficiency.Generate(items...)

// Process in parallel
outCh, errCh := efficiency.Process(inCh, processorFn)

// Collect results
for result := range outCh {
    // handle result
}
```

**Cautions**:
- Results may arrive out of order
- Collect all results before checking errors
- Use for I/O-bound work; CPU-bound may not benefit

---

### service — Service Patterns

**Import**: `github.com/andygeiss/cloud-native-utils/service`

**Purpose**: Common service patterns and function type definitions.

**Key Types**:

| Type | Description |
|----------|-------------|
| `Function[I, O]` | Generic function signature `func(ctx, input) (output, error)` |

**When to use**:
- Defining typed functions for stability wrappers
- Service composition patterns

**Integration pattern**: Adapters layer (used with stability patterns)

**Example**:
```go
import "github.com/andygeiss/cloud-native-utils/service"

var fn service.Function[MyInput, MyOutput] = func(ctx context.Context, in MyInput) (MyOutput, error) {
    // implementation
}
```

---

### assert — Testing Utilities

**Import**: `github.com/andygeiss/cloud-native-utils/assert`

**Purpose**: Fluent assertion library for tests.

**Key Functions**:

| Function | Description |
|----------|-------------|
| `That(t, actual).IsEqualTo(expected)` | Equality assertion |
| `That(t, actual).IsNil()` | Nil assertion |
| `That(t, actual).IsNotNil()` | Not-nil assertion |
| `That(t, actual).IsTrue()` | Boolean true assertion |
| `That(t, actual).IsFalse()` | Boolean false assertion |
| `That(t, slice).HasLength(n)` | Slice length assertion |
| `That(t, str).Contains(substr)` | String contains assertion |

**When to use**:
- All unit tests
- Integration tests
- Any test assertions

**Integration pattern**: Test files (`*_test.go`)

**Example**:
```go
import "github.com/andygeiss/cloud-native-utils/assert"

func TestMyFunction(t *testing.T) {
    result := MyFunction()
    assert.That(t, result).IsEqualTo(expected)
    assert.That(t, err).IsNil()
}
```

---

## Cross-cutting Concerns and Recommended Patterns

### Resilience
- **Always** use `stability` package for external calls
- Apply patterns in order: timeout → retry → breaker → throttle → debounce
- Configure defaults, allow override via `With*` methods

### Event Publishing
- Define events in domain with `event.Event` interface
- Publish via adapter using `messaging.Dispatcher`
- Keep events immutable with constructor functions

### Storage
- Use `resource.Access` for persistence
- In-memory for tests, file-based for development, extend for production databases
- Wrap with encryption when needed

### Testing
- Use `assert` package for all assertions
- Prefer in-memory backends for unit tests
- Table-driven tests with descriptive names

---

## Vendors to Avoid

| Library | Reason | Alternative |
|---------|--------|-------------|
| `github.com/stretchr/testify` | Duplication with `cloud-native-utils/assert` | Use `assert` |
| Custom retry loops | Already provided by stability | Use `stability.Retry` |
| Custom circuit breakers | Already provided by stability | Use `stability.Breaker` |
| Manual encryption | Complex, error-prone | Use `security.Encrypt/Decrypt` |

---

## Migration and Version Notes

### v0.4.12 (Current)

No breaking changes from v0.4.x series. All patterns documented above are stable.

### Upgrading

```bash
go get github.com/andygeiss/cloud-native-utils@latest
go mod tidy
```

Check [cloud-native-utils releases](https://github.com/andygeiss/cloud-native-utils/releases) for version-specific notes.

---

## Related Documentation

- [CONTEXT.md](CONTEXT.md) — Architecture and conventions
- [README.md](README.md) — User-facing documentation
- [cloud-native-utils](https://pkg.go.dev/github.com/andygeiss/cloud-native-utils) — Package documentation
