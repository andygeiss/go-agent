package outbound

import (
	"context"
	"encoding/json"

	"github.com/andygeiss/cloud-native-utils/security"
	"github.com/andygeiss/go-agent/internal/domain/agent"
)

// EncryptedConversationStore wraps a ConversationStore to encrypt messages at rest.
// Uses AES-GCM encryption from cloud-native-utils/security.
type EncryptedConversationStore struct {
	store *ConversationStore
	key   [32]byte
}

// NewEncryptedConversationStore creates an encrypted wrapper around a ConversationStore.
// The key should be a 32-byte AES key, typically from security.GenerateKey() or security.Getenv().
func NewEncryptedConversationStore(store *ConversationStore, key [32]byte) *EncryptedConversationStore {
	return &EncryptedConversationStore{
		store: store,
		key:   key,
	}
}

// Save encrypts and persists the conversation history for an agent.
func (s *EncryptedConversationStore) Save(ctx context.Context, agentID agent.AgentID, messages []agent.Message) error {
	// Serialize messages to JSON
	plaintext, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	// Encrypt the serialized messages
	ciphertext := security.Encrypt(plaintext, s.key)

	// Store as a single encrypted message
	encrypted := []agent.Message{
		{
			Role:    agent.RoleSystem,
			Content: string(ciphertext),
		},
	}

	return s.store.Save(ctx, agentID, encrypted)
}

// Load retrieves and decrypts the conversation history for an agent.
// Returns an empty slice if no conversation exists.
func (s *EncryptedConversationStore) Load(ctx context.Context, agentID agent.AgentID) ([]agent.Message, error) {
	// Load the encrypted data
	encrypted, err := s.store.Load(ctx, agentID)
	if err != nil {
		return nil, err
	}

	// Return empty slice if no messages
	if len(encrypted) == 0 {
		return []agent.Message{}, nil
	}

	// Decrypt the stored ciphertext
	ciphertext := []byte(encrypted[0].Content)
	plaintext, err := security.Decrypt(ciphertext, s.key)
	if err != nil {
		return nil, err
	}

	// Deserialize the messages
	var messages []agent.Message
	if err := json.Unmarshal(plaintext, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// Clear removes the conversation history for an agent.
func (s *EncryptedConversationStore) Clear(ctx context.Context, agentID agent.AgentID) error {
	return s.store.Clear(ctx, agentID)
}
