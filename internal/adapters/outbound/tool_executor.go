package outbound

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/andygeiss/cloud-native-utils/stability"
	"github.com/andygeiss/go-agent/internal/domain/agent"
)

// Default configuration for tool execution (alphabetically sorted).
const (
	defaultToolTimeout = 30 * time.Second // Maximum time for a tool to execute
)

// ToolExecutor implements the agent.ToolExecutor interface.
// It provides tool registration and execution with timeout protection.
// Tool execution is wrapped with timeout to prevent runaway tools.
type ToolExecutor struct {
	logger      *slog.Logger
	tools       map[string]agent.ToolFunc
	definitions []agent.ToolDefinition
	toolTimeout time.Duration
}

// NewToolExecutor creates a new ToolExecutor without any registered tools.
// Use RegisterTool and RegisterToolDefinition to add tools.
// Tool execution is wrapped with a default 30s timeout.
func NewToolExecutor() *ToolExecutor {
	return &ToolExecutor{
		definitions: make([]agent.ToolDefinition, 0),
		tools:       make(map[string]agent.ToolFunc),
		toolTimeout: defaultToolTimeout,
	}
}

// Execute runs the specified tool with the given input arguments.
// Execution is wrapped with a timeout to prevent runaway tools.
func (e *ToolExecutor) Execute(ctx context.Context, toolName string, arguments string) (string, error) {
	fn, ok := e.tools[toolName]
	if !ok {
		if e.logger != nil {
			e.logger.Warn("tool not found", "tool", toolName)
		}
		return "", fmt.Errorf("tool not found: %s", toolName)
	}

	start := time.Now()

	if e.logger != nil {
		e.logger.Debug("tool execution started", "tool", toolName)
	}

	// Wrap the tool function with timeout using stability pattern
	wrappedFn := stability.Timeout(
		func(ctx context.Context, args string) (string, error) {
			return fn(ctx, args)
		},
		e.toolTimeout,
	)

	result, err := wrappedFn(ctx, arguments)

	if e.logger != nil {
		duration := time.Since(start)
		if err != nil {
			e.logger.Error("tool execution failed",
				"tool", toolName,
				"duration", duration,
				"error", err.Error(),
			)
		} else {
			e.logger.Debug("tool execution completed",
				"tool", toolName,
				"duration", duration,
			)
		}
	}

	return result, err
}

// GetAvailableTools returns the list of available tool names.
func (e *ToolExecutor) GetAvailableTools() []string {
	names := make([]string, 0, len(e.tools))
	for name := range e.tools {
		names = append(names, name)
	}
	return names
}

// GetToolDefinitions returns the tool definitions for the LLM.
func (e *ToolExecutor) GetToolDefinitions() []agent.ToolDefinition {
	return e.definitions
}

// HasTool returns true if the specified tool is available.
func (e *ToolExecutor) HasTool(toolName string) bool {
	_, ok := e.tools[toolName]
	return ok
}

// RegisterTool registers a new tool function with the executor.
func (e *ToolExecutor) RegisterTool(name string, fn agent.ToolFunc) {
	e.tools[name] = fn
}

// RegisterToolDefinition registers a tool definition for the LLM.
func (e *ToolExecutor) RegisterToolDefinition(def agent.ToolDefinition) {
	e.definitions = append(e.definitions, def)
}

// WithLogger sets an optional structured logger for the executor.
// When set, the executor logs tool executions at debug level.
func (e *ToolExecutor) WithLogger(logger *slog.Logger) *ToolExecutor {
	e.logger = logger
	return e
}

// WithToolTimeout sets the timeout for tool execution.
func (e *ToolExecutor) WithToolTimeout(timeout time.Duration) *ToolExecutor {
	e.toolTimeout = timeout
	return e
}
