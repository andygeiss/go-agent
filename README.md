<p align="center">
<img src="https://github.com/andygeiss/go-ddd-hex-starter/blob/main/cmd/server/assets/static/img/icon-192.png?raw=true" width="100"/>
</p>

# Go Agent

[![Go Reference](https://pkg.go.dev/badge/github.com/andygeiss/go-agent.svg)](https://pkg.go.dev/github.com/andygeiss/go-agent)
[![License](https://img.shields.io/github/license/andygeiss/go-agent)](https://github.com/andygeiss/go-agent/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/v/release/andygeiss/go-agent)](https://github.com/andygeiss/go-agent/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygeiss/go-agent)](https://goreportcard.com/report/github.com/andygeiss/go-agent)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/85ef3344ec784fe9b8dd9052e6172b5d)](https://app.codacy.com/gh/andygeiss/go-agent/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/85ef3344ec784fe9b8dd9052e6172b5d)](https://app.codacy.com/gh/andygeiss/go-agent/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage)

A Go-based AI Agent library implementing the **Observe → Decide → Act → Update** loop pattern to interact with Large Language Models (LLMs) and execute tools.

## Features

- **Reusable Library** - Import `pkg/agent` to build LLM-powered applications
- **LLM Integration** - OpenAI-compatible API support (works with LM Studio, OpenAI, etc.)
- **Tool Calling** - Extensible tool system with typed parameter definitions
- **Parallel Tool Execution** - Execute multiple tool calls concurrently for improved performance
- **Conversation Persistence** - Save and restore conversation history with pluggable storage backends
- **Encryption at Rest** - AES-GCM encryption for sensitive conversation data
- **Event-Driven** - Domain events for observability and extensibility
- **Hooks/Middleware** - Lifecycle callbacks for logging, metrics, authorization
- **Functional Options** - Clean configuration with `With*` option functions
- **Typed Errors** - Structured error handling with `errors.Is`/`errors.As` support
- **Memory Management** - Configurable message limits to prevent context overflow

## Quick Start

### Prerequisites

- Go 1.25 or higher
- [just](https://github.com/casey/just) command runner
- [golangci-lint](https://golangci-lint.run/) for formatting and linting
- [LM Studio](https://lmstudio.ai/) (or any OpenAI-compatible API) for LLM inference

### Installation

```bash
# Clone the repository
git clone https://github.com/andygeiss/go-agent.git
cd go-agent

# Install development dependencies
just setup

# Copy environment configuration
cp .env.example .env
```

### Running the Agent

```bash
# Start LM Studio with a model loaded, then:
just run
```

This starts an interactive CLI where you can chat with the agent:

```
🤖 Go Agent Demo - LM Studio Chat
==================================
Connecting to LM Studio at: http://localhost:1234
Using model: default
Max iterations: 10 | Max messages: 50

Commands: 'quit'/'exit' to stop, 'clear' to reset, 'stats' for agent stats

You: What time is it?
🤖 Assistant: The current time is 2026-01-08T15:30:45Z.

You: stats
📊 Agent Statistics
-------------------
Agent ID:        demo-agent
Messages:        2
Tasks:           1 (✓ 1 completed, ✗ 0 failed)
Max iterations:  10
Max messages:    50

You: quit

📈 Session summary: 1 tasks (✓ 1, ✗ 0), 2 messages
Goodbye! 👋
```

## Project Structure

```
go-agent/
├── cmd/cli/                    # CLI application entry point
├── internal/
│   ├── adapters/outbound/      # Infrastructure adapters (LLM, tools, events)
│   └── domain/chat/            # Chat domain use cases
├── pkg/
│   ├── agent/                  # Reusable agent library
│   │   ├── types.go            # ID types, Role, Status constants
│   │   ├── agent.go            # Agent aggregate with options
│   │   ├── errors.go           # Typed errors (LLMError, ToolError, TaskError)
│   │   ├── hooks.go            # Lifecycle hooks/middleware
│   │   ├── task.go             # Task entity with timestamps
│   │   ├── task_service.go     # Agent loop orchestration
│   │   ├── message.go          # Conversation messages
│   │   ├── llm_response.go     # LLM response wrapper
│   │   ├── result.go           # Task result with metrics
│   │   ├── tool_call.go        # Tool call entity
│   │   ├── tool_definition.go  # Tool definitions with parameter types
│   │   ├── ports.go            # Interfaces (LLMClient, ToolExecutor)
│   │   └── events/             # Domain events
│   ├── event/                  # Event interfaces
│   └── openai/                 # OpenAI API structures
```

## Available Commands

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

## Configuration

The agent is configured via environment variables. Copy `.env.example` to `.env` and configure:

```bash
# LM Studio connection
LM_STUDIO_URL=http://localhost:1234
LM_STUDIO_MODEL=your-model-name
```

### CLI Options

```bash
go run ./cmd/cli \
    -url http://localhost:1234 \    # LM Studio API URL
    -model <model-name> \           # Model to use
    -max-iterations 10 \            # Max agent loop iterations per task
    -max-messages 50 \              # Max messages to retain (0=unlimited)
    -verbose                         # Show detailed metrics after each response
```

### CLI Commands

| Command | Description |
|---------|-------------|
| `quit` / `exit` | Exit with session summary |
| `clear` | Clear conversation history |
| `stats` | Show agent statistics |

## Architecture

The project provides a reusable agent library in `pkg/agent/`:

```
┌─────────────────────────────────────────────────────────────┐
│                      Application (CLI)                      │
└─────────────────────────────┬───────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────┐
│                     pkg/agent Library                       │
│  • Agent, Task, Message     • TaskService (Agent Loop)      │
│  • LLMClient interface      • ToolExecutor interface        │
└─────────────────────────────┬───────────────────────────────┘
                              │ implements
┌─────────────────────────────▼───────────────────────────────┐
│                    Adapter Layer                            │
│  • OpenAIClient (LLM)  • ToolExecutor  • EventPublisher     │
└─────────────────────────────────────────────────────────────┘
```

### Using the Library

```go
import (
    "github.com/andygeiss/go-agent/pkg/agent"
)

// Create agent infrastructure
taskService := agent.NewTaskService(llmClient, toolExecutor, publisher)

// Optional: Add hooks for logging/metrics
hooks := agent.NewHooks().
    WithAfterToolCall(func(ctx context.Context, ag *agent.Agent, tc *agent.ToolCall) error {
        log.Printf("Tool %s executed: %s", tc.Name, tc.Result)
        return nil
    })
taskService.WithHooks(hooks)

// Optional: Enable parallel tool execution
taskService.WithParallelToolExecution()

// Create agent with options
ag := agent.NewAgent("my-agent", "You are a helpful assistant",
    agent.WithMaxIterations(20),
    agent.WithMaxMessages(100),
    agent.WithMetadata(agent.Metadata{"version": "1.0"}),
)

// Run a task
task := agent.NewTask("task-1", "chat", "Hello!")
result, err := taskService.RunTask(ctx, &ag, task)

// Access execution metrics
fmt.Printf("Completed in %s with %d iterations, %d tool calls\n",
    result.Duration, result.IterationCount, result.ToolCallCount)
```

The agent operates in a continuous loop:
1. **Observe** - Gather current state and conversation history
2. **Decide** - Call LLM with context and available tools
3. **Act** - Execute tool calls if requested
4. **Update** - Update state and continue or complete

For detailed architectural documentation, see [CONTEXT.md](CONTEXT.md).

## Built-in Tools

The agent comes with demo tools:

| Tool | Description |
|------|-------------|
| `get_current_time` | Returns the current date and time |
| `calculate` | Performs arithmetic calculations |

### Adding Custom Tools

Register new tools in `internal/adapters/outbound/tool_executor.go`:

```go
executor.RegisterTool("my_tool", func(ctx context.Context, args string) (string, error) {
    // Parse args (JSON) and execute
    return "result", nil
})
```

Define tool parameters with types:

```go
toolDef := agent.NewToolDefinition("my_tool", "Description of my tool").
    WithParameterDef(agent.NewParameterDefinition("query", agent.ParamTypeString).
        WithDescription("The search query").
        WithRequired()).
    WithParameterDef(agent.NewParameterDefinition("limit", agent.ParamTypeInteger).
        WithDescription("Max results").
        WithDefault("10"))
```

### Parallel Tool Execution

When the LLM returns multiple tool calls in a single response, execute them concurrently:

```go
// Enable parallel execution for improved performance with I/O-bound tools
taskService := agent.NewTaskService(llmClient, toolExecutor, publisher).
    WithParallelToolExecution()

// Without parallel execution: tools run sequentially (default)
// With parallel execution: tools run concurrently using efficiency.Process
```

**Note**: Parallel execution provides significant benefits for I/O-bound tools (API calls, file operations) but adds coordination overhead for CPU-bound operations.

### Error Handling

The library provides typed errors for robust error handling:

```go
result, err := taskService.RunTask(ctx, &ag, task)
if err != nil {
    // Check for specific error types
    if errors.Is(err, agent.ErrMaxIterationsReached) {
        log.Println("Task exceeded iteration limit")
    }
    
    var toolErr *agent.ToolError
    if errors.As(err, &toolErr) {
        log.Printf("Tool %s failed: %s", toolErr.ToolName, toolErr.Message)
    }
}
```

### Hooks/Middleware

Add cross-cutting concerns without modifying core logic:

```go
hooks := agent.NewHooks().
    WithBeforeTask(func(ctx context.Context, ag *agent.Agent, task *agent.Task) error {
        log.Printf("Starting task: %s", task.ID)
        return nil
    }).
    WithAfterTask(func(ctx context.Context, ag *agent.Agent, task *agent.Task) error {
        log.Printf("Task completed in %s", task.Duration())
        return nil
    }).
    WithBeforeLLMCall(func(ctx context.Context, ag *agent.Agent, task *agent.Task) error {
        // Rate limiting, logging, etc.
        return nil
    }).
    WithAfterToolCall(func(ctx context.Context, ag *agent.Agent, tc *agent.ToolCall) error {
        // Log tool execution, cache results, etc.
        return nil
    })

taskService.WithHooks(hooks)
```

### Conversation Persistence

Save and restore conversation history with pluggable storage backends:

```go
import (
    "github.com/andygeiss/go-agent/internal/adapters/outbound"
)

// In-memory storage (for testing)
store := outbound.NewInMemoryConversationStore()

// JSON file storage (for production)
store := outbound.NewJsonFileConversationStore("conversations.json")

// Save conversation
ctx := context.Background()
err := store.Save(ctx, agent.AgentID("my-agent"), messages)

// Load conversation
messages, err := store.Load(ctx, agent.AgentID("my-agent"))

// Clear conversation
err := store.Clear(ctx, agent.AgentID("my-agent"))
```

### Encrypted Storage

Protect sensitive conversation data with AES-GCM encryption:

```go
import (
    "github.com/andygeiss/cloud-native-utils/security"
    "github.com/andygeiss/go-agent/internal/adapters/outbound"
)

// Generate a 32-byte encryption key (store securely!)
key := security.GenerateKey()

// Create encrypted store
baseStore := outbound.NewJsonFileConversationStore("conversations.json")
encStore := outbound.NewEncryptedConversationStore(baseStore, key)

// Use like any ConversationStore - encryption/decryption is automatic
err := encStore.Save(ctx, agentID, messages)
messages, err := encStore.Load(ctx, agentID)
```

## Testing

```bash
# Run all unit tests
just test

# Run with verbose output
go test -v ./internal/...

# Run integration tests (requires LM Studio running)
just test-integration

# Run benchmarks
go test -bench=. -benchmem ./pkg/agent/
```

### Benchmarks

Performance benchmarks for core operations:

| Benchmark | Time/op | Allocs/op |
|-----------|---------|-----------|
| `DirectCompletion` | ~500ns | 11 |
| `SingleToolCall` | ~727ns | 17 |
| `MultipleToolCalls_Sequential` | ~925ns | 20 |
| `MultipleToolCalls_Parallel` | ~8.7µs | 44 |
| `Message_Create` | ~2ns | 0 |
| `Agent_Create` | ~45ns | 2 |
| `Event_Create` | ~1.6ns | 0 |

*Measured on Apple M4 Pro. Parallel execution shows higher overhead in synthetic benchmarks but provides real benefits with I/O-bound tool operations.*

## Docker

```bash
# Build the image
just build

# Run with docker-compose
just up

# Stop services
just down
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Follow the coding conventions in [CONTEXT.md](CONTEXT.md)
4. Run `just fmt` and `just lint` before committing
5. Add tests for new functionality
6. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Documentation

- [CONTEXT.md](CONTEXT.md) — Architecture, conventions, and project contracts
- [VENDOR.md](VENDOR.md) — Approved vendor libraries and usage patterns
- [AGENTS.md](AGENTS.md) — AI agent definitions for this repository

## Acknowledgments

- Built with [Go](https://go.dev)
- Architecture inspired by [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/) and [Domain-Driven Design](https://www.domainlanguage.com/ddd/)
