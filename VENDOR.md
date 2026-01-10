# VENDOR.md

## Overview

This document describes the external vendor libraries used in go-agent and provides guidance on when and how to use them. The primary vendor is `cloud-native-utils`, a companion library providing reusable patterns for cloud-native Go applications.

**Guiding principle:** Prefer these approved vendors over custom implementations. If a vendor already provides functionality you need, use it.

---

## Approved Vendor Libraries

### cloud-native-utils

- **Purpose**: Comprehensive utility library for cloud-native Go applications providing resilience patterns, messaging, resource access, security, testing, and functional utilities.
- **Repository**: [github.com/andygeiss/cloud-native-utils](https://github.com/andygeiss/cloud-native-utils)
- **Version**: v0.4.12 (see `go.mod`)

This is the primary vendor dependency. It provides cross-cutting concerns that should **always** be used instead of rolling custom implementations.

---

## Package Reference (alphabetically sorted)

### assert — Testing Utilities

**Import**: `github.com/andygeiss/cloud-native-utils/assert`

**Purpose**: Fluent assertion library for tests.

**Key Functions** (alphabetically sorted):

| Function | Description |
|----------|-------------|
| `That(t, slice).HasLength(n)` | Slice length assertion |
| `That(t, str).Contains(substr)` | String contains assertion |
| `That(t, actual).IsEqualTo(expected)` | Equality assertion |
| `That(t, actual).IsFalse()` | Boolean false assertion |
| `That(t, actual).IsNil()` | Nil assertion |
| `That(t, actual).IsNotNil()` | Not-nil assertion |
| `That(t, actual).IsTrue()` | Boolean true assertion |

**When to use**:
- All unit tests
- Any test assertions
- Integration tests

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

### efficiency — Parallel Processing

**Import**: `github.com/andygeiss/cloud-native-utils/efficiency`

**Purpose**: Worker pool and channel-based parallel processing.

**Key Functions** (alphabetically sorted):

| Function | Description |
|----------|-------------|
| `Generate(items...)` | Creates a channel from items |
| `Process(inCh, fn)` | Processes items in parallel, returns output and error channels |

**When to use**:
- Batch processing
- I/O-bound operations that benefit from concurrency
- Parallel tool execution

**Integration pattern**: Domain layer (`internal/domain/agent/service.go`)

**Example** (from `service.go`):
```go
import "github.com/andygeiss/cloud-native-utils/efficiency"

// Generate input channel
inCh := efficiency.Generate(inputs...)

// Process in parallel
outCh, errCh := efficiency.Process(inCh, processorFn)

// Collect results
for result := range outCh {
    // handle result
}
```

**Cautions**:
- Collect all results before checking errors
- Results may arrive out of order
- Use for I/O-bound work; CPU-bound may not benefit

---

### event — Event Interface

**Import**: `github.com/andygeiss/cloud-native-utils/event`

**Purpose**: Defines the standard event interface for domain events.

**Key Types**:

| Type | Description |
|------|-------------|
| `Event` | Interface requiring `Topic() string` method |

**When to use**:
- Any event type that needs to be published
- Defining domain events

**Integration pattern**: Domain layer (`internal/domain/agent/events.go`)

**Example**:
```go
import "github.com/andygeiss/cloud-native-utils/event"

type EventTaskCompleted struct {
    Output string `json:"output"`
    TaskID string `json:"task_id"`
}

func (e EventTaskCompleted) Topic() string {
    return "agent.task.completed"
}
```

---

### messaging — Event Dispatching

**Import**: `github.com/andygeiss/cloud-native-utils/messaging`

**Purpose**: Decoupled event publishing and subscription.

**Key Types** (alphabetically sorted):

| Type | Description |
|------|-------------|
| `Dispatcher` | Interface for publishing messages |
| `Message` | Topic + payload container |
| `NewExternalDispatcher()` | Creates dispatcher for external event systems |
| `NewMessage(topic, payload)` | Creates a message with topic and byte payload |

**When to use**:
- Decoupling components via events
- Integration with external event systems (Kafka, etc.)
- Publishing domain events

**Integration pattern**: Adapters layer (`internal/adapters/outbound/event_publisher.go`)

**Example**:
```go
import "github.com/andygeiss/cloud-native-utils/messaging"

dispatcher := messaging.NewExternalDispatcher()
msg := messaging.NewMessage("agent.task.completed", jsonPayload)
dispatcher.Publish(ctx, msg)
```

---

### resource — Generic Storage Access

**Import**: `github.com/andygeiss/cloud-native-utils/resource`

**Purpose**: Generic CRUD interface with multiple backend implementations.

**Key Types** (alphabetically sorted):

| Type | Description |
|------|-------------|
| `Access[K, V]` | Generic interface for key-value storage |
| `NewInMemoryAccess[K, V]()` | In-memory storage (testing, ephemeral data) |
| `NewJsonFileAccess[K, V](path)` | JSON file persistence |
| `NewYamlFileAccess[K, V](path)` | YAML file persistence |

**Key Methods** on `Access[K, V]` (alphabetically sorted):
- `Create(ctx, key, value)` — Create new record
- `Delete(ctx, key)` — Delete record
- `Read(ctx, key)` — Read single record
- `ReadAll(ctx)` — Read all records
- `Update(ctx, key, value)` — Update existing record

**Error Constants**:
- `ErrorResourceAlreadyExists` — Returned when creating duplicate key
- `ErrorResourceNotFound` — Returned when key doesn't exist

**When to use**:
- Persisting domain entities
- Simple file-based storage (conversations, memory notes)
- Testing with in-memory backends

**Integration pattern**: Adapters layer (`internal/adapters/outbound/conversation_store.go`, `memory_store.go`)

**Example** (from adapters):
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

**Key Functions** (alphabetically sorted):

| Function | Description |
|----------|-------------|
| `Decrypt(ciphertext, key)` | Decrypts AES-GCM encrypted data |
| `Encrypt(plaintext, key)` | Encrypts data with AES-GCM |
| `GenerateKey()` | Generates a random 32-byte key |
| `Getenv(varName)` | Reads key from environment variable |

**When to use**:
- Any data that should not be readable in storage
- Encrypting sensitive data at rest
- Storing conversation history securely

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
- Each encryption generates a unique nonce (ciphertext is non-deterministic)
- Key must be exactly 32 bytes (AES-256)
- Store keys securely (environment variables, secrets manager)

---

### service — Service Patterns

**Import**: `github.com/andygeiss/cloud-native-utils/service`

**Purpose**: Common service patterns and function type definitions.

**Key Types**:

| Type | Description |
|------|-------------|
| `Function[I, O]` | Generic function signature `func(ctx, input) (output, error)` |

**When to use**:
- Defining typed functions for stability wrappers
- Service composition patterns

**Integration pattern**: Adapters layer (used with stability patterns)

**Example** (from `openai_client.go`):
```go
import "github.com/andygeiss/cloud-native-utils/service"

var fn service.Function[MyInput, MyOutput] = func(ctx context.Context, in MyInput) (MyOutput, error) {
    // implementation
}

// Chain with stability patterns
fn = stability.Timeout(fn, 120*time.Second)
fn = stability.Retry(fn, 3, 2*time.Second)
fn = stability.Breaker(fn, 5)
```

---

### slices — Functional Slice Utilities

**Import**: `github.com/andygeiss/cloud-native-utils/slices`

**Purpose**: Functional programming utilities for slices (filter, map, etc.).

**Key Functions** (alphabetically sorted):

| Function | Description |
|----------|-------------|
| `Contains(slice, element)` | Checks if element exists |
| `Filter(slice, predicate)` | Returns elements matching predicate |
| `Map(slice, transform)` | Transforms each element |
| `Sort(slice, less)` | Sorts slice with custom comparator |

**When to use**:
- Any slice manipulation
- Filtering collections (completed tasks, matching notes)
- Transforming data between layers

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

### stability — Resilience Patterns

**Import**: `github.com/andygeiss/cloud-native-utils/stability`

**Purpose**: Wraps functions with resilience patterns to handle failures gracefully.

**Key Functions** (alphabetically sorted):

| Function | Description |
|----------|-------------|
| `Breaker(fn, threshold)` | Opens circuit after consecutive failures, preventing cascading failures |
| `Debounce(fn, period)` | Coalesces rapid successive calls into a single execution |
| `Retry(fn, attempts, delay)` | Retries failed calls with fixed delay between attempts |
| `Throttle(fn, tokens, refill, period)` | Rate limits calls using token bucket algorithm |
| `Timeout(fn, duration)` | Cancels execution if it exceeds the given duration |

**When to use**:
- Any external API call (LLM, HTTP, database)
- Operations that may fail transiently
- Protecting downstream systems from overload
- Rate-limited services

**Integration pattern**: Adapters layer (`internal/adapters/outbound/`)

**Example** (from `openai_client.go`):
```go
import "github.com/andygeiss/cloud-native-utils/stability"

// Wrap with stability patterns (innermost to outermost)
var fn service.Function[Input, Output] = baseFn
fn = stability.Timeout(fn, 120*time.Second)  // 1. Timeout per attempt
fn = stability.Retry(fn, 3, 2*time.Second)   // 2. Retry wraps timeout
fn = stability.Breaker(fn, 5)                 // 3. Breaker prevents retries when open
if throttleEnabled {
    fn = stability.Throttle(fn, tokens, refill, period)  // 4. Rate limit
}
if debounceEnabled {
    fn = stability.Debounce(fn, period)       // 5. Coalesce rapid calls
}
```

**Cautions**:
- Circuit breaker should be outermost to prevent retrying when circuit is open
- Order matters: apply innermost first (timeout → retry → breaker → throttle → debounce)
- Retry should wrap timeout so each attempt has its own timeout
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

## Cross-cutting Concerns and Recommended Patterns (alphabetically sorted)

### Event Publishing
- Define events in domain with `event.Event` interface
- Keep events immutable with constructor functions
- Publish via adapter using `messaging.Dispatcher`

### Resilience
- **Always** use `stability` package for external calls
- Apply patterns in order: timeout → retry → breaker → throttle → debounce
- Configure defaults, allow override via `With*` methods

### Storage
- In-memory for tests, file-based for development, extend for production databases
- Use `resource.Access` for persistence
- Wrap with `security.Encrypt/Decrypt` when needed

### Testing
- Prefer in-memory backends for unit tests
- Table-driven tests with descriptive names
- Use `assert` package for all assertions

---

## Vendors to Avoid

| Library | Reason | Alternative |
|---------|--------|-------------|
| `github.com/stretchr/testify` | Duplication with `cloud-native-utils/assert` | Use `assert` |
| Custom circuit breakers | Already provided by stability | Use `stability.Breaker` |
| Custom retry loops | Already provided by stability | Use `stability.Retry` |
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
- [cloud-native-utils GoDoc](https://pkg.go.dev/github.com/andygeiss/cloud-native-utils) — Package documentation
