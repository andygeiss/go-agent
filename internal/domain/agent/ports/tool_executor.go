package ports

import (
	"context"

	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
)

// ToolExecutor represents the interface for executing tool calls.
// Tools are functions that the agent can use to interact with the outside world.
// The adapter layer implements this interface with actual tool implementations.
type ToolExecutor interface {
	// Execute runs the specified tool with the given input arguments.
	// It returns the result of the tool execution as a string.
	Execute(ctx context.Context, toolName string, arguments string) (string, error)

	// GetAvailableTools returns the list of available tool names.
	GetAvailableTools() []string

	// GetToolDefinitions returns the tool definitions for the LLM.
	// These definitions tell the LLM what tools are available and how to use them.
	GetToolDefinitions() []immutable.ToolDefinition

	// HasTool returns true if the specified tool is available.
	HasTool(toolName string) bool
}
