package outbound

import (
	"context"

	"github.com/andygeiss/cloud-native-utils/resource"
	"github.com/andygeiss/go-agent/pkg/agent"
)

// Conversation represents the persisted conversation data for an agent.
type Conversation struct {
	AgentID  string          `json:"agent_id"`
	Messages []agent.Message `json:"messages"`
}

// ConversationStore persists conversation history using a generic resource.Access backend.
// Supports any backend: InMemoryAccess, JsonFileAccess, YamlFileAccess, SqliteAccess.
type ConversationStore struct {
	access resource.Access[string, Conversation]
}

// NewConversationStore creates a ConversationStore with the given storage backend.
// Example backends:
//   - resource.NewInMemoryAccess[string, Conversation]() - for testing
//   - resource.NewJsonFileAccess[string, Conversation]("conversations.json") - for file persistence
func NewConversationStore(access resource.Access[string, Conversation]) *ConversationStore {
	return &ConversationStore{access: access}
}

// NewInMemoryConversationStore creates a ConversationStore backed by in-memory storage.
// Useful for testing or when persistence is not required.
func NewInMemoryConversationStore() *ConversationStore {
	return NewConversationStore(resource.NewInMemoryAccess[string, Conversation]())
}

// NewJsonFileConversationStore creates a ConversationStore backed by a JSON file.
// The file is created if it does not exist.
func NewJsonFileConversationStore(path string) *ConversationStore {
	return NewConversationStore(resource.NewJsonFileAccess[string, Conversation](path))
}

// Save persists the conversation history for an agent.
// Creates a new record if none exists, or updates the existing one.
func (s *ConversationStore) Save(ctx context.Context, agentID agent.AgentID, messages []agent.Message) error {
	key := string(agentID)
	conversation := Conversation{
		AgentID:  key,
		Messages: messages,
	}

	// Try to create new conversation first (handles non-existent files)
	err := s.access.Create(ctx, key, conversation)
	if err != nil && err.Error() == resource.ErrorResourceAlreadyExists {
		// Update existing conversation
		return s.access.Update(ctx, key, conversation)
	}
	return err
}

// Load retrieves the conversation history for an agent.
// Returns an empty slice if no conversation exists.
func (s *ConversationStore) Load(ctx context.Context, agentID agent.AgentID) ([]agent.Message, error) {
	key := string(agentID)
	conversation, err := s.access.Read(ctx, key)
	if err != nil {
		if err.Error() == resource.ErrorResourceNotFound {
			return []agent.Message{}, nil
		}
		return nil, err
	}
	return conversation.Messages, nil
}

// Clear removes the conversation history for an agent.
// Returns nil if the conversation does not exist.
func (s *ConversationStore) Clear(ctx context.Context, agentID agent.AgentID) error {
	key := string(agentID)
	err := s.access.Delete(ctx, key)
	if err != nil && err.Error() == resource.ErrorResourceNotFound {
		return nil // Not an error if conversation doesn't exist
	}
	return err
}
