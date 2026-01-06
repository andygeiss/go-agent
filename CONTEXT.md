# CONTEXT.md

## 1. Project Purpose

This repository implements a **Go-based AI Agent** as a reusable library (`pkg/agent/`). The agent follows an **observe → decide → act → update** loop pattern to interact with Large Language Models (LLMs) and execute tools.

The project demonstrates:
- Clean, reusable agent library with minimal dependencies
- Integration with OpenAI-compatible APIs (e.g., LM Studio)
- Tool calling capabilities for LLM agents
- Event-driven patterns for observability

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

The project provides a **reusable agent library** in `pkg/agent/` with adapter implementations in `internal/`:

```
┌─────────────────────────────────────────────────────────────────┐
│                         cmd/cli                                 │
│                    (Application Entry Point)                    │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                        pkg/agent                                │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Agent, Task, Message, ToolCall, Result, LLMResponse     │   │
│  │  TaskService, LLMClient, ToolExecutor, EventPublisher    │   │
│  │  Types: AgentID, TaskID, ToolCallID, Role, Status        │   │
│  │  ToolDefinition                                          │   │
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

---

## 4. Directory Structure (Contract)

```
go-agent/
├── cmd/
│   └── cli/                    # CLI application entry point
│       ├── main.go
│       └── main_test.go
├── internal/
│   └── adapters/
│       └── outbound/           # Infrastructure adapters (LLM, tools, events)
│           ├── openai_client.go
│           ├── tool_executor.go
│           ├── tool_executor_test.go
│           ├── event_publisher.go
│           └── event_publisher_test.go
├── pkg/
│   ├── agent/                  # Reusable agent library
│   │   ├── types.go            # ID types, Role, Status constants
│   │   ├── agent.go            # Agent aggregate root
│   │   ├── llm_response.go     # LLM response wrapper
│   │   ├── message.go          # Conversation messages
│   │   ├── task.go             # Task entity
│   │   ├── tool_call.go        # Tool call entity
│   │   ├── result.go           # Task execution result
│   │   ├── tool_definition.go  # Tool definition for LLM
│   │   ├── ports.go            # Interfaces (LLMClient, ToolExecutor, EventPublisher)
│   │   ├── task_service.go     # Agent loop orchestration
│   │   └── events/             # Domain events
│   │       ├── events.go       # All event types and topic constants
│   │       └── events_test.go
│   ├── event/                  # Reusable event interfaces
│   └── openai/                 # OpenAI API payload structures
├── tools/                      # Development/profiling scripts
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
- Adapters implement port interfaces defined in domain
- Use constructor functions (e.g., `NewAgent()`, `NewTask()`)
- Use fluent/builder pattern with `With*` methods for optional configuration

### 5.2 Naming

| Element | Convention | Example |
|---------|------------|---------|
| **Files** | `snake_case.go` | `task_service.go`, `llm_response.go` |
| **Test files** | `*_test.go` | `task_service_test.go` |
| **Packages** | lowercase, single word | `agent`, `events`, `openai` |
| **Structs** | PascalCase | `Agent`, `TaskService`, `LLMResponse` |
| **Interfaces** | PascalCase, verb or noun | `LLMClient`, `ToolExecutor`, `EventPublisher` |
| **Constructors** | `New<Type>` | `NewAgent()`, `NewTask()` |
| **Builder methods** | `With<Field>` | `WithMaxIterations()`, `WithParameter()` |
| **ID types** | `<Entity>ID` | `AgentID`, `TaskID`, `ToolCallID` |
| **Status types** | `<Entity>Status` | `TaskStatus`, `ToolCallStatus` |
| **Event types** | `Event<Action>` | `EventTaskStarted`, `EventTaskCompleted` |
| **Event topics** | `Topic<Action>` | `TopicTaskStarted` |
| **Constants** | PascalCase with prefix | `RoleSystem`, `TaskStatusPending` |

### 5.3 Error Handling & Logging

- Return errors using `fmt.Errorf("context: %w", err)` for wrapping
- Domain services return `(Result, error)` tuples
- Use typed errors where meaningful
- CLI uses `fmt.Printf` with emoji prefixes for user feedback:
  - `🤖` Assistant messages
  - `❌` Errors
  - `⚠️` Warnings
  - `🗑️` Actions

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

---

## 6. Cross-Cutting Concerns and Reusable Patterns

### Event System
- **Interface**: `pkg/event.Event` (must implement `Topic() string`)
- **Publisher**: `agent.EventPublisher` interface in `pkg/agent/ports.go`
- **Domain events**: `pkg/agent/events/`
  - `events.go` - All event types and topic constants
- Events are immutable structs with constructor functions

### Tool System
- **Interface**: `agent.ToolExecutor` in `pkg/agent/ports.go`
- **Registration**: Tools registered via `RegisterTool(name, fn)` in adapters
- **Definition**: `agent.ToolDefinition` with name, description, parameters

### LLM Integration
- **Interface**: `agent.LLMClient` in `pkg/agent/ports.go`
- **Implementation**: OpenAI-compatible API adapter in `adapters/outbound/`
- **Configuration**: Base URL and model via environment/flags

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
taskService := agent.NewTaskService(llmClient, toolExecutor, publisher)
ag := agent.NewAgent("my-agent", "You are a helpful assistant")
task := agent.NewTask("task-1", "chat", "Hello!")
result, err := taskService.RunTask(ctx, &ag, task)
```

### Customization Points
- Add new tools by implementing `ToolExecutor`
- Subscribe to events: `TaskStarted`, `TaskCompleted`, `TaskFailed`, `ToolCallExecuted`
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

### Platform Assumptions
- macOS/Linux development environment
- Docker/Podman for containerization
- LM Studio for local LLM inference (OpenAI-compatible API)

### Limitations & Technical Debt
- No inbound adapters (HTTP/gRPC) - CLI only
- Single bounded context (`agent`)
- Tool executor has limited demo tools (extensible)

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
4. **Add tests** following the established patterns
5. **Update this file** if adding new architectural patterns or conventions

### Combining with Other Documents
- `CONTEXT.md` - Architecture and conventions (this file)
- `README.md` - User-facing documentation (if exists)
- `.golangci.yml` - Detailed lint rules
- `.justfile` - Available commands and workflows
