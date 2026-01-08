# CONTEXT.md

## 1. Project Purpose

This repository implements a **Go-based AI Agent** as a reusable library (`pkg/agent/`). The agent follows an **observe → decide → act → update** loop pattern to interact with Large Language Models (LLMs) and execute tools.

The project demonstrates:
- Clean, reusable agent library with minimal dependencies
- Integration with OpenAI-compatible APIs (e.g., LM Studio)
- Tool calling capabilities for LLM agents
- Event-driven patterns for observability
- Hooks/middleware system for extensibility
- Typed errors for robust error handling

---

## 2. Technology Stack

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.25+ |
| **Build System** | `just` (justfile task runner) |
| **Linting** | `golangci-lint` with comprehensive ruleset |
| **Testing** | Go standard `testing` package + `github.com/andygeiss/cloud-native-utils/assert` |
| **Containerization** | Docker (multi-stage build), Podman |
| **External Services** | LM Studio (OpenAI-compatible API), Kafka (optional) |

### Key Dependencies
- `github.com/andygeiss/cloud-native-utils` - Assertion utilities and cloud-native helpers

---

## 3. High-Level Architecture

The project provides a **reusable agent library** in `pkg/agent/` with domain use cases in `internal/domain/` and adapter implementations in `internal/adapters/`:

```
┌─────────────────────────────────────────────────────────────────┐
│                         cmd/cli                                 │
│                    (Application Entry Point)                    │
└────────────────────────────┬────────────────────────────────────┘
                             │ uses
┌────────────────────────────▼────────────────────────────────────┐
│                   internal/domain/chat                          │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  SendMessageUseCase, ClearConversationUseCase,           │   │
│  │  GetAgentStatsUseCase, TaskRunner (port)                 │   │
│  └──────────────────────────────────────────────────────────┘   │
└────────────────────────────┬────────────────────────────────────┘
                             │ uses
┌────────────────────────────▼────────────────────────────────────┐
│                        pkg/agent                                │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Agent, Task, Message, ToolCall, Result, LLMResponse     │   │
│  │  TaskService, LLMClient, ToolExecutor, EventPublisher    │   │
│  │  Types: AgentID, TaskID, ToolCallID, Role, Status        │   │
│  │  ToolDefinition, ParameterDefinition, ParameterType      │   │
│  │  Hooks, Errors (typed), TokenUsage, Metadata             │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    pkg/agent/events                      │   │
│  │  EventTaskStarted, EventTaskCompleted, EventTaskFailed,  │   │
│  │  EventToolCallExecuted                                   │   │
│  └──────────────────────────────────────────────────────────┘   │
└────────────────────────────┬────────────────────────────────────┘
                             │ implements interfaces
┌────────────────────────────▼────────────────────────────────────┐
│                 internal/adapters/outbound                      │
│  ┌────────────────┐  ┌────────────────┐  ┌──────────────────┐   │
│  │  OpenAIClient  │  │  ToolExecutor  │  │  EventPublisher  │   │
│  │ (LLM adapter)  │  │ (Tool adapter) │  │ (Event adapter)  │   │
│  └────────────────┘  └────────────────┘  └──────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                             │ uses
┌────────────────────────────▼────────────────────────────────────┐
│                           pkg/                                  │
│  ┌────────────────┐  ┌────────────────┐                         │
│  │     openai     │  │     event      │                         │
│  │ (API payloads) │  │  (interfaces)  │                         │
│  └────────────────┘  └────────────────┘                         │
└─────────────────────────────────────────────────────────────────┘
```

### Architectural Style
- **Reusable Library**: Agent framework exported via `pkg/agent/`
- **Agent Loop Pattern**: observe → decide → act → update
- **Interface-based Design**: LLMClient, ToolExecutor, EventPublisher interfaces
- **Functional Options**: Agent configuration via `With*` option functions
- **Hooks/Middleware**: Lifecycle callbacks for extensibility
- **Use Case Pattern**: Domain use cases in `internal/domain/` for application logic

---

## 4. Directory Structure (Contract)

```
go-agent/
├── cmd/
│   └── cli/                    # CLI application entry point
│       ├── main.go
│       └── main_test.go
├── internal/
│   ├── adapters/
│   │   └── outbound/           # Infrastructure adapters (LLM, tools, events)
│   │       ├── openai_client.go
│   │       ├── openai_client_test.go
│   │       ├── tool_executor.go
│   │       ├── tool_executor_test.go
│   │       ├── event_publisher.go
│   │       └── event_publisher_test.go
│   └── domain/
│       └── chat/               # Chat domain use cases
│           ├── ports.go              # TaskRunner interface
│           ├── send_message.go       # SendMessageUseCase
│           ├── send_message_test.go
│           ├── clear_conversation.go # ClearConversationUseCase
│           ├── clear_conversation_test.go
│           ├── get_agent_stats.go    # GetAgentStatsUseCase
│           └── get_agent_stats_test.go
├── pkg/
│   ├── agent/                  # Reusable agent library
│   │   ├── types.go            # ID types, Role, Status constants
│   │   ├── agent.go            # Agent aggregate root with options
│   │   ├── errors.go           # Typed errors (LLMError, ToolError, TaskError)
│   │   ├── hooks.go            # Lifecycle hooks/middleware
│   │   ├── llm_response.go     # LLM response wrapper
│   │   ├── message.go          # Conversation messages
│   │   ├── task.go             # Task entity with timestamps
│   │   ├── tool_call.go        # Tool call entity
│   │   ├── result.go           # Task execution result with metrics
│   │   ├── tool_definition.go  # Tool definition with parameter types
│   │   ├── ports.go            # Interfaces (LLMClient, ToolExecutor, EventPublisher)
│   │   ├── task_service.go     # Agent loop orchestration
│   │   └── events/             # Domain events
│   │       ├── events.go       # All event types and topic constants
│   │       └── events_test.go
│   ├── event/                  # Reusable event interfaces
│   └── openai/                 # OpenAI API payload structures
├── .justfile                   # Task runner configuration
├── .golangci.yml               # Linter configuration
├── Dockerfile                  # Multi-stage container build
├── docker-compose.yml          # Local development services
└── .env.example                # Environment variable template
```

### Rules for New Code

| What | Where |
|------|-------|
| **Agent library code** | `pkg/agent/` |
| **Domain events** | `pkg/agent/events/` |
| **Domain use cases** | `internal/domain/<context>/` |
| **Infrastructure adapters** | `internal/adapters/outbound/` |
| **Reusable utilities** | `pkg/` |
| **Application entry points** | `cmd/` |
| **Tests** | Same directory as source, `*_test.go` suffix |

---

## 5. Coding Conventions

### 5.1 General

- Keep modules small and focused
- Prefer pure functions where possible
- Domain layer must not import adapter layer
- Use cases orchestrate domain logic and depend on ports (interfaces)
- Adapters implement port interfaces defined in domain
- Use constructor functions (e.g., `NewAgent()`, `NewTask()`, `NewSendMessageUseCase()`)
- Use functional options pattern with `With*` methods for configuration
- Use builder pattern with method chaining for complex objects

### 5.2 Naming

| Element | Convention | Example |
|---------|------------|---------|
| **Files** | `snake_case.go` | `task_service.go`, `llm_response.go` |
| **Test files** | `*_test.go` | `task_service_test.go` |
| **Packages** | lowercase, single word | `agent`, `events`, `openai` |
| **Structs** | PascalCase | `Agent`, `TaskService`, `LLMResponse` |
| **Interfaces** | PascalCase, verb or noun | `LLMClient`, `ToolExecutor`, `EventPublisher` |
| **Constructors** | `New<Type>` | `NewAgent()`, `NewTask()`, `NewHooks()` |
| **Options** | `With<Field>` returning `Option` | `WithMaxIterations()`, `WithMetadata()` |
| **Builder methods** | `With<Field>` returning self | `WithParameter()`, `WithDuration()` |
| **ID types** | `<Entity>ID` | `AgentID`, `TaskID`, `ToolCallID` |
| **Status types** | `<Entity>Status` | `TaskStatus`, `ToolCallStatus` |
| **Event types** | `Event<Action>` | `EventTaskStarted`, `EventTaskCompleted` |
| **Event topics** | `Topic<Action>` | `TopicTaskStarted` |
| **Error types** | `<Context>Error` | `LLMError`, `ToolError`, `TaskError` |
| **Constants** | PascalCase with prefix | `RoleSystem`, `TaskStatusPending` |

### 5.3 Error Handling & Logging

**Sentinel Errors** (in `pkg/agent/errors.go`):
- `ErrMaxIterationsReached` - Agent exceeded max iterations
- `ErrContextCanceled` - Context was canceled
- `ErrNoResponse` - LLM returned empty response
- `ErrToolNotFound` - Unknown tool requested
- `ErrInvalidArguments` - Malformed tool arguments

**Typed Errors**:
- `LLMError` - Wraps LLM client errors with context
- `ToolError` - Wraps tool execution errors with tool name
- `TaskError` - Wraps task errors with task ID

All typed errors implement `Unwrap()` for `errors.Is`/`errors.As` support:
```go
if errors.Is(err, agent.ErrMaxIterationsReached) { ... }
var toolErr *agent.ToolError
if errors.As(err, &toolErr) { ... }
```

**CLI Feedback** (emoji prefixes):
- `🤖` Assistant messages
- `❌` Errors
- `⚠️` Warnings
- `🗑️` Actions
- `📊` Statistics
- `📈` Summary
- `🔧` Tool calls

### 5.4 Testing

- **Framework**: Go standard `testing` package
- **Assertions**: `github.com/andygeiss/cloud-native-utils/assert`
- **Naming**: `Test_<Type>_<Method>_<Scenario>_Should_<Expected>`
  - Example: `Test_Agent_AddTask_With_OneTask_Should_HaveOneTask`
- **Structure**: Arrange → Act → Assert pattern with comments
- **Location**: Tests in same directory as source
- **Integration tests**: Tagged with `//go:build integration`

```go
func Test_Agent_NewAgent_With_ValidParams_Should_Return_Agent(t *testing.T) {
    // Arrange & Act
    ag := agent.NewAgent("test-id", "test prompt")
    
    // Assert
    assert.That(t, "ID must match", ag.ID, agent.AgentID("test-id"))
}
```

### 5.5 Formatting & Linting

- **Formatter**: `golangci-lint fmt ./...`
- **Linter**: `golangci-lint run ./...`
- **Config**: `.golangci.yml`

Key lint rules:
- All linters enabled by default, with specific exclusions documented
- `interface{}` auto-replaced with `any`
- Field alignment warnings for structs (optimize or add `//nolint:govet` with reason)
- Use `slices.Contains()` instead of manual loops
- Maximum cyclomatic complexity: 10 (cyclop linter)
- No named returns (nonamedreturns linter)

---

## 6. Cross-Cutting Concerns and Reusable Patterns

### Functional Options Pattern
Used for configuring `Agent` instances:
```go
ag := agent.NewAgent("id", "prompt",
    agent.WithMaxIterations(20),
    agent.WithMaxMessages(100),
    agent.WithMetadata(agent.Metadata{"key": "value"}),
)
```

### Hooks/Middleware System
Lifecycle callbacks for task execution extensibility:
```go
hooks := agent.NewHooks().
    WithBeforeTask(func(ctx context.Context, ag *agent.Agent, task *agent.Task) error { ... }).
    WithAfterTask(func(ctx context.Context, ag *agent.Agent, task *agent.Task) error { ... }).
    WithBeforeLLMCall(func(ctx context.Context, ag *agent.Agent, task *agent.Task) error { ... }).
    WithAfterLLMCall(func(ctx context.Context, ag *agent.Agent, task *agent.Task) error { ... }).
    WithBeforeToolCall(func(ctx context.Context, ag *agent.Agent, tc *agent.ToolCall) error { ... }).
    WithAfterToolCall(func(ctx context.Context, ag *agent.Agent, tc *agent.ToolCall) error { ... })

taskService.WithHooks(hooks)
```

### Use Case Pattern
Application use cases live in `internal/domain/<context>/`:
```go
// Create use cases with dependencies
sendMessage := chat.NewSendMessageUseCase(taskService, &agent)
clearConversation := chat.NewClearConversationUseCase(&agent)
getAgentStats := chat.NewGetAgentStatsUseCase(&agent)

// Execute use case
output, err := sendMessage.Execute(ctx, chat.SendMessageInput{Message: "Hello"})
```

Use cases:
- **SendMessageUseCase** - Send a message and get a response
- **ClearConversationUseCase** - Clear conversation history
- **GetAgentStatsUseCase** - Retrieve agent statistics

### Event System
- **Interface**: `pkg/event.Event` (must implement `Topic() string`)
- **Publisher**: `agent.EventPublisher` interface in `pkg/agent/ports.go`
- **Domain events**: `pkg/agent/events/`
  - `events.go` - All event types and topic constants
- Events are immutable structs with constructor functions

### Tool System
- **Interface**: `agent.ToolExecutor` in `pkg/agent/ports.go`
- **Registration**: Tools registered via `RegisterTool(name, fn)` in adapters
- **Definition**: `agent.ToolDefinition` with typed parameters:
```go
td := agent.NewToolDefinition("calculate", "Perform calculation").
    WithParameterDef(agent.NewParameterDefinition("expression", agent.ParamTypeString).
        WithDescription("Math expression").
        WithRequired())
```

### Parameter Types
Supported parameter types for tool definitions:
- `ParamTypeString`, `ParamTypeNumber`, `ParamTypeInteger`
- `ParamTypeBoolean`, `ParamTypeArray`, `ParamTypeObject`

### LLM Integration
- **Interface**: `agent.LLMClient` in `pkg/agent/ports.go`
- **Implementation**: OpenAI-compatible API adapter in `adapters/outbound/`
- **Configuration**: Base URL and model via environment/flags

### Memory Management
Agent supports message trimming to prevent context overflow:
```go
ag := agent.NewAgent("id", "prompt", agent.WithMaxMessages(100))
// Older messages are automatically trimmed when limit exceeded
```

### Result Metrics
Task results include execution metrics:
```go
result.Duration       // Execution time
result.IterationCount // Agent loop iterations
result.ToolCallCount  // Tool calls made
result.Tokens         // TokenUsage{PromptTokens, CompletionTokens, TotalTokens}
```

### Task Timestamps
Tasks track lifecycle timing:
```go
task.CreatedAt   // When task was created
task.StartedAt   // When task started executing
task.CompletedAt // When task completed/failed
task.Duration()  // Execution duration
task.WaitTime()  // Queue wait time
task.Iterations  // Number of agent loop iterations
```

### Configuration Management
- **Environment**: `.env` file (local-only, copy from `.env.example`)
- **Flags**: Command-line flags for runtime configuration
- **Docker**: Environment variables via `docker-compose.yml`

---

## 7. Using the Agent Library

### Import the Library
```go
import (
    "github.com/andygeiss/go-agent/pkg/agent"
    "github.com/andygeiss/go-agent/pkg/agent/events"
)
```

### Implement the Interfaces
1. Implement `agent.LLMClient` for your LLM provider
2. Implement `agent.ToolExecutor` for your tool registry
3. Implement `agent.EventPublisher` for observability

### Create and Run Tasks
```go
// Create service with dependencies
taskService := agent.NewTaskService(llmClient, toolExecutor, publisher)

// Optional: Add hooks for logging/metrics
hooks := agent.NewHooks().
    WithAfterToolCall(func(ctx context.Context, ag *agent.Agent, tc *agent.ToolCall) error {
        log.Printf("Tool %s completed: %s", tc.Name, tc.Result)
        return nil
    })
taskService.WithHooks(hooks)

// Create agent with options
ag := agent.NewAgent("my-agent", "You are a helpful assistant",
    agent.WithMaxIterations(20),
    agent.WithMaxMessages(100),
    agent.WithMetadata(agent.Metadata{"version": "1.0"}),
)

// Run a task
task := agent.NewTask("task-1", "chat", "Hello!")
result, err := taskService.RunTask(ctx, &ag, task)

// Access metrics
fmt.Printf("Completed in %s with %d iterations\n", result.Duration, result.IterationCount)
```

### Customization Points
- Add new tools by implementing `ToolExecutor`
- Subscribe to events: `TaskStarted`, `TaskCompleted`, `TaskFailed`, `ToolCallExecuted`
- Use hooks for: logging, metrics, rate limiting, authorization, caching
- Extend agent capabilities by composing with your own types

---

## 8. Key Commands & Workflows

| Command | Description |
|---------|-------------|
| `just setup` | Install dependencies (golangci-lint, just) |
| `just fmt` | Format Go code |
| `just lint` | Run linter checks |
| `just test` | Run unit tests with coverage |
| `just test-integration` | Run integration tests (requires LM Studio) |
| `just run` | Run CLI application locally |
| `just build` | Build Docker image |
| `just up` | Start all services (build + docker-compose) |
| `just down` | Stop all services |
| `just profile` | Generate CPU profile and visualization |

### CLI Options
```bash
go run ./cmd/cli \
    -url http://localhost:1234 \    # LM Studio API URL
    -model <model-name> \           # Model to use
    -max-iterations 10 \            # Max agent loop iterations
    -max-messages 50 \              # Max messages to retain (0=unlimited)
    -verbose                         # Show detailed metrics
```

### CLI Commands
- `quit` / `exit` - Exit the CLI with session summary
- `clear` - Clear conversation history
- `stats` - Show agent statistics (messages, tasks, completion rates)

### Environment Setup
```bash
# Copy environment template
cp .env.example .env

# Edit .env with your configuration
# Required for integration tests:
# - LM_STUDIO_URL=http://localhost:1234
# - LM_STUDIO_MODEL=<your-model>
```

---

## 9. Important Notes & Constraints

### Security
- `.env` files are local-only (gitignored)
- No secrets in source code
- API keys passed via environment variables

### Performance
- Struct field alignment optimized for memory (lint enforced)
- Use `slices.Contains()` over manual loops
- Profile-guided optimization (PGO) supported via `just profile`
- Memory management via `MaxMessages` prevents unbounded growth

### Platform Assumptions
- macOS/Linux development environment
- Docker/Podman for containerization
- LM Studio for local LLM inference (OpenAI-compatible API)

### Limitations & Technical Debt
- No inbound adapters (HTTP/gRPC) - CLI only
- Single bounded context (`agent`)
- Tool executor has limited demo tools (extensible)
- Token usage tracking not yet implemented in OpenAI adapter

---

## 10. How AI Tools and RAG Should Use This File

This file serves as the **authoritative project context** for:
- AI coding agents working on this repository
- RAG systems retrieving project information
- Developers onboarding to the codebase

### Instructions for AI Agents
1. **Always read `CONTEXT.md` first** before major changes or refactors
2. **Treat architectural boundaries as constraints** - domain must not import adapters
3. **Follow naming conventions** exactly as documented
4. **Use functional options** for Agent configuration
5. **Use typed errors** for error handling (LLMError, ToolError, TaskError)
6. **Add hooks** for cross-cutting concerns instead of modifying core logic
7. **Add tests** following the established patterns
8. **Update this file** if adding new architectural patterns or conventions

### Combining with Other Documents
- `CONTEXT.md` - Architecture and conventions (this file)
- `README.md` - User-facing documentation
- `VENDOR.md` - Vendor library documentation
- `.golangci.yml` - Detailed lint rules
- `.justfile` - Available commands and workflows
