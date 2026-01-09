# CONTEXT.md

## 1. Project purpose

**go-agent** is a reusable AI agent framework for Go implementing the observe → decide → act → update loop pattern for LLM-based task execution. It provides:

- A clean, domain-driven architecture for building AI agents with tool use capabilities
- OpenAI-compatible API integration (works with LM Studio and other local LLMs)
- Built-in resilience patterns (retry, circuit breaker, timeout, throttling)
- Event-driven architecture for task lifecycle observability
- Memory system for long-term agent context

This repository serves as both a **production-ready library** and a **reference implementation** for building agentic applications in Go using hexagonal architecture patterns.

---

## 2. Technology stack

| Category | Technology |
|----------|------------|
| Language | Go 1.25+ |
| LLM API | OpenAI-compatible (LM Studio, OpenAI, etc.) |
| Architecture | Hexagonal (Ports & Adapters) / DDD |
| Utility Library | `github.com/andygeiss/cloud-native-utils` |
| Build | Go modules, multi-stage Docker |
| Profiling | PGO (Profile-Guided Optimization) |
| Container Runtime | Docker / Docker Compose |

### Key dependencies

- `cloud-native-utils/messaging` — Event dispatching
- `cloud-native-utils/stability` — Timeout, retry, circuit breaker patterns
- `cloud-native-utils/efficiency` — Worker pools for parallel execution
- `cloud-native-utils/resource` — Generic storage access (in-memory, JSON file)
- `cloud-native-utils/slices` — Functional slice utilities

---

## 3. High-level architecture

The project follows **hexagonal architecture** (ports and adapters) with **domain-driven design** principles:

```
┌─────────────────────────────────────────────────────────────────┐
│                         cmd/cli                                 │
│                    (Application Entry)                          │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                     internal/domain                             │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ agent/        Core agent aggregate, task service, types     ││
│  │ chatting/     Use cases: SendMessage, ClearConversation     ││
│  │ memorizing/   Use cases: WriteNote, SearchNotes, GetNote    ││
│  │ tooling/      Tool implementations (calculate, time, etc.)  ││
│  │ openai/       OpenAI API data structures (value objects)    ││
│  └─────────────────────────────────────────────────────────────┘│
└──────────────────────────────┬──────────────────────────────────┘
                               │ depends on interfaces (ports)
┌──────────────────────────────▼──────────────────────────────────┐
│                   internal/adapters/outbound                    │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ openai_client.go       LLMClient implementation             ││
│  │ tool_executor.go       ToolExecutor implementation          ││
│  │ event_publisher.go     EventPublisher implementation        ││
│  │ memory_store.go        MemoryStore implementation           ││
│  │ conversation_store.go  ConversationStore implementation     ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

### Agent loop pattern

The core agent implements the observe → decide → act → update loop:

1. **Observe**: Receive user input, build message context
2. **Decide**: Call LLM with messages and tool definitions
3. **Act**: Execute any tool calls requested by LLM
4. **Update**: Add results to conversation, check termination conditions
5. **Repeat** until task completes, fails, or max iterations reached

---

## 4. Directory structure (contract)

```
go-agent/
├── cmd/
│   └── cli/                    # CLI application entry point
│       ├── main.go             # Main function, flag parsing, wiring
│       └── main_test.go        # Integration tests
├── internal/
│   ├── adapters/
│   │   └── outbound/           # Infrastructure adapters (ports implementations)
│   │       ├── openai_client.go        # LLMClient → OpenAI API
│   │       ├── tool_executor.go        # ToolExecutor → tool registry
│   │       ├── event_publisher.go      # EventPublisher → messaging
│   │       ├── memory_store.go         # MemoryStore → storage backend
│   │       ├── conversation_store.go   # ConversationStore → file persistence
│   │       └── encrypted_*.go          # Encrypted variants
│   └── domain/
│       ├── agent/              # Core domain: Agent aggregate, Task, Message, etc.
│       │   ├── agent.go        # Agent aggregate root
│       │   ├── task.go         # Task entity
│       │   ├── task_service.go # Task orchestration (agent loop)
│       │   ├── message.go      # Message value object
│       │   ├── tool_*.go       # Tool-related types
│       │   ├── memory*.go      # Memory types and interfaces
│       │   ├── hooks.go        # Lifecycle hooks
│       │   ├── events.go       # Domain events
│       │   ├── errors.go       # Domain errors
│       │   ├── llm.go          # LLMClient interface (port)
│       │   ├── tools.go        # ToolExecutor interface (port)
│       │   └── shared.go       # Shared types (IDs, Result, Role, Status)
│       ├── chatting/           # Chatting use cases
│       │   ├── send_message.go
│       │   ├── clear_conversation.go
│       │   └── get_agent_stats.go
│       ├── memorizing/         # Memory management use cases
│       │   ├── service.go      # Memory service
│       │   ├── write_note.go
│       │   ├── search_notes.go
│       │   ├── get_note.go
│       │   └── delete_note.go
│       ├── tooling/            # Tool implementations
│       │   ├── calculate.go    # Arithmetic calculator
│       │   ├── time.go         # Current time
│       │   └── memory_tools.go # Memory read/write tools
│       └── openai/             # OpenAI API types (value objects)
│           └── *.go            # Request/response structures
├── AGENTS.md                   # Agent definitions index
├── CONTEXT.md                  # This file
├── README.md                   # User-facing documentation
├── go.mod / go.sum             # Go modules
├── Dockerfile                  # Multi-stage build
└── docker-compose.yml          # Container orchestration
```

### Rules for new code

| Code type | Location |
|-----------|----------|
| New domain entities/aggregates | `internal/domain/agent/` |
| New use cases | `internal/domain/<bounded-context>/` |
| New tool implementations | `internal/domain/tooling/` |
| New infrastructure adapters | `internal/adapters/outbound/` |
| CLI commands/flags | `cmd/cli/` |
| Tests | Same directory as implementation (`*_test.go`) |
| OpenAI API types | `internal/domain/openai/` |

---

## 5. Coding conventions

### 5.1 General

- **Small, focused files**: One type or concept per file
- **Interface segregation**: Define interfaces in the domain, implement in adapters
- **Functional options**: Use `With*` pattern for optional configuration
- **Value objects with builders**: Use method chaining for building immutable types
- **Pure domain logic**: Domain layer has no external dependencies
- **Dependency injection**: Wire dependencies at the application boundary (`cmd/`)

### 5.2 Naming

| Element | Convention | Example |
|---------|------------|---------|
| Files | `snake_case.go` | `tool_executor.go` |
| Packages | lowercase, short | `agent`, `chatting` |
| Types | `PascalCase` | `TaskService`, `ToolCall` |
| Interfaces | `PascalCase`, describe behavior | `LLMClient`, `ToolExecutor` |
| ID types | `*ID` suffix | `TaskID`, `AgentID`, `NoteID` |
| Status enums | `*Status` suffix with constants | `TaskStatus`, `ToolCallStatus` |
| Options | `Option` type with `With*` functions | `WithMaxIterations(10)` |
| Constructors | `New*` | `NewAgent()`, `NewTask()` |
| Builder methods | `With*` | `WithToolCalls()`, `WithDuration()` |
| Use cases | `*UseCase` | `SendMessageUseCase` |
| Events | `Event*` | `EventTaskStarted` |
| Errors | `Err*` sentinel, `*Error` struct | `ErrToolNotFound`, `ToolError` |

### 5.3 Error handling & logging

**Error patterns:**
- Define sentinel errors with `errors.New()` for expected conditions
- Create typed error structs (`LLMError`, `TaskError`, `ToolError`) with `Unwrap()` for error chains
- Return errors up the call stack; handle at appropriate boundaries
- Use `fmt.Errorf("context: %w", err)` to wrap errors with context

**Logging:**
- Use `log/slog` for structured logging
- Inject logger via `With*` methods (optional)
- Log at debug level for normal operations, error level for failures
- Include relevant context: tool name, duration, task ID

### 5.4 Testing

**Framework:** Standard `testing` package

**Organization:**
- Test files: `*_test.go` in same package
- Unit tests: Test single functions/methods in isolation
- Table-driven tests: Use `tests := []struct{...}` pattern
- Benchmarks: `Benchmark*` functions in `*_test.go`

**Patterns:**
- Create test helpers in `shared_test.go`
- Use in-memory implementations for testing adapters
- Test error conditions explicitly

### 5.5 Formatting & linting

- **Formatter:** `gofmt` / `goimports`
- **Linter:** `go vet`, standard Go toolchain
- Run `go fmt ./...` before committing
- Keep imports grouped: stdlib, external, internal

---

## 6. Cross-cutting concerns and reusable patterns

### Resilience (via cloud-native-utils/stability)

```go
// Timeout wrapper
stability.Timeout(fn, 30*time.Second)

// Retry with backoff
stability.Retry(fn, 3, 2*time.Second)

// Circuit breaker
stability.CircuitBreaker(fn, threshold)
```

The `OpenAIClient` wraps LLM calls with configurable resilience:
- HTTP timeout: 60s (configurable)
- LLM call timeout: 120s (configurable)
- Retry: 3 attempts with 2s delay (configurable)
- Circuit breaker: opens after 5 failures (configurable)
- Throttle: disabled by default (configurable)

### Event publishing

Domain events are published via `EventPublisher` interface:
- `agent.task.started` — Task begins execution
- `agent.task.completed` — Task finishes successfully
- `agent.task.failed` — Task terminates with error
- `agent.toolcall.executed` — Tool call completes

### Hooks for extensibility

```go
hooks := agent.NewHooks().
    WithBeforeTask(func(ctx, agent, task) error { ... }).
    WithAfterToolCall(func(ctx, agent, toolCall) error { ... })

taskService.WithHooks(hooks)
```

### Tool registration

```go
tool := tooling.NewCalculateTool()
executor.RegisterTool("calculate", tool.Func)
executor.RegisterToolDefinition(tool.Definition)
```

### Memory system

Memory notes store long-term context:
- `MemoryNote` — atomic unit with metadata, tags, importance
- `MemoryStore` — interface with in-memory and JSON file implementations
- Search by query, filter by user/session/task/tags

---

## 7. Using this repo as a template

### Invariants (must preserve)

- Hexagonal architecture separation (domain → adapters → cmd)
- Interface definitions in domain layer
- Functional options pattern for configuration
- Event-driven task lifecycle
- Structured error types

### Customization points

| Customization | Location | How |
|---------------|----------|-----|
| Add new tools | `internal/domain/tooling/` | Implement `agent.Tool` with `ToolFunc` and `ToolDefinition` |
| Add use cases | `internal/domain/<context>/` | Create `*UseCase` struct with `Execute()` method |
| Custom LLM backend | `internal/adapters/outbound/` | Implement `agent.LLMClient` interface |
| Custom storage | `internal/adapters/outbound/` | Implement `agent.MemoryStore` interface |
| Custom events | `internal/domain/agent/events.go` | Add event types implementing `event.Event` |

### Steps to create a new project from this template

1. **Clone/copy** the repository
2. **Update module path** in `go.mod`
3. **Update metadata**: README.md, LICENSE, docker-compose.yml
4. **Remove example tools** in `internal/domain/tooling/` (or keep as reference)
5. **Add domain-specific tools** following the established pattern
6. **Extend use cases** in `internal/domain/chatting/` or create new bounded contexts
7. **Configure adapters** for your infrastructure (LLM endpoint, storage backend)

---

## 8. Key commands & workflows

### Development

```bash
# Run CLI with LM Studio
go run ./cmd/cli -url http://localhost:1234 -model <model-name>

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./internal/domain/agent/

# Format code
go fmt ./...

# Vet code
go vet ./...
```

### Build

```bash
# Build binary
go build -o go-agent ./cmd/cli

# Build with optimizations (production)
go build -ldflags "-s -w" -o go-agent ./cmd/cli

# Build with PGO (requires cpuprofile.pprof)
go build -ldflags "-s -w" -pgo cpuprofile.pprof -o go-agent ./cmd/cli
```

### Docker

```bash
# Build image
docker build -t go-agent .

# Run with docker-compose
docker-compose up -d

# View logs
docker-compose logs -f app
```

### CLI flags

| Flag | Default | Description |
|------|---------|-------------|
| `-url` | `http://localhost:1234` | LM Studio API base URL |
| `-model` | `$LM_STUDIO_MODEL` | Model name |
| `-max-iterations` | `10` | Max iterations per task |
| `-max-messages` | `50` | Max messages to retain (0 = unlimited) |
| `-verbose` | `false` | Show detailed metrics |

---

## 9. Important notes & constraints

### Security

- **No secrets in code**: Use environment variables for API keys
- Tool execution has timeout protection (default 30s)
- Consider encrypted conversation store for sensitive data

### Performance

- LLM calls are the bottleneck; tune timeouts appropriately
- Use `WithParallelToolExecution()` for I/O-bound tool calls
- Message history trimming prevents unbounded memory growth

### Platform assumptions

- Go 1.25+ (uses latest language features)
- OpenAI-compatible LLM API (LM Studio, OpenAI, vLLM, etc.)
- Docker for containerized deployment

### Known limitations

- Basic text search for memory (no embedding/vector similarity)
- Single-agent design (no multi-agent orchestration)
- Synchronous execution model

---

## 10. How AI tools and RAG should use this file

### Priority order for context

1. **CONTEXT.md** (this file) — Architecture, conventions, contracts
2. **README.md** — Project purpose, quick start, user-facing docs
3. **VENDOR.md** — External library usage patterns (if exists)

### Instructions for AI agents

- **Read CONTEXT.md first** before making architectural changes
- **Follow the directory structure contract** when adding new code
- **Use established patterns**: functional options, interface segregation, typed errors
- **Check existing implementations** before creating new tools or adapters
- **Prefer cloud-native-utils** patterns for resilience and efficiency
- **Update CONTEXT.md** after significant architectural changes

### When to reference this file

- Creating new domain entities or aggregates
- Adding new use cases or bounded contexts
- Implementing new adapters
- Understanding error handling patterns
- Reviewing code for convention compliance
