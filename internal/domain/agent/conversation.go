package agent

import "context"

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
