package outbound

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/andygeiss/cloud-native-utils/stability"
	"github.com/andygeiss/go-agent/pkg/agent"
)

// Default configuration for tool execution.
const (
	defaultToolTimeout = 30 * time.Second // Maximum time for a tool to execute
)

// ToolExecutor implements the agent.ToolExecutor interface.
// It provides a minimal set of demo tools for testing the agent loop.
// Tool execution is wrapped with timeout to prevent runaway tools.
type ToolExecutor struct {
	tools       map[string]ToolFunc
	toolTimeout time.Duration
}

// ToolFunc is a function type for tool implementations.
type ToolFunc func(ctx context.Context, arguments string) (string, error)

// NewToolExecutor creates a new ToolExecutor with demo tools.
// Tool execution is wrapped with a default 30s timeout.
func NewToolExecutor() *ToolExecutor {
	executor := &ToolExecutor{
		tools:       make(map[string]ToolFunc),
		toolTimeout: defaultToolTimeout,
	}

	// Register demo tools
	executor.RegisterTool("get_current_time", executor.getCurrentTime)
	executor.RegisterTool("calculate", executor.calculate)

	return executor
}

// WithToolTimeout sets the timeout for tool execution.
func (e *ToolExecutor) WithToolTimeout(timeout time.Duration) *ToolExecutor {
	e.toolTimeout = timeout
	return e
}

// RegisterTool registers a new tool with the executor.
func (e *ToolExecutor) RegisterTool(name string, fn ToolFunc) {
	e.tools[name] = fn
}

// Execute runs the specified tool with the given input arguments.
// Execution is wrapped with a timeout to prevent runaway tools.
func (e *ToolExecutor) Execute(ctx context.Context, toolName string, arguments string) (string, error) {
	fn, ok := e.tools[toolName]
	if !ok {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}

	// Wrap the tool function with timeout using stability pattern
	wrappedFn := stability.Timeout(
		func(ctx context.Context, args string) (string, error) {
			return fn(ctx, args)
		},
		e.toolTimeout,
	)

	return wrappedFn(ctx, arguments)
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
	return []agent.ToolDefinition{
		agent.NewToolDefinition("get_current_time", "Get the current date and time"),
		agent.NewToolDefinition("calculate", "Perform a simple arithmetic calculation").
			WithParameter("expression", "The arithmetic expression to evaluate (e.g., '2 + 2')"),
	}
}

// HasTool returns true if the specified tool is available.
func (e *ToolExecutor) HasTool(toolName string) bool {
	_, ok := e.tools[toolName]
	return ok
}

// getCurrentTime returns the current date and time.
func (e *ToolExecutor) getCurrentTime(_ context.Context, _ string) (string, error) {
	return time.Now().Format(time.RFC3339), nil
}

// calculateArgs represents the arguments for the calculate tool.
type calculateArgs struct {
	Expression string `json:"expression"`
}

// calculate performs a simple arithmetic calculation.
func (e *ToolExecutor) calculate(_ context.Context, arguments string) (string, error) {
	var args calculateArgs
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// For demo purposes, we just echo the expression
	// A real implementation would parse and evaluate the expression
	return fmt.Sprintf("Result of '%s' = (expression evaluation not implemented in demo)", args.Expression), nil
}
