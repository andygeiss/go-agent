# CONTEXT.md

## 1. Project purpose

**go-agent** is a reusable AI agent framework for Go implementing the observe → decide → act → update loop pattern for LLM-based task execution. It provides:

- A clean, domain-driven architecture for building AI agents with tool use capabilities
- OpenAI-compatible API integration (works with LM Studio and other local LLMs)
- Built-in resilience patterns (debounce, retry, circuit breaker, throttle, timeout)
- Event-driven architecture for task lifecycle observability
- Memory system for long-term agent context with search and filtering

This repository serves as both a **production-ready library** and a **reference implementation** for building agentic applications in Go using hexagonal architecture patterns.

---

## 2. Technology stack

| Category | Technology |
|----------|------------|
| Architecture | Hexagonal (Ports & Adapters) / DDD |
| Build | Go modules, multi-stage Docker |
| Container Runtime | Docker / Docker Compose |
| Language | Go 1.25.5+ |
| LLM API | OpenAI-compatible (LM Studio, Ollama, OpenAI, vLLM, etc.) |
| Profiling | PGO (Profile-Guided Optimization) |
| Utility Library | `github.com/andygeiss/cloud-native-utils` v0.4.12 |

### Key dependencies (alphabetically sorted)

- `cloud-native-utils/efficiency` — Worker pools for parallel tool execution
- `cloud-native-utils/event` — Event interface for domain events
- `cloud-native-utils/messaging` — Event dispatching
- `cloud-native-utils/resource` — Generic storage access (in-memory, JSON file, YAML file)
- `cloud-native-utils/security` — AES-GCM encryption for data at rest
- `cloud-native-utils/service` — Generic function type for stability wrappers
- `cloud-native-utils/slices` — Functional slice utilities (filter, map, contains)
- `cloud-native-utils/stability` — Breaker, debounce, retry, throttle, timeout patterns

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

1. **Observe**: Receive user input, build message context with system prompt
2. **Decide**: Call LLM with messages and tool definitions
3. **Act**: Execute any tool calls requested by LLM (sequential or parallel)
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
│   │   ├── inbound/            # Inbound adapters (data sources)
│   │   │   ├── file_walker.go              # FileWalker → filesystem traversal
│   │   │   └── file_walker_test.go         # Tests
│   │   └── outbound/           # Outbound adapters (ports implementations)
│   │       ├── conversation_store.go       # ConversationStore → resource.Access
│   │       ├── encrypted_conversation_store.go # Encrypted variant with AES-GCM
│   │       ├── event_publisher.go          # EventPublisher → messaging.Dispatcher
│   │       ├── index_store.go              # IndexStore → resource.Access
│   │       ├── memory_store.go             # MemoryStore → resource.Access
│   │       ├── openai_client.go            # LLMClient → OpenAI-compatible API
│   │       └── tool_executor.go            # ToolExecutor → tool registry
│   └── domain/
│       ├── agent/              # Core domain: Agent aggregate, Task, Message, etc.
│       │   ├── agent.go        # Agent aggregate root + Metadata + Options
│       │   ├── errors.go       # Domain errors (LLMError, TaskError, ToolError)
│       │   ├── events.go       # Domain events (EventTask*, EventToolCall*)
│       │   ├── memory_note.go  # MemoryNote entity with builder pattern
│       │   ├── message.go      # Message + LLMResponse + ToolCall
│       │   ├── ports.go        # All interfaces (ConversationStore, EventPublisher, LLMClient, MemoryStore, TaskRunner, ToolExecutor)
│       │   ├── service.go      # TaskService + Hooks
│       │   ├── shared.go       # ID types, Result, Role, Status, TokenUsage, Tool
│       │   ├── task.go         # Task entity with lifecycle methods
│       │   └── tool_definition.go # ToolDefinition + ParameterDefinition + validation
│       ├── chatting/           # Chatting use cases
│       │   └── service.go      # AgentStats + ClearConversationUseCase + GetAgentStatsUseCase + SendMessageUseCase
│       ├── indexing/           # File system indexing bounded context
│       │   ├── ports.go        # FileWalker + IndexStore interfaces
│       │   ├── service.go      # Service: Scan, ChangedSince, DiffSnapshots
│       │   └── snapshot.go     # FileInfo + Snapshot + DiffResult + HashFile
│       ├── memorizing/         # Memory management use cases
│       │   ├── errors.go       # Sentinel errors (ErrNoteIDEmpty, ErrNoteNil)
│       │   └── service.go      # DeleteNoteUseCase + GetNoteUseCase + SearchNotesUseCase + Service + WriteNoteUseCase
│       ├── openai/             # OpenAI API types
│       │   ├── openai.go       # Package doc
│       │   ├── request.go      # ChatCompletionRequest + Message
│       │   ├── response.go     # ChatCompletionResponse + ChatCompletionChoice + ChatCompletionUsage
│       │   └── tool.go         # FunctionCall + FunctionDefinition + Tool + ToolCall
│       └── tooling/            # Tool implementations
│           ├── calculate.go    # Arithmetic calculator with expression parser
│           ├── index_tools.go  # IndexToolService (IndexScan, IndexChangedSince, IndexDiffSnapshot)
│           ├── memory_tools.go # MemoryToolService (MemoryGet, MemorySearch, MemoryWrite)
│           └── time.go         # Current time (RFC3339 format)
├── AGENTS.md                   # Agent definitions index
├── CONTEXT.md                  # This file (architecture documentation)
├── Dockerfile                  # Multi-stage build
├── docker-compose.yml          # Container orchestration
├── go.mod / go.sum             # Go modules
├── README.md                   # User-facing documentation
└── VENDOR.md                   # Vendor library documentation
```

### Rules for new code

| Code type | Location |
|-----------|----------|
| CLI commands/flags | `cmd/cli/` |
| Inbound adapters (data sources) | `internal/adapters/inbound/` |
| New domain entities/aggregates | `internal/domain/agent/` |
| New tool implementations | `internal/domain/tooling/` |
| New use cases | `internal/domain/<bounded-context>/service.go` |
| OpenAI API types | `internal/domain/openai/` |
| Outbound adapters (infrastructure) | `internal/adapters/outbound/` |
| Tests | Same directory as implementation (`*_test.go`) |

---

## 5. Coding conventions

### 5.1 General

- **Functionality-based files**: Related types live together (Go stdlib idioms)
- **Interface segregation**: Define interfaces in `ports.go`, implement in adapters
- **Functional options**: Use `With*` pattern for optional configuration
- **Value objects with builders**: Use method chaining for building immutable types
- **Pure domain logic**: Domain layer has no external dependencies
- **Dependency injection**: Wire dependencies at the application boundary (`cmd/`)

### 5.2 Naming

| Element | Convention | Example |
|---------|------------|---------|
| Builder methods | `With*` | `WithDuration()`, `WithToolCalls()` |
| Constructors | `New*` | `NewAgent()`, `NewTask()` |
| Errors | `Err*` sentinel, `*Error` struct | `ErrToolNotFound`, `ToolError` |
| Events | `Event*` | `EventTaskCompleted`, `EventTaskStarted` |
| Files | `snake_case.go` | `memory_note.go`, `tool_executor.go` |
| ID types | `*ID` suffix | `AgentID`, `NoteID`, `TaskID`, `ToolCallID` |
| Interfaces | `PascalCase`, describe behavior | `EventPublisher`, `LLMClient`, `MemoryStore`, `ToolExecutor` |
| Options | `Option` type with `With*` functions | `WithMaxIterations(10)`, `WithMaxMessages(50)` |
| Packages | lowercase, short | `agent`, `chatting`, `memorizing`, `tooling` |
| Status enums | `*Status` suffix with constants | `TaskStatus`, `ToolCallStatus` |
| Types | `PascalCase` | `MemoryNote`, `TaskService`, `ToolCall` |
| Use cases | `*UseCase` | `ClearConversationUseCase`, `SendMessageUseCase` |

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
- Benchmarks: `Benchmark*` functions in `*_test.go` (see `cmd/cli/main_test.go` for PGO benchmarks)

**Benchmark categories** (in `cmd/cli/main_test.go`):
- **FSWalker Benchmarks** — Real file system walking with/without ignore patterns
- **Full Stack Benchmarks** — End-to-end use case execution with mock LLM
- **Index Store Benchmarks** — Snapshot save/get operations
- **Index Tool Service Benchmarks** — Tool-based indexing operations
- **Indexing Service Benchmarks** — Scan, ChangedSince, DiffSnapshots at 100/1000/10000 files
- **Memory Store Benchmarks** — Raw adapter layer operations at 100, 1000, 10000 notes
- **Memory Tools Benchmarks** — Tool-based memory operations with JSON parsing
- **Memory Use Case Benchmarks** — Domain layer use cases (Write, Search, Get, Delete)
- **MemoryNote Object Benchmarks** — Object creation and method performance
- **Message Handling Benchmarks** — Message creation and trimming
- **Snapshot Benchmarks** — Snapshot object creation and method performance
- **Task Service Benchmarks** — Task execution with various tool patterns
- **Tool Execution Benchmarks** — Real tool execution (calculate, time)

**Patterns:**
- Create test helpers in `shared_test.go`
- Use in-memory implementations for testing adapters
- Test error conditions explicitly
- Use `b.Loop()` for benchmarks (Go 1.24+)
- Pre-populate stores before benchmarks with `b.ResetTimer()`

### 5.5 Formatting & linting

- **Formatter:** `gofmt` / `goimports`
- **Linter:** `go vet`, standard Go toolchain
- Run `go fmt ./...` before committing
- Keep imports grouped: stdlib, external, internal

### 5.6 File organization (Go stdlib idioms)

Structure Go files by **functionality**, not by DDD element type:

| File | Contents | Rationale |
|------|----------|----------|
| `<aggregate>.go` | Aggregate root + related entities + value objects | Self-contained concept |
| `errors.go` | Sentinel errors + typed error structs | Domain-specific error conditions |
| `events.go` | Event types only (no interfaces) | Events are published payloads |
| `message.go` | Message + LLMResponse + ToolCall | Related types grouped together |
| `ports.go` | All inbound + outbound interfaces | Clear "API surface" for adapters |
| `request.go` | API request types | Outbound API data structures |
| `response.go` | API response types | Inbound API data structures |
| `service.go` | Domain service + hooks + use cases | Service orchestrates, use cases consolidated |
| `tool.go` | Tool-related types (definitions, calls) | Tool abstraction grouped |

**Guiding principles:**
1. **Discoverability** — `ports.go` is the entry point for adapter implementations
2. **Go idioms** — Match stdlib organization (e.g., `net/http` keeps related types together)
3. **Locality** — Open one file to understand a complete concept

**Anti-patterns to avoid:**
- One type per file (excessive fragmentation)
- Separate files for entities, events, value objects (DDD theater)

---

## 6. Cross-cutting concerns and reusable patterns

### Event publishing

Domain events are published via `EventPublisher` interface (alphabetically sorted):
- `agent.task.completed` — Task finishes successfully
- `agent.task.failed` — Task terminates with error
- `agent.task.started` — Task begins execution
- `agent.toolcall.executed` — Tool call completes

### Hooks for extensibility

```go
hooks := agent.NewHooks().
    WithBeforeTask(func(ctx context.Context, agent *Agent, task *Task) error { ... }).
    WithAfterToolCall(func(ctx context.Context, agent *Agent, toolCall *ToolCall) error { ... })

taskService.WithHooks(hooks)
```

Available hooks (alphabetically sorted):
- `AfterLLMCall` — Called after each LLM response is received
- `AfterTask` — Called after a task completes (success or failure)
- `AfterToolCall` — Called after each tool execution completes
- `BeforeLLMCall` — Called before each LLM request
- `BeforeTask` — Called before a task starts executing
- `BeforeToolCall` — Called before each tool execution

### Memory system

Memory notes store long-term context:
- `MemoryNote` — Atomic unit with metadata, tags, keywords, importance (1-5 scale)
- `MemoryStore` — Interface with in-memory and JSON file implementations
- `MemorySearchOptions` — Filter by SessionID, TaskID, UserID, or Tags
- `SourceType` — Categorizes note origin (fact, plan_step, preference, summary, tool_result, user_message)

### Resilience (via cloud-native-utils/stability)

```go
// Timeout wrapper
stability.Timeout(fn, 30*time.Second)

// Retry with backoff
stability.Retry(fn, 3, 2*time.Second)

// Circuit breaker
stability.Breaker(fn, threshold)

// Throttle (token bucket)
stability.Throttle(fn, maxTokens, refill, period)

// Debounce (coalesce rapid calls)
stability.Debounce(fn, period)
```

The `OpenAIClient` wraps LLM calls with configurable resilience:
- Circuit breaker: opens after 5 failures (configurable via `WithCircuitBreaker`)
- Debounce: disabled by default (configurable via `WithDebounce`)
- HTTP timeout: 60s (configurable via `WithHTTPClient`)
- LLM call timeout: 120s (configurable via `WithLLMTimeout`)
- Retry: 3 attempts with 2s delay (configurable via `WithRetry`)
- Throttle: disabled by default (configurable via `WithThrottle`)

### Tool registration

```go
tool := tooling.NewCalculateTool()
executor.RegisterTool("calculate", tool.Func)
executor.RegisterToolDefinition(tool.Definition)
```

Built-in tools (alphabetically sorted):
- `calculate` — Safe arithmetic expression evaluator with operator precedence
- `get_current_time` — Returns current date/time in RFC3339 format
- `index.changed_since` — Find files modified after a timestamp
- `index.diff_snapshot` — Compare two snapshots to find added/changed/removed files
- `index.scan` — Scan directories and create a file system snapshot
- `memory_get` — Retrieve a specific note by ID
- `memory_search` — Search notes with query and filters
- `memory_write` — Store a new memory note

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
| Custom events | `internal/domain/agent/events.go` | Add event types implementing `event.Event` |
| Custom LLM backend | `internal/adapters/outbound/` | Implement `agent.LLMClient` interface |
| Custom storage | `internal/adapters/outbound/` | Implement `agent.ConversationStore` or `agent.MemoryStore` interface |

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

### Build

```bash
# Build binary
go build -o go-agent ./cmd/cli

# Build with optimizations (production)
go build -ldflags "-s -w" -o go-agent ./cmd/cli

# Build with PGO (requires cpuprofile.pprof)
go build -ldflags "-s -w" -pgo cpuprofile.pprof -o go-agent ./cmd/cli
```

### CLI flags

| Flag | Default | Description |
|------|---------|-------------|
| `-index-file` | `""` | JSON file for persistent indexing (empty = in-memory) |
| `-max-iterations` | `10` | Max iterations per task |
| `-max-messages` | `50` | Max messages to retain (0 = unlimited) |
| `-memory-file` | `""` | JSON file for persistent memory (empty = in-memory) |
| `-model` | `$LM_STUDIO_MODEL` | Model name |
| `-parallel-tools` | `false` | Execute tools in parallel |
| `-url` | `http://localhost:1234` | LM Studio API base URL |
| `-verbose` | `false` | Show detailed metrics |

### Development

```bash
# Run CLI with LM Studio
go run ./cmd/cli -url http://localhost:1234 -model <model-name>

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run all benchmarks (for PGO profiling)
go test -bench=. ./cmd/cli/...

# Run memory benchmarks only
go test -bench=Memory ./cmd/cli/...

# Format code
go fmt ./...

# Vet code
go vet ./...
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

---

## 9. Important notes & constraints

### Known limitations

- Basic text search for memory (no embedding/vector similarity)
- Single-agent design (no multi-agent orchestration)
- Synchronous execution model (async support via goroutines)

### Performance

- LLM calls are the bottleneck; tune timeouts appropriately
- Message history trimming prevents unbounded memory growth
- Use `WithParallelToolExecution()` for I/O-bound tool calls

### Platform assumptions

- Docker for containerized deployment
- Go 1.25+ (uses latest language features)
- OpenAI-compatible LLM API (LM Studio, Ollama, OpenAI, vLLM, etc.)

### Security

- **No secrets in code**: Use environment variables for API keys
- Consider `EncryptedConversationStore` for sensitive conversation data
- Tool execution has timeout protection (default 30s)

---

## 10. How AI tools and RAG should use this file

### Instructions for AI agents

- **Check existing implementations** before creating new tools or adapters
- **Follow the directory structure contract** when adding new code
- **Prefer cloud-native-utils** patterns for resilience and efficiency
- **Read CONTEXT.md first** before making architectural changes
- **Update CONTEXT.md** after significant architectural changes
- **Use established patterns**: functional options, interface segregation, typed errors

### Priority order for context

1. **CONTEXT.md** (this file) — Architecture, conventions, contracts
2. **README.md** — Project purpose, quick start, user-facing docs
3. **VENDOR.md** — External library usage patterns

### When to reference this file

- Adding new adapters or use cases
- Creating new domain entities or aggregates
- Reviewing code for convention compliance
- Understanding error handling patterns
