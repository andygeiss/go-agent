package agent

import (
	"context"

	"github.com/andygeiss/go-agent/pkg/event"
)

// LLMClient is the interface for communicating with a language model.
// Implementations translate between domain types and LLM-specific APIs.
type LLMClient interface {
	// Run sends messages to the LLM and returns its response.
	Run(ctx context.Context, messages []Message, tools []ToolDefinition) (LLMResponse, error)
}

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
}

// EventPublisher is the interface for publishing domain events.
type EventPublisher interface {
	// Publish sends an event to subscribers.
	Publish(ctx context.Context, e event.Event) error
}

// ConversationStore is the interface for persisting conversation history.
// Implementations can use in-memory, JSON file, or database storage.
type ConversationStore interface {
	// Save persists the conversation history for an agent.
	Save(ctx context.Context, agentID AgentID, messages []Message) error
	// Load retrieves the conversation history for an agent.
	Load(ctx context.Context, agentID AgentID) ([]Message, error)
	// Clear removes the conversation history for an agent.
	Clear(ctx context.Context, agentID AgentID) error
}
