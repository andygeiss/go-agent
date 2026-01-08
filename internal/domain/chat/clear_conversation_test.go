package chat_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/chat"
	"github.com/andygeiss/go-agent/pkg/agent"
)

func Test_ClearConversationUseCase_Execute_Should_ClearMessages(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("test-agent", "test prompt")
	ag.AddMessage(agent.NewMessage(agent.RoleUser, "Hello"))
	ag.AddMessage(agent.NewMessage(agent.RoleAssistant, "Hi!"))
	uc := chat.NewClearConversationUseCase(&ag)

	// Act
	uc.Execute()

	// Assert
	assert.That(t, "message count must be 0", ag.MessageCount(), 0)
}

func Test_ClearConversationUseCase_Execute_With_EmptyHistory_Should_NotPanic(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("test-agent", "test prompt")
	uc := chat.NewClearConversationUseCase(&ag)

	// Act & Assert (no panic)
	uc.Execute()
	assert.That(t, "message count must be 0", ag.MessageCount(), 0)
}
