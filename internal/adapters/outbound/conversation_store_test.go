package outbound_test

import (
	"context"
	"os"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/pkg/agent"
)

func Test_ConversationStore_Save_Should_PersistMessages(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryConversationStore()
	ctx := context.Background()
	agentID := agent.AgentID("test-agent")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
		agent.NewMessage(agent.RoleAssistant, "Hi there!"),
	}

	// Act
	err := store.Save(ctx, agentID, messages)

	// Assert
	assert.That(t, "err must be nil", err, nil)
}

func Test_ConversationStore_Load_Should_ReturnSavedMessages(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryConversationStore()
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

func Test_ConversationStore_Load_Should_ReturnEmptySliceForNonExistent(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryConversationStore()
	ctx := context.Background()
	agentID := agent.AgentID("non-existent")

	// Act
	loaded, err := store.Load(ctx, agentID)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "loaded must be empty", len(loaded), 0)
}

func Test_ConversationStore_Clear_Should_RemoveMessages(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryConversationStore()
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

func Test_ConversationStore_Clear_Should_NotErrorForNonExistent(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryConversationStore()
	ctx := context.Background()
	agentID := agent.AgentID("non-existent")

	// Act
	err := store.Clear(ctx, agentID)

	// Assert
	assert.That(t, "err must be nil for non-existent", err, nil)
}

func Test_ConversationStore_Save_Should_UpdateExistingConversation(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryConversationStore()
	ctx := context.Background()
	agentID := agent.AgentID("test-agent")
	initial := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
	}
	_ = store.Save(ctx, agentID, initial)

	updated := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
		agent.NewMessage(agent.RoleAssistant, "Hi!"),
		agent.NewMessage(agent.RoleUser, "How are you?"),
	}

	// Act
	err := store.Save(ctx, agentID, updated)
	loaded, _ := store.Load(ctx, agentID)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "loaded must have 3 messages", len(loaded), 3)
	assert.That(t, "last message content", loaded[2].Content, "How are you?")
}

func Test_ConversationStore_JsonFile_Should_PersistToDisk(t *testing.T) {
	// Arrange
	path := "./test_conversations.json"
	t.Cleanup(func() { _ = os.Remove(path) })
	store := outbound.NewJsonFileConversationStore(path)
	ctx := context.Background()
	agentID := agent.AgentID("test-agent")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
		agent.NewMessage(agent.RoleAssistant, "Hi there!"),
	}

	// Act
	err := store.Save(ctx, agentID, messages)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	_, statErr := os.Stat(path)
	assert.That(t, "file must exist", statErr, nil)
}

func Test_ConversationStore_JsonFile_Should_SurviveReload(t *testing.T) {
	// Arrange
	path := "./test_conversations_reload.json"
	t.Cleanup(func() { _ = os.Remove(path) })
	store1 := outbound.NewJsonFileConversationStore(path)
	ctx := context.Background()
	agentID := agent.AgentID("test-agent")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
		agent.NewMessage(agent.RoleAssistant, "Hi there!"),
	}
	_ = store1.Save(ctx, agentID, messages)

	// Create new store instance (simulates app restart)
	store2 := outbound.NewJsonFileConversationStore(path)

	// Act
	loaded, err := store2.Load(ctx, agentID)

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "loaded must have 2 messages", len(loaded), 2)
	assert.That(t, "first message content", loaded[0].Content, "Hello")
}

func Test_ConversationStore_MultipleAgents_Should_BeSeparate(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryConversationStore()
	ctx := context.Background()
	agent1 := agent.AgentID("agent-1")
	agent2 := agent.AgentID("agent-2")
	messages1 := []agent.Message{agent.NewMessage(agent.RoleUser, "Hello from agent 1")}
	messages2 := []agent.Message{agent.NewMessage(agent.RoleUser, "Hello from agent 2")}
	_ = store.Save(ctx, agent1, messages1)
	_ = store.Save(ctx, agent2, messages2)

	// Act
	loaded1, _ := store.Load(ctx, agent1)
	loaded2, _ := store.Load(ctx, agent2)

	// Assert
	assert.That(t, "agent 1 must have 1 message", len(loaded1), 1)
	assert.That(t, "agent 2 must have 1 message", len(loaded2), 1)
	assert.That(t, "agent 1 content", loaded1[0].Content, "Hello from agent 1")
	assert.That(t, "agent 2 content", loaded2[0].Content, "Hello from agent 2")
}
