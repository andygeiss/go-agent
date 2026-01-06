<p align="center">
<img src="https://github.com/andygeiss/go-ddd-hex-starter/blob/main/cmd/server/assets/static/img/icon-192.png?raw=true" width="100"/>
</p>

# Go Agent

[![Go Reference](https://pkg.go.dev/badge/github.com/andygeiss/go-agent.svg)](https://pkg.go.dev/github.com/andygeiss/go-agent)
[![License](https://img.shields.io/github/license/andygeiss/go-agent)](https://github.com/andygeiss/go-agent/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/v/release/andygeiss/go-agent)](https://github.com/andygeiss/go-agent/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygeiss/go-agent)](https://goreportcard.com/report/github.com/andygeiss/go-agent)

A Go-based AI Agent library implementing the **Observe → Decide → Act → Update** loop pattern to interact with Large Language Models (LLMs) and execute tools.

## Features

- **Reusable Library** - Import `pkg/agent` to build LLM-powered applications
- **LLM Integration** - OpenAI-compatible API support (works with LM Studio, OpenAI, etc.)
- **Tool Calling** - Extensible tool system for agent capabilities
- **Event-Driven** - Domain events for observability and extensibility

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

Type your message and press Enter. Type 'quit' or 'exit' to stop.

You: What time is it?
🤖 Assistant: The current time is 2026-01-06T15:30:45Z.

You: quit
Goodbye! 👋
```

## Project Structure

```
go-agent/
├── cmd/cli/                    # CLI application entry point
├── internal/
│   └── adapters/outbound/      # Infrastructure adapters (LLM, tools, events)
├── pkg/
│   ├── agent/                  # Reusable agent library
│   │   ├── types.go            # ID types, Role, Status constants
│   │   ├── agent.go            # Agent aggregate
│   │   ├── task.go             # Task entity
│   │   ├── message.go          # Conversation messages
│   │   ├── tool_call.go        # Tool call entity
│   │   ├── ports.go            # Interfaces (LLMClient, ToolExecutor)
│   │   ├── task_service.go     # Agent loop orchestration
│   │   └── events/             # Domain events
│   ├── event/                  # Event interfaces
│   └── openai/                 # OpenAI API structures
└── tools/                      # Development scripts
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

Command-line flags are also available:

```bash
just run -- -url http://localhost:1234 -model your-model
```

## Architecture

The project provides a reusable agent library in `pkg/agent/`:

```
┌─────────────────────────────────────────────────────────────┐
│                      Application (CLI)                       │
└─────────────────────────────┬───────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────┐
│                     pkg/agent Library                        │
│  • Agent, Task, Message     • TaskService (Agent Loop)      │
│  • LLMClient interface      • ToolExecutor interface        │
└─────────────────────────────┬───────────────────────────────┘
                              │ implements
┌─────────────────────────────▼───────────────────────────────┐
│                    Adapter Layer                             │
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
ag := agent.NewAgent("my-agent", "You are a helpful assistant")

// Run a task
task := agent.NewTask("task-1", "chat", "Hello!")
result, err := taskService.RunTask(ctx, &ag, task)
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

## Testing

```bash
# Run all unit tests
just test

# Run with verbose output
go test -v ./internal/...

# Run integration tests (requires LM Studio running)
just test-integration
```

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

## Acknowledgments

- Built with [Go](https://go.dev)
- Architecture inspired by [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/) and [Domain-Driven Design](https://www.domainlanguage.com/ddd/)
