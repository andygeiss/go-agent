package outbound_test

import (
	"context"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/security"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/pkg/agent"
)

func Test_EncryptedConversationStore_Save_Should_EncryptMessages(t *testing.T) {
	// Arrange
	baseStore := outbound.NewInMemoryConversationStore()
	key := security.GenerateKey()
	store := outbound.NewEncryptedConversationStore(baseStore, key)
	ctx := context.Background()
	agentID := agent.AgentID("test-agent")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Secret message"),
	}

	// Act
	err := store.Save(ctx, agentID, messages)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	// Verify the base store has encrypted data
	raw, _ := baseStore.Load(ctx, agentID)
	assert.That(t, "raw must have 1 message", len(raw), 1)
	// The content should not be readable plaintext
	assert.That(t, "content must not be plaintext", raw[0].Content != "Secret message", true)
}

func Test_EncryptedConversationStore_Load_Should_DecryptMessages(t *testing.T) {
	// Arrange
	baseStore := outbound.NewInMemoryConversationStore()
	key := security.GenerateKey()
	store := outbound.NewEncryptedConversationStore(baseStore, key)
	ctx := context.Background()
	agentID := agent.AgentID("test-agent")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
		agent.NewMessage(agent.RoleAssistant, "Hi there!"),
	}
	_ = store.Save(ctx, agentID, messages)

	// Act
	loaded, err := store.Load(ctx, agentID)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "loaded must have 2 messages", len(loaded), 2)
	assert.That(t, "first message content", loaded[0].Content, "Hello")
	assert.That(t, "second message content", loaded[1].Content, "Hi there!")
}

func Test_EncryptedConversationStore_Load_Should_ReturnEmptyForNonExistent(t *testing.T) {
	// Arrange
	baseStore := outbound.NewInMemoryConversationStore()
	key := security.GenerateKey()
	store := outbound.NewEncryptedConversationStore(baseStore, key)
	ctx := context.Background()
	agentID := agent.AgentID("non-existent")

	// Act
	loaded, err := store.Load(ctx, agentID)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "loaded must be empty", len(loaded), 0)
}

func Test_EncryptedConversationStore_Load_Should_FailWithWrongKey(t *testing.T) {
	// Arrange
	baseStore := outbound.NewInMemoryConversationStore()
	key1 := security.GenerateKey()
	key2 := security.GenerateKey()
	store1 := outbound.NewEncryptedConversationStore(baseStore, key1)
	store2 := outbound.NewEncryptedConversationStore(baseStore, key2)
	ctx := context.Background()
	agentID := agent.AgentID("test-agent")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Secret"),
	}
	_ = store1.Save(ctx, agentID, messages)

	// Act
	_, err := store2.Load(ctx, agentID)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
}

func Test_EncryptedConversationStore_Clear_Should_RemoveMessages(t *testing.T) {
	// Arrange
	baseStore := outbound.NewInMemoryConversationStore()
	key := security.GenerateKey()
	store := outbound.NewEncryptedConversationStore(baseStore, key)
	ctx := context.Background()
	agentID := agent.AgentID("test-agent")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
	}
	_ = store.Save(ctx, agentID, messages)

	// Act
	err := store.Clear(ctx, agentID)
	loaded, _ := store.Load(ctx, agentID)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "loaded must be empty after clear", len(loaded), 0)
}

func Test_EncryptedConversationStore_ToolCalls_Should_Preserve(t *testing.T) {
	// Arrange
	baseStore := outbound.NewInMemoryConversationStore()
	key := security.GenerateKey()
	store := outbound.NewEncryptedConversationStore(baseStore, key)
	ctx := context.Background()
	agentID := agent.AgentID("test-agent")
	toolCall := agent.ToolCall{
		ID:        "call-1",
		Name:      "get_time",
		Arguments: `{"timezone": "UTC"}`,
	}
	messages := []agent.Message{
		agent.NewMessage(agent.RoleAssistant, "").WithToolCalls([]agent.ToolCall{toolCall}),
		agent.NewMessage(agent.RoleTool, "2024-01-01T00:00:00Z").WithToolCallID("call-1"),
	}
	_ = store.Save(ctx, agentID, messages)

	// Act
	loaded, err := store.Load(ctx, agentID)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "loaded must have 2 messages", len(loaded), 2)
	assert.That(t, "first message role", string(loaded[0].Role), string(agent.RoleAssistant))
	assert.That(t, "tool call preserved", len(loaded[0].ToolCalls), 1)
	assert.That(t, "tool call name", loaded[0].ToolCalls[0].Name, "get_time")
	assert.That(t, "second message tool call ID", string(loaded[1].ToolCallID), "call-1")
}
