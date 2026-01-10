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

A reusable AI agent framework for Go implementing the **observe → decide → act → update** loop pattern for LLM-based task execution.

---

## Overview

**go-agent** provides a clean, production-ready foundation for building AI agents with tool use capabilities in Go. It features:

- 🏗️ **Hexagonal Architecture** — Clean separation of domain logic and infrastructure
- 🔄 **Agent Loop Pattern** — Observe → Decide → Act → Update cycle for autonomous task execution
- 📡 **Event-Driven** — Observable task lifecycle via domain events
- 🧠 **Memory System** — Long-term context storage with search and filtering capabilities
- 🛡️ **Resilience Patterns** — Breaker, debounce, retry, throttle, and timeout
- 🔧 **Tool Use** — Extensible tool system with type-safe definitions

Works with any OpenAI-compatible API (LM Studio, Ollama, OpenAI, vLLM, etc.).

---

## Table of Contents

- [Architecture](#architecture)
- [CLI Usage](#cli-usage)
- [Configuration](#configuration)
- [Contributing](#contributing)
- [Creating Custom Tools](#creating-custom-tools)
- [Docker](#docker)
- [Features](#features)
- [Installation](#installation)
- [License](#license)
- [Project Structure](#project-structure)
- [Quick Start](#quick-start)
- [Testing](#testing)

---

## Installation

```bash
go get github.com/andygeiss/go-agent
```

**Requirements:** Go 1.25.5+

---

## Quick Start

### Run the CLI Demo

```bash
# Clone the repository
git clone https://github.com/andygeiss/go-agent.git
cd go-agent

# Start LM Studio (or any OpenAI-compatible server) on localhost:1234

# Run the CLI
go run ./cmd/cli -model <your-model-name>
```

### Use as a Library

```go
package main

import (
    "context"
    "github.com/andygeiss/cloud-native-utils/messaging"
    "github.com/andygeiss/go-agent/internal/adapters/outbound"
    "github.com/andygeiss/go-agent/internal/domain/agent"
    "github.com/andygeiss/go-agent/internal/domain/tooling"
)

func main() {
    // Create infrastructure
    dispatcher := messaging.NewExternalDispatcher()
    llmClient := outbound.NewOpenAIClient("http://localhost:1234", "your-model")
    toolExecutor := outbound.NewToolExecutor()
    publisher := outbound.NewEventPublisher(dispatcher)

    // Register tools
    calcTool := tooling.NewCalculateTool()
    toolExecutor.RegisterTool("calculate", calcTool.Func)
    toolExecutor.RegisterToolDefinition(calcTool.Definition)

    // Create agent
    ag := agent.NewAgent("my-agent", "You are a helpful assistant.",
        agent.WithMaxIterations(10),
        agent.WithMaxMessages(50),
    )

    // Create task service and run
    taskService := agent.NewTaskService(llmClient, toolExecutor, publisher)
    task := agent.NewTask("task-1", "chat", "What is 42 * 17?")
    
    result, _ := taskService.RunTask(context.Background(), &ag, task)
    println(result.Output)
}
```

---

## Architecture

The project follows **hexagonal architecture** (ports and adapters) with domain-driven design:

```
┌─────────────────────────────────────────────────────────────────┐
│                         cmd/cli                                 │
│                    (Application Entry)                          │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                     internal/domain                             │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ agent/        Core aggregate, task service, types           ││
│  │ chatting/     Use cases: SendMessage, ClearConversation     ││
│  │ memorizing/   Use cases: WriteNote, SearchNotes             ││
│  │ tooling/      Tool implementations                          ││
│  │ openai/       OpenAI API types (request, response, tool)    ││
│  └─────────────────────────────────────────────────────────────┘│
└──────────────────────────────┬──────────────────────────────────┘
                               │ depends on interfaces (ports)
┌──────────────────────────────▼──────────────────────────────────┐
│                   internal/adapters/outbound                    │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ openai_client.go    LLMClient implementation                ││
│  │ tool_executor.go    ToolExecutor implementation             ││
│  │ event_publisher.go  EventPublisher implementation           ││
│  │ memory_store.go     MemoryStore implementation              ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

### Agent Loop

The core agent implements an iterative loop:

1. **Observe** — Receive user input, build message context
2. **Decide** — Call LLM with messages and available tools
3. **Act** — Execute any tool calls requested by the LLM
4. **Update** — Add results to conversation, check termination
5. **Repeat** — Continue until task completes or max iterations reached

For detailed architecture documentation, see [CONTEXT.md](CONTEXT.md).

---

## Features

### Built-in Tools (alphabetically sorted)

| Tool | Description |
|------|-------------|
| `calculate` | Safe arithmetic expression evaluator with operator precedence |
| `get_current_time` | Returns current date and time in RFC3339 format |
| `memory_get` | Retrieve a specific memory note by ID |
| `memory_search` | Search memory notes with query and filters |
| `memory_write` | Store a new memory note with metadata |

### Domain Events (alphabetically sorted)

Subscribe to task lifecycle events:

- `agent.task.completed` — Task finishes successfully
- `agent.task.failed` — Task terminates with error
- `agent.task.started` — Task begins execution
- `agent.toolcall.executed` — Tool call completes

### Lifecycle Hooks (alphabetically sorted)

```go
hooks := agent.NewHooks().
    WithAfterLLMCall(func(ctx context.Context, ag *agent.Agent, t *agent.Task) error {
        log.Println("LLM response received")
        return nil
    }).
    WithAfterTask(func(ctx context.Context, ag *agent.Agent, t *agent.Task) error {
        log.Println("Task finished:", t.Status)
        return nil
    }).
    WithAfterToolCall(func(ctx context.Context, ag *agent.Agent, tc *agent.ToolCall) error {
        log.Println("Tool executed:", tc.Name, "→", tc.Result)
        return nil
    }).
    WithBeforeLLMCall(func(ctx context.Context, ag *agent.Agent, t *agent.Task) error {
        log.Println("Calling LLM...")
        return nil
    }).
    WithBeforeTask(func(ctx context.Context, ag *agent.Agent, t *agent.Task) error {
        log.Println("Starting task:", t.Name)
        return nil
    }).
    WithBeforeToolCall(func(ctx context.Context, ag *agent.Agent, tc *agent.ToolCall) error {
        log.Println("Executing tool:", tc.Name)
        return nil
    })

taskService.WithHooks(hooks)
```

### Resilience Patterns

The `OpenAIClient` includes configurable resilience (alphabetically sorted):

- **Circuit Breaker**: Opens after 5 consecutive failures (configurable)
- **Debounce**: Coalesces rapid calls (disabled by default)
- **Retry**: 3 attempts with 2s delay (configurable)
- **Throttling**: Rate limiting via token bucket (disabled by default)
- **Timeout**: HTTP (60s) and LLM call (120s) timeouts (configurable)

---

## CLI Usage

```bash
go run ./cmd/cli [flags]
```

### Commands (during chat, alphabetically sorted)

| Command | Description |
|---------|-------------|
| `clear` | Reset conversation history |
| `quit` / `exit` | Exit the CLI |
| `stats` | Show agent statistics |

### Flags (alphabetically sorted)

| Flag | Default | Description |
|------|---------|-------------|
| `-max-iterations` | `10` | Max iterations per task |
| `-max-messages` | `50` | Max messages to retain (0 = unlimited) |
| `-model` | `$LM_STUDIO_MODEL` | Model name |
| `-url` | `http://localhost:1234` | LLM API base URL |
| `-verbose` | `false` | Show detailed metrics |

---

## Creating Custom Tools

```go
package tooling

import (
    "context"
    "github.com/andygeiss/go-agent/internal/domain/agent"
)

// Define the tool
func NewMyTool() agent.Tool {
    return agent.Tool{
        ID: "my_tool",
        Definition: agent.NewToolDefinition("my_tool", "Description of what it does").
            WithParameter("input", "The input parameter"),
        Func: MyToolFunc,
    }
}

// Implement the function
func MyToolFunc(ctx context.Context, arguments string) (string, error) {
    var args struct {
        Input string `json:"input"`
    }
    if err := agent.DecodeArgs(arguments, &args); err != nil {
        return "", err
    }
    
    // Your tool logic here
    return "result", nil
}
```

Register the tool:

```go
myTool := tooling.NewMyTool()
executor.RegisterTool("my_tool", myTool.Func)
executor.RegisterToolDefinition(myTool.Definition)
```

---

## Configuration

### Agent Options (alphabetically sorted)

```go
agent.NewAgent("id", "system prompt",
    agent.WithMaxIterations(20),      // Max loop iterations per task
    agent.WithMaxMessages(100),       // Message history limit (0 = unlimited)
    agent.WithMetadata(agent.Metadata{
        "model": "gpt-4",
        "user":  "alice",
    }),
)
```

### LLM Client Options (alphabetically sorted)

```go
client := outbound.NewOpenAIClient(baseURL, model).
    WithCircuitBreaker(10).                     // Open after 10 failures
    WithDebounce(500 * time.Millisecond).       // Coalesce rapid calls
    WithHTTPClient(customClient).               // Custom HTTP client
    WithLLMTimeout(180 * time.Second).          // LLM call timeout
    WithLogger(slog.Default()).                 // Structured logging
    WithRetry(5, 3*time.Second).                // 5 attempts, 3s delay
    WithThrottle(100, 10, time.Second)          // tokens, refill, period
```

### Task Service Options

```go
taskService := agent.NewTaskService(llm, executor, publisher).
    WithHooks(hooks).                 // Lifecycle hooks
    WithParallelToolExecution()       // Enable parallel tool calls
```

---

## Docker

### Build

```bash
docker build -t go-agent .
```

### Run with Docker Compose

```bash
# Create .env file with required variables
echo "APP_SHORTNAME=go-agent" > .env
echo "USER=$(whoami)" >> .env

# Start services
docker-compose up -d
```

---

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run all benchmarks (PGO profiling)
go test -bench=. ./cmd/cli/...

# Run specific benchmark categories
go test -bench=Memory ./cmd/cli/...           # Memory system benchmarks
go test -bench=FullStack ./cmd/cli/...        # End-to-end benchmarks
go test -bench=TaskService ./cmd/cli/...      # Task service benchmarks

# Run benchmarks with custom time
go test -bench=. -benchtime=1s ./cmd/cli/...
```

### Benchmark Categories

The CLI benchmarks (`cmd/cli/main_test.go`) cover all domain contexts:

| Category | Description |
|----------|-------------|
| `Benchmark_FullStack_*` | End-to-end agent with tools |
| `Benchmark_MemoryNote_*` | MemoryNote object creation/methods |
| `Benchmark_MemoryStore_*` | Raw store ops at 100/1000/10000 notes |
| `Benchmark_MemoryTools_*` | Tool-based memory operations |
| `Benchmark_MemorizingService_*` | Complete memory workflow |
| `Benchmark_*NoteUseCase` | Domain use case benchmarks |
| `Benchmark_RealToolExecutor_*` | calculate, time tools |
| `Benchmark_SendMessageUseCase_*` | Chat use case execution |
| `Benchmark_TaskService_*` | Task service with hooks/parallelism |

---

## Project Structure

```
go-agent/
├── cmd/cli/                # CLI application
├── internal/
│   ├── adapters/outbound/  # Infrastructure implementations
│   │   ├── conversation_store.go       # ConversationStore → resource.Access
│   │   ├── encrypted_conversation_store.go # AES-GCM encrypted variant
│   │   ├── event_publisher.go          # EventPublisher → messaging.Dispatcher
│   │   ├── memory_store.go             # MemoryStore → resource.Access
│   │   ├── openai_client.go            # LLMClient → OpenAI-compatible API
│   │   └── tool_executor.go            # ToolExecutor → tool registry
│   └── domain/
│       ├── agent/          # Core domain (Agent, Task, Message, Hooks, Events)
│       ├── chatting/       # Chat use cases (SendMessage, ClearConversation, GetAgentStats)
│       ├── memorizing/     # Memory use cases (WriteNote, GetNote, SearchNotes, DeleteNote)
│       ├── openai/         # OpenAI API types (Request, Response, Tool)
│       └── tooling/        # Tool implementations (calculate, time, memory_tools)
├── AGENTS.md               # AI agent definitions
├── CONTEXT.md              # Architecture documentation
├── Dockerfile
├── docker-compose.yml
├── README.md               # This file
└── VENDOR.md               # Vendor library documentation
```

---

## Contributing

1. Read [CONTEXT.md](CONTEXT.md) for architecture and conventions
2. Check [VENDOR.md](VENDOR.md) for approved vendor patterns
3. Follow the hexagonal architecture pattern
4. Add tests for new functionality
5. Run `go fmt ./...` and `go vet ./...` before committing

---

## License

[MIT License](LICENSE)
