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
| `efficiency` | Parallel processing (Generate, Process) | `pkg/agent/task_service.go` |
| `logging` | Structured JSON logging via `log/slog` | `adapters/outbound/openai_client.go`, `adapters/outbound/tool_executor.go` |
| `messaging` | Message dispatcher for event publishing | `adapters/outbound/event_publisher.go` |
| `resource` | Generic CRUD storage (InMemory, JsonFile, etc.) | `adapters/outbound/conversation_store.go` |
| `security` | AES-GCM encryption for data at rest | `adapters/outbound/encrypted_conversation_store.go` |
| `service` | Context-aware function type `Function[IN, OUT]` | `adapters/outbound/openai_client.go`, `adapters/outbound/tool_executor.go` |
| `slices` | Generic slice utilities (Filter, Map, Unique, etc.) | `pkg/agent/agent.go`, `pkg/agent/tool_definition.go`, `adapters/outbound/openai_client.go` |
| `stability` | Resilience patterns (timeout, retry, circuit breaker, debounce) | `adapters/outbound/openai_client.go`, `adapters/outbound/tool_executor.go` |

#### When to Use

- **Testing assertions**: Use `assert.That(t, description, actual, expected)` for all test assertions
- **Structured logging**: Use `logging.NewJsonLogger()` for JSON-formatted logs with level control
- **Event publishing**: Use `messaging.Dispatcher` for publishing domain events
- **Slice transformations**: Use `slices.Filter`, `slices.Map`, `slices.Unique` for collection operations
- **Resilience patterns**: Use `stability.Timeout`, `stability.Retry`, `stability.Breaker`, `stability.Debounce` for external API calls
- **Parallel processing**: Use `efficiency.Generate`, `efficiency.Process` for concurrent workloads
- **Conversation persistence**: Use `resource.Access` with `InMemoryAccess`, `JsonFileAccess`, or `SqliteAccess`
- **Encryption at rest**: Use `security.Encrypt/Decrypt` with AES-GCM for sensitive data
- **Context-aware functions**: Use `service.Function[IN, OUT]` as the universal function signature

#### When NOT to Use

- Don't roll custom timeout/retry logic — use `stability` package instead
- Don't create new function signatures — align with `service.Function[IN, OUT]`
- Don't use external logging libraries (logrus, zap) — use `log/slog` via `logging.NewJsonLogger()`
- Don't write manual filter/map loops — use `slices.Filter`, `slices.Map` instead
- Don't implement custom storage backends — use `resource.Access` implementations

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

**Slice utilities (used in agent and adapters):**

```go
import "github.com/andygeiss/cloud-native-utils/slices"

// Filter: select elements matching predicate
requiredParams := slices.Filter(params, func(p ParameterDefinition) bool {
    return p.Required
})

// Map: transform elements to new type
names := slices.Map(requiredParams, func(p ParameterDefinition) string {
    return p.Name
})

// Combined Filter + Map pattern (common for extracting filtered fields)
completedCount := len(slices.Filter(tasks, func(t *Task) bool {
    return t.Status == TaskStatusCompleted
}))

// Other utilities available
slices.Contains(slice, element)     // Check if element exists
slices.ContainsAny(slice, elements) // Check if any elements exist
slices.Unique(slice)                // Remove duplicates
slices.First(slice)                 // Get first element (value, ok)
slices.Last(slice)                  // Get last element (value, ok)
slices.IndexOf(slice, element)      // Find element index (-1 if not found)
slices.Copy(slice)                  // Shallow copy
```

**Debounce (used in OpenAIClient):**

```go
import "github.com/andygeiss/cloud-native-utils/stability"

// Debounce coalesces rapid successive calls within a time window
// Useful for reducing API calls when input changes rapidly
client := outbound.NewOpenAIClient(baseURL, model).
    WithDebounce(500 * time.Millisecond)
```

**Parallel tool execution (used in TaskService):**

```go
import (
    "github.com/andygeiss/cloud-native-utils/efficiency"
    "github.com/andygeiss/cloud-native-utils/service"
)

// Enable parallel tool execution (uses CPU-count workers)
taskService := agent.NewTaskService(llm, executor, publisher).
    WithParallelToolExecution()

// Manual parallel processing pattern:
// 1. Generate channel from values
inCh := efficiency.Generate(toolCalls...)

// 2. Process with worker pool (runtime.NumCPU() workers)
processFn := func(ctx context.Context, tc ToolCall) (Result, error) {
    return executeToolCall(ctx, tc)
}
outCh, errCh := efficiency.Process(inCh, processFn)

// 3. Collect results
for result := range outCh {
    results = append(results, result)
}
```

**Conversation persistence (used in ConversationStore):**

```go
import "github.com/andygeiss/cloud-native-utils/resource"

// In-memory storage (for testing)
store := outbound.NewInMemoryConversationStore()

// JSON file storage (for persistence)
store := outbound.NewJsonFileConversationStore("conversations.json")

// Generic Access interface:
// - Create(ctx, key, value) error
// - Read(ctx, key) (*V, error)
// - ReadAll(ctx) ([]V, error)
// - Update(ctx, key, value) error
// - Delete(ctx, key) error

// Save/Load conversation history
_ = store.Save(ctx, agentID, messages)
messages, _ := store.Load(ctx, agentID)
_ = store.Clear(ctx, agentID)
```

**Encryption at rest (used in EncryptedConversationStore):**

```go
import "github.com/andygeiss/cloud-native-utils/security"

// Generate a 32-byte AES key (store securely!)
key := security.GenerateKey()

// Or load from environment variable (hex-encoded 32-byte key)
key := security.Getenv("ENCRYPTION_KEY")

// Encrypt plaintext
ciphertext := security.Encrypt([]byte("sensitive data"), key)

// Decrypt ciphertext
plaintext, err := security.Decrypt(ciphertext, key)

// Password hashing (bcrypt, cost 14)
hash, err := security.Password([]byte("p@ssw0rd"))
ok := security.IsPasswordValid(hash, []byte("p@ssw0rd"))

// Encrypted conversation store wrapper
baseStore := outbound.NewJsonFileConversationStore("conversations.json")
encryptedStore := outbound.NewEncryptedConversationStore(baseStore, key)
_ = encryptedStore.Save(ctx, agentID, messages)
messages, _ = encryptedStore.Load(ctx, agentID)
```

#### Available But Not Currently Used

The library offers many other packages that could be useful for future features:

| Package | Potential Use Case |
|---------|-------------------|

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
