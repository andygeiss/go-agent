package outbound

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/andygeiss/go-agent/pkg/agent"
)

// ToolExecutor implements the agent.ToolExecutor interface.
// It provides a minimal set of demo tools for testing the agent loop.
type ToolExecutor struct {
	tools map[string]ToolFunc
}

// ToolFunc is a function type for tool implementations.
type ToolFunc func(ctx context.Context, arguments string) (string, error)

// NewToolExecutor creates a new SimpleToolExecutor with demo tools.
func NewToolExecutor() *ToolExecutor {
	executor := &ToolExecutor{
		tools: make(map[string]ToolFunc),
	}

	// Register demo tools
	executor.RegisterTool("get_current_time", executor.getCurrentTime)
	executor.RegisterTool("calculate", executor.calculate)

	return executor
}

// RegisterTool registers a new tool with the executor.
func (e *ToolExecutor) RegisterTool(name string, fn ToolFunc) {
	e.tools[name] = fn
}

// Execute runs the specified tool with the given input arguments.
func (e *ToolExecutor) Execute(ctx context.Context, toolName string, arguments string) (string, error) {
	fn, ok := e.tools[toolName]
	if !ok {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}
	return fn(ctx, arguments)
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
