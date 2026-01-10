package agent

import (
	"context"

	"github.com/andygeiss/cloud-native-utils/event"
)

// ConversationStore is the interface for persisting conversation history.
// Implementations can use in-memory, JSON file, or database storage.
type ConversationStore interface {
	// Clear removes the conversation history for an agent.
	Clear(ctx context.Context, agentID AgentID) error
	// Load retrieves the conversation history for an agent.
	Load(ctx context.Context, agentID AgentID) ([]Message, error)
	// Save persists the conversation history for an agent.
	Save(ctx context.Context, agentID AgentID, messages []Message) error
}

// EventPublisher is the interface for publishing domain events.
type EventPublisher interface {
	// Publish sends an event to subscribers.
	Publish(ctx context.Context, e event.Event) error
}

// LLMClient is the interface for communicating with a language model.
// Implementations translate between domain types and LLM-specific APIs.
type LLMClient interface {
	// Run sends messages to the LLM and returns its response.
	Run(ctx context.Context, messages []Message, tools []ToolDefinition) (LLMResponse, error)
}

// MemorySearchOptions configures the search behavior.
type MemorySearchOptions struct {
	SessionID     string       // Filter by session ID
	TaskID        string       // Filter by task ID
	UserID        string       // Filter by user ID
	SourceTypes   []SourceType // Filter by source types (any match)
	Tags          []string     // Filter by tags (any match)
	MinImportance int          // Filter by minimum importance (1-5, 0 = no filter)
}

// MemoryStore is the interface for persisting and retrieving memory notes.
// Implementations can use in-memory, JSON file, or database storage with embeddings.
type MemoryStore interface {
	// Delete removes a note by ID.
	Delete(ctx context.Context, id NoteID) error
	// Get retrieves a specific note by ID.
	Get(ctx context.Context, id NoteID) (*MemoryNote, error)
	// Search retrieves notes matching the query and filters.
	Search(ctx context.Context, query string, limit int, opts *MemorySearchOptions) ([]*MemoryNote, error)
	// Write stores a new memory note.
	Write(ctx context.Context, note *MemoryNote) error
}

// TaskRunner executes tasks for an agent.
type TaskRunner interface {
	// RunTask executes a task and returns the result.
	RunTask(ctx context.Context, agent *Agent, task *Task) (Result, error)
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
	// RegisterTool registers a tool function with the given name.
	RegisterTool(name string, fn ToolFunc)
	// RegisterToolDefinition registers a tool definition for the LLM.
	RegisterToolDefinition(def ToolDefinition)
}

// ToolFunc is a function type for tool implementations.
// It receives a context and JSON arguments string, returning a result or error.
type ToolFunc func(ctx context.Context, arguments string) (string, error)
