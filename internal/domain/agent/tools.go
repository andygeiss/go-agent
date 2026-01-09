package agent

import "context"

// ToolExecutor is the interface for executing tools requested by the LLM.
// It manages tool registration and execution.
type ToolExecutor interface {
	// Execute runs a tool with the given name and arguments.
	Execute(ctx context.Context, toolName string, arguments string) (string, error)
	// GetAvailableTools returns the list of registered tool names.
	GetAvailableTools() []string
	// GetToolDefinitions returns tool definitions for the LLM.
	GetToolDefinitions() []ToolDefinition
	// HasTool checks if a tool with the given name exists.
	HasTool(toolName string) bool
	// RegisterTool registers a tool function with the given name.
	RegisterTool(name string, fn ToolFunc)
	// RegisterToolDefinition registers a tool definition for the LLM.
	RegisterToolDefinition(def ToolDefinition)
}

// ToolFunc is a function type for tool implementations.
// It receives a context and JSON arguments string, returning a result or error.
type ToolFunc func(ctx context.Context, arguments string) (string, error)

// TaskRunner executes tasks for an agent.
type TaskRunner interface {
	// RunTask executes a task and returns the result.
	RunTask(ctx context.Context, agent *Agent, task *Task) (Result, error)
}
