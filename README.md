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

- 🔄 **Agent Loop Pattern** — Observe → Decide → Act → Update cycle for autonomous task execution
- 🔧 **Tool Use** — Extensible tool system with type-safe definitions
- 🛡️ **Built-in Resilience** — Retry, circuit breaker, timeout, and throttling patterns
- 📡 **Event-Driven** — Observable task lifecycle via domain events
- 🧠 **Memory System** — Long-term context storage with search capabilities
- 🏗️ **Hexagonal Architecture** — Clean separation of domain logic and infrastructure

Works with any OpenAI-compatible API (LM Studio, OpenAI, vLLM, Ollama, etc.).

---

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Features](#features)
- [CLI Usage](#cli-usage)
- [Creating Custom Tools](#creating-custom-tools)
- [Configuration](#configuration)
- [Docker](#docker)
- [Project Structure](#project-structure)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

---

## Installation

```bash
go get github.com/andygeiss/go-agent
```

**Requirements:** Go 1.25+

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

### Built-in Tools

| Tool | Description |
|------|-------------|
| `calculate` | Safe arithmetic expression evaluator |
| `get_current_time` | Returns current date and time |
| Memory tools | Read/write/search long-term memory |

### Domain Events

Subscribe to task lifecycle events:

- `agent.task.started` — Task begins execution
- `agent.task.completed` — Task finishes successfully
- `agent.task.failed` — Task terminates with error
- `agent.toolcall.executed` — Tool call completes

### Lifecycle Hooks

```go
hooks := agent.NewHooks().
    WithBeforeTask(func(ctx context.Context, ag *agent.Agent, t *agent.Task) error {
        log.Println("Starting task:", t.Name)
        return nil
    }).
    WithAfterToolCall(func(ctx context.Context, ag *agent.Agent, tc *agent.ToolCall) error {
        log.Println("Tool executed:", tc.Name, "→", tc.Result)
        return nil
    })

taskService.WithHooks(hooks)
```

### Resilience Patterns

The `OpenAIClient` includes configurable resilience:

- **Timeout**: HTTP (60s) and LLM call (120s) timeouts
- **Retry**: 3 attempts with 2s delay
- **Circuit Breaker**: Opens after 5 consecutive failures
- **Throttling**: Rate limiting (disabled by default)

---

## CLI Usage

```bash
go run ./cmd/cli [flags]
```

### Commands (during chat)

| Command | Description |
|---------|-------------|
| `quit` / `exit` | Exit the CLI |
| `clear` | Reset conversation history |
| `stats` | Show agent statistics |

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-url` | `http://localhost:1234` | LLM API base URL |
| `-model` | `$LM_STUDIO_MODEL` | Model name |
| `-max-iterations` | `10` | Max iterations per task |
| `-max-messages` | `50` | Max messages to retain (0 = unlimited) |
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

### Agent Options

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

### LLM Client Options

```go
client := outbound.NewOpenAIClient(baseURL, model).
    WithLLMTimeout(180 * time.Second).
    WithRetry(5, 3*time.Second).
    WithCircuitBreaker(10).
    WithThrottle(100, 10, time.Second)  // tokens, refill, period
```

### Task Service Options

```go
taskService := agent.NewTaskService(llm, executor, publisher).
    WithHooks(hooks).
    WithParallelToolExecution()  // Enable parallel tool calls
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

# Run benchmarks
go test -bench=. ./internal/domain/agent/
```

---

## Project Structure

```
go-agent/
├── cmd/cli/              # CLI application
├── internal/
│   ├── adapters/outbound/  # Infrastructure implementations
│   └── domain/
│       ├── agent/          # Core domain (Agent, Task, Message)
│       ├── chatting/       # Chat use cases
│       ├── memorizing/     # Memory use cases
│       ├── tooling/        # Tool implementations
│       └── openai/         # OpenAI API types
├── AGENTS.md             # AI agent definitions
├── CONTEXT.md            # Architecture documentation
├── Dockerfile
└── docker-compose.yml
```

---

## Contributing

1. Read [CONTEXT.md](CONTEXT.md) for architecture and conventions
2. Follow the hexagonal architecture pattern
3. Add tests for new functionality
4. Run `go fmt ./...` and `go vet ./...` before committing

---

## License

[MIT License](LICENSE)
