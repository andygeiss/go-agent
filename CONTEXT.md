# CONTEXT.md

This document serves as the authoritative project context for AI coding agents, retrieval systems, and developers. It describes the architecture, conventions, and contracts of the go-agent repository.

---

## 1. Project Purpose

Go Agent is a **production-ready** Go library that implements the **Observe → Decide → Act → Update** loop pattern for building LLM-powered applications. It provides:

- A domain-driven agent framework with typed entities (Agent, Task, Message, ToolCall)
- Port/adapter architecture for pluggable LLM clients and tool executors
- Event-driven observability through domain events
- Lifecycle hooks for cross-cutting concerns (logging, metrics, authorization)
- Parallel tool execution for improved performance with I/O-bound operations
- Conversation persistence with in-memory and JSON file storage backends
- AES-GCM encryption for sensitive conversation data at rest
- Resilience patterns (timeout, retry, circuit breaker, debounce, throttle)
- Comprehensive test coverage (~78%) with performance benchmarks

The library is designed to be imported and extended, with a reference CLI application demonstrating integration with LM Studio (OpenAI-compatible API).

---

## 2. Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.25+ |
| LLM API | OpenAI-compatible (LM Studio, OpenAI, etc.) |
| Testing | `testing` + `github.com/andygeiss/cloud-native-utils/assert` |
| Event Bus | `github.com/andygeiss/cloud-native-utils/messaging` |
| Resilience | `github.com/andygeiss/cloud-native-utils/stability` |
| Concurrency | `github.com/andygeiss/cloud-native-utils/efficiency` |
| Persistence | `github.com/andygeiss/cloud-native-utils/resource` |
| Security | `github.com/andygeiss/cloud-native-utils/security` |
| Logging | `log/slog` via `github.com/andygeiss/cloud-native-utils/logging` |
| Slice Utils | `github.com/andygeiss/cloud-native-utils/slices` |
| HTTP | Standard library `net/http` |
| Build | `go build` with PGO support |
| Container | Multi-stage Docker (scratch runtime) |
| Task Runner | [just](https://github.com/casey/just) |
| Linting | golangci-lint |

---

## 3. High-Level Architecture

The project follows **Hexagonal Architecture** (Ports and Adapters) combined with **Domain-Driven Design** principles.

```
┌──────────────────────────────────────────────────────────────────┐
│                    Application Layer (cmd/cli)                   │
│                   • CLI entry point, flags, I/O                  │
└───────────────────────────────┬──────────────────────────────────┘
                                │ uses
┌───────────────────────────────▼──────────────────────────────────┐
│                    Domain Layer (internal/domain)                │
│                  • Use cases (SendMessage, etc.)                 │
│                  • Orchestrates pkg/agent library                │
└───────────────────────────────┬──────────────────────────────────┘
                                │ depends on
┌───────────────────────────────▼──────────────────────────────────┐
│                  Core Library (pkg/agent)                        │
│   • Agent aggregate (conversation, tasks, iteration control)     │
│   • Task entity (lifecycle: Pending → Running → Completed)       │
│   • TaskService (agent loop orchestration)                       │
│   • Ports: LLMClient, ToolExecutor, EventPublisher               │
│   • Domain events: TaskStarted, TaskCompleted, ToolCallExecuted  │
└───────────────────────────────┬──────────────────────────────────┘
                                │ implemented by
┌───────────────────────────────▼──────────────────────────────────┐
│               Adapter Layer (internal/adapters/outbound)         │
│   • OpenAIClient (LLM port → OpenAI API)                         │
│   • ToolExecutor (tool registration and execution)               │
│   • EventPublisher (event port → messaging dispatcher)           │
└──────────────────────────────────────────────────────────────────┘
```

### Key Boundaries

| Layer | Responsibility | Dependencies |
|-------|----------------|--------------|
| `cmd/` | Application entry, configuration | Domain, pkg/agent |
| `internal/domain/` | Use cases, business rules | pkg/agent (via ports) |
| `internal/adapters/` | Infrastructure implementations | pkg/agent ports, external APIs |
| `pkg/agent/` | Core library, domain model | cloud-native-utils/event (interfaces only) |
| `pkg/openai/` | OpenAI API data structures | None |

---

## 4. Directory Structure (Contract)

```
go-agent/
├── cmd/
│   └── cli/                    # CLI application
│       ├── main.go             # Entry point, flag parsing, setup
│       └── main_test.go        # Integration tests
├── internal/
│   ├── adapters/
│   │   └── outbound/           # Infrastructure adapters
│   │       ├── conversation_store.go         # ConversationStore implementation
│   │       ├── encrypted_conversation_store.go # Encrypted storage wrapper
│   │       ├── event_publisher.go            # EventPublisher implementation
│   │       ├── openai_client.go              # LLMClient implementation
│   │       └── tool_executor.go              # ToolExecutor implementation
│   └── domain/
│       ├── chatting/           # Chatting domain
│       │   ├── service.go            # Use cases (SendMessage, ClearConversation, GetAgentStats)
│       │   ├── service_test.go       # Use case tests
│       │   └── value_objects.go      # Domain-specific types
│       └── tooling/            # Tooling domain
│           ├── aggregate.go          # Tool aggregate
│           ├── aggregate_test.go     # Tool aggregate tests
│           ├── entities.go           # Tool definitions
│           ├── service.go            # Tool implementations (Calculate, GetCurrentTime)
│           ├── service_test.go       # Tool tests
│           └── value_objects.go      # Tool-specific types
├── pkg/
│   ├── agent/                  # Reusable agent library (import this)
│   │   ├── aggregate.go        # Agent aggregate root
│   │   ├── benchmark_test.go   # Performance benchmarks
│   │   ├── entities.go         # LLMResponse, Message, Task, ToolCall, ToolDefinition
│   │   ├── errors.go           # Typed errors (LLMError, ToolError, TaskError)
│   │   ├── events.go           # Domain events (TaskStarted, TaskCompleted, ToolCallExecuted)
│   │   ├── ports_outbound.go   # LLMClient, ToolExecutor, EventPublisher, ConversationStore interfaces
│   │   ├── service.go          # Hooks, TaskService (agent loop orchestration)
│   │   └── value_objects.go    # ID types, Result, TokenUsage, Role/Status constants
│   └── openai/                 # OpenAI API data structures
│       ├── chat_completion_*.go # Request/response types
│       ├── message.go          # Chat message format
│       ├── tool.go             # Tool definition format
│       └── tool_call.go        # Tool call format
├── AGENTS.md                   # AI agent definitions index
├── CONTEXT.md                  # This file
├── Dockerfile                  # Multi-stage container build
├── README.md                   # User documentation
├── VENDOR.md                   # Vendor library documentation
├── docker-compose.yml          # Local development services
└── go.mod                      # Go module definition
```

### Rules for New Code

| Concern | Location |
|---------|----------|
| New domain use cases | `internal/domain/<domain>/` |
| New infrastructure adapters | `internal/adapters/outbound/` |
| Core agent library extensions | `pkg/agent/` |
| New domain events | `pkg/agent/events.go` |
| OpenAI API structures | `pkg/openai/` |
| Tests | Same directory as source, `*_test.go` |

---

## 5. Coding Conventions

### 5.1 General

- **Small, focused packages**: Each package has a single responsibility
- **Pure functions where possible**: Minimize side effects
- **Dependency injection**: Pass dependencies through constructors
- **Functional options pattern**: Use `With*` functions for configuration
- **Value objects for IDs**: `AgentID`, `TaskID`, `ToolCallID` provide type safety
- **Method chaining**: Builder pattern with `With*` methods returning the modified type
- **No global state**: All state is contained in structs

### 5.2 Naming

| Element | Convention | Example |
|---------|------------|---------|
| Files | `snake_case.go` | `task_service.go` |
| Test files | `<name>_test.go` | `task_service_test.go` |
| Packages | Single lowercase word | `agent`, `chatting`, `outbound` |
| Types/Structs | `PascalCase` | `TaskService`, `LLMResponse` |
| Interfaces | `PascalCase`, noun or `-er` suffix | `LLMClient`, `ToolExecutor` |
| Constructors | `New<Type>` | `NewAgent()`, `NewTaskService()` |
| Options | `With<Property>` | `WithMaxIterations()` |
| Constants | `PascalCase` for exported, grouped by type | `RoleSystem`, `TaskStatusPending` |
| ID types | `<Entity>ID` | `AgentID`, `TaskID` |
| Status types | `<Entity>Status` | `TaskStatus`, `ToolCallStatus` |

### 5.3 Error Handling & Logging

**Error Handling:**
- Use sentinel errors for common conditions: `ErrMaxIterationsReached`, `ErrToolNotFound`
- Use typed errors for rich context: `LLMError`, `ToolError`, `TaskError`
- Implement `Unwrap()` for `errors.Is`/`errors.As` support
- Wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
- Return errors up the call stack; let callers decide how to handle

**Logging:**
- No logging in library code (`pkg/agent/`)
- Adapters support optional structured logging via `WithLogger(*slog.Logger)`
- Use `logging.NewJsonLogger()` from cloud-native-utils for JSON output
- Log level controlled by `LOGGING_LEVEL` environment variable (DEBUG, INFO, WARN, ERROR)
- Use hooks for additional observability in applications

### 5.4 Testing

**Framework:** Standard `testing` package with `cloud-native-utils/assert`

**Naming Convention:**
```
Test_<Type>_<Method>_<Scenario>_Should_<Expected>
```

Examples:
- `Test_Agent_AddMessage_Should_AppendToHistory`
- `Test_TaskService_RunTask_With_MaxIterations_Should_ReturnError`

**Structure:** Arrange-Act-Assert pattern
```go
func Test_Example_Should_Work(t *testing.T) {
    // Arrange
    input := "test"
    
    // Act
    result := processInput(input)
    
    // Assert
    assert.That(t, "result must match expected", result, "expected")
}
```

**Test Location:** Same directory as source code, `*_test.go` suffix

**Mock Strategy:** 
- Create mock implementations of interfaces in test files
- Use interface boundaries for testability
- No external mocking libraries

### 5.5 Formatting & Linting

**Tools:**
- `gofmt` / `goimports`: Standard Go formatting
- `golangci-lint`: Comprehensive linting

**Key Rules:**
- Run `just fmt` before committing
- Run `just lint` to check for issues
- No lint warnings in CI

---

## 6. Cross-Cutting Concerns and Reusable Patterns

### Agent Loop Pattern

The core abstraction is the **Observe → Decide → Act → Update** loop in `TaskService`:

1. **Observe**: Build message context with system prompt
2. **Decide**: Call LLM with messages and tool definitions
3. **Act**: Execute any tool calls from LLM response
4. **Update**: Add messages to conversation, check for completion

### Ports (Interfaces)

Located in `pkg/agent/ports_outbound.go`:

| Port | Responsibility | Adapter |
|------|----------------|---------|
| `LLMClient` | Send messages to LLM, get responses | `OpenAIClient` |
| `ToolExecutor` | Register and execute tools | `ToolExecutor` |
| `EventPublisher` | Publish domain events | `EventPublisher` |
| `ConversationStore` | Persist and load conversation history | `ConversationStore`, `EncryptedConversationStore` |

### Hooks/Middleware

Located in `pkg/agent/service.go`. Available hooks:

| Hook | When Called | Use Case |
|------|-------------|----------|
| `BeforeTask` | Before task starts | Validation, setup, logging |
| `AfterTask` | After task completes | Cleanup, metrics |
| `BeforeLLMCall` | Before each LLM request | Rate limiting, request logging |
| `AfterLLMCall` | After each LLM response | Response caching, logging |
| `BeforeToolCall` | Before each tool execution | Authorization, argument validation |
| `AfterToolCall` | After each tool execution | Result logging, caching |

### Domain Events

Located in `pkg/agent/events.go`:

| Event | Topic | When Emitted |
|-------|-------|--------------|
| `EventTaskStarted` | `agent.task.started` | Task begins execution |
| `EventTaskCompleted` | `agent.task.completed` | Task finishes successfully |
| `EventTaskFailed` | `agent.task.failed` | Task terminates with error |
| `EventToolCallExecuted` | `agent.toolcall.executed` | Tool call completes |

### Tool Definition

Use typed parameters with `ParameterDefinition`:

```go
toolDef := agent.NewToolDefinition("search", "Search for information").
    WithParameterDef(agent.NewParameterDefinition("query", agent.ParamTypeString).
        WithDescription("The search query").
        WithRequired()).
    WithParameterDef(agent.NewParameterDefinition("limit", agent.ParamTypeInteger).
        WithDescription("Maximum results").
        WithDefault("10"))
```

### Typed Errors

Use `errors.Is` and `errors.As` for error handling:

```go
if errors.Is(err, agent.ErrMaxIterationsReached) { ... }

var toolErr *agent.ToolError
if errors.As(err, &toolErr) {
    log.Printf("Tool %s failed: %s", toolErr.ToolName, toolErr.Message)
}
```

### Vendor Libraries

See [VENDOR.md](VENDOR.md) for approved vendor libraries and usage patterns. Key guidance:

- Use `cloud-native-utils/assert` for testing assertions
- Use `cloud-native-utils/messaging` for event publishing
- Use `cloud-native-utils/stability` for resilience patterns (Timeout, Retry, Breaker, Debounce)
- Use `cloud-native-utils/efficiency` for parallel processing (Generate, Process)
- Use `cloud-native-utils/resource` for persistent storage (Access interface)
- Use `cloud-native-utils/security` for encryption (AES-GCM)
- Use `cloud-native-utils/slices` for Filter/Map/Unique operations
- Prefer Go standard library for HTTP, JSON, context

### Resilience Patterns

External calls (LLM, tools) are wrapped with resilience patterns from `cloud-native-utils/stability`:

| Pattern | Purpose | Default |
|---------|---------|---------|
| `stability.Timeout` | Enforce maximum execution time | LLM: 120s, Tools: 30s |
| `stability.Retry` | Handle transient failures | 3 attempts, 2s delay |
| `stability.Breaker` | Prevent cascading failures | Opens after 5 failures |
| `stability.Throttle` | Rate limit API calls | Disabled (opt-in) |

**LLM Call Flow (with throttling enabled):**
```
Request → Timeout(120s) → Retry(3, 2s) → Breaker(5) → Throttle(tokens) → HTTP POST
```

**Tool Execution Flow:**
```
Execute → Timeout(30s) → Tool Function
```

All settings are configurable via builder methods:
```go
client := outbound.NewOpenAIClient(baseURL, model).
    WithLLMTimeout(90 * time.Second).
    WithRetry(5, 3*time.Second).
    WithCircuitBreaker(10).
    WithThrottle(10, 2, time.Second)  // 10 calls max, refill 2/sec
```

### Parallel Tool Execution

When the LLM returns multiple tool calls, they can be executed concurrently using `cloud-native-utils/efficiency`:

```go
taskService := agent.NewTaskService(llmClient, toolExecutor, publisher).
    WithParallelToolExecution()
```

**Sequential (default):**
```
Tool1 → Tool2 → Tool3 (total time = sum of all)
```

**Parallel:**
```
Tool1 ─┐
Tool2 ─┼─→ Results (total time = max of all)
Tool3 ─┘
```

Parallel execution is ideal for I/O-bound tools (API calls, file operations). For CPU-bound operations, sequential may be more efficient due to coordination overhead.

### Conversation Persistence

Persist conversation history using `cloud-native-utils/resource`:

```go
// In-memory storage (for testing)
store := outbound.NewInMemoryConversationStore()

// JSON file storage (for production)
store := outbound.NewJsonFileConversationStore("conversations.json")

// ConversationStore interface
type ConversationStore interface {
    Save(ctx context.Context, agentID AgentID, messages []Message) error
    Load(ctx context.Context, agentID AgentID) ([]Message, error)
    Clear(ctx context.Context, agentID AgentID) error
}
```

### Encrypted Storage

Protect sensitive conversation data with AES-GCM encryption from `cloud-native-utils/security`:

```go
// Generate encryption key (store securely!)
key := security.GenerateKey()

// Wrap any ConversationStore with encryption
baseStore := outbound.NewJsonFileConversationStore("conversations.json")
encStore := outbound.NewEncryptedConversationStore(baseStore, key)

// Use like any ConversationStore
err := encStore.Save(ctx, agentID, messages)
messages, err := encStore.Load(ctx, agentID)
```

---

## 7. Using This Repo as a Template

### What Must Be Preserved

- **Directory structure**: `cmd/`, `internal/`, `pkg/` layout
- **Port/adapter pattern**: Infrastructure behind interfaces
- **Agent library API**: `NewAgent()`, `NewTaskService()`, hooks pattern
- **Testing conventions**: Naming, Arrange-Act-Assert, assert library
- **Error handling**: Typed errors with `Unwrap()` support

### What Can Be Customized

| Customization | Location |
|---------------|----------|
| New tools | `internal/adapters/outbound/tool_executor.go` |
| New LLM providers | New adapter implementing `LLMClient` |
| New use cases | `internal/domain/<domain>/` |
| Custom hooks | Application code using `WithHooks()` |
| System prompts | Application configuration |
| Agent options | `WithMaxIterations()`, `WithMaxMessages()`, etc. |

### Steps to Create a New Project

1. **Clone/copy this repository**
2. **Update module path** in `go.mod`
3. **Update package imports** throughout
4. **Customize CLI** in `cmd/cli/main.go` or create new entry points
5. **Add domain-specific tools** in `tool_executor.go`
6. **Add domain-specific use cases** in `internal/domain/`
7. **Update README.md** with project-specific documentation

---

## 8. Key Commands & Workflows

| Command | Description |
|---------|-------------|
| `just bench` | Run performance benchmarks |
| `just build` | Build Docker image |
| `just down` | Stop services |
| `just fmt` | Format Go code |
| `just lint` | Run linter checks |
| `just run` | Run CLI application locally |
| `just setup` | Install development dependencies |
| `just test` | Run unit tests with coverage |
| `just test-integration` | Run integration tests (requires LM Studio) |
| `just up` | Start services with docker-compose |

### Environment Configuration

Copy `.env.example` to `.env` and configure:

```bash
LM_STUDIO_URL=http://localhost:1234
LM_STUDIO_MODEL=your-model-name
```

### CLI Options

```bash
go run ./cmd/cli \
    -url http://localhost:1234 \
    -model <model-name> \
    -max-iterations 10 \
    -max-messages 50 \
    -verbose
```

---

## 9. Important Notes & Constraints

### Security

- No secrets in code; use environment variables
- The library does not handle authentication; adapters should implement as needed
- Tool execution is sandboxed to registered functions only

### Performance

- `MaxMessages` prevents unbounded memory growth
- `MaxIterations` prevents infinite loops
- Profile-Guided Optimization (PGO) supported in Docker builds

### Platform

- Requires Go 1.25+
- Docker builds use `scratch` base (minimal image)
- Designed for OpenAI-compatible APIs (LM Studio, OpenAI, Ollama with adapter)

### Limitations

- Single-agent design (no multi-agent orchestration)
- Synchronous execution (no async tool calls)
- Demo tools included for reference (production use requires custom tools)

---

## 10. How AI Tools and RAG Should Use This File

### Priority

This file is the **top-priority context** for repository-wide work. Read it before:
- Making architectural changes
- Adding new packages or modules
- Refactoring existing code
- Understanding project conventions

### Usage Guidelines

1. **Always read `CONTEXT.md` first** before major changes
2. **Treat rules as constraints** unless explicitly updated
3. **Reference this file** when documenting architectural decisions
4. **Cross-reference with**:
   - [README.md](README.md) for user-facing documentation
   - [VENDOR.md](VENDOR.md) for approved vendor libraries
   - [AGENTS.md](AGENTS.md) for AI agent definitions

### Agent Collaboration

When making changes:
1. `coding-assistant` implements code changes
2. `CONTEXT-maintainer` updates this file if architecture changes
3. `README-maintainer` updates README.md if user-facing docs change
4. `VENDOR-maintainer` updates VENDOR.md if dependencies change

See [AGENTS.md](AGENTS.md) for full agent definitions and collaboration patterns.
