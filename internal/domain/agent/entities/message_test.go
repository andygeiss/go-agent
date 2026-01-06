package entities_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent/entities"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
)

func Test_Message_NewMessage_With_ValidParams_Should_ReturnMessage(t *testing.T) {
	// Arrange
	role := immutable.RoleUser
	content := "Hello, assistant!"

	// Act
	msg := entities.NewMessage(role, content)

	// Assert
	assert.That(t, "message role must match", msg.Role, role)
	assert.That(t, "message content must match", msg.Content, content)
}

func Test_Message_WithToolCalls_With_ToolCalls_Should_HaveToolCalls(t *testing.T) {
	// Arrange
	msg := entities.NewMessage(immutable.RoleAssistant, "content")
	toolCalls := []entities.ToolCall{
		entities.NewToolCall("tc-1", "search", `{}`),
	}

	// Act
	msg = msg.WithToolCalls(toolCalls)

	// Assert
	assert.That(t, "message must have tool calls", len(msg.ToolCalls), 1)
	assert.That(t, "tool call name must match", msg.ToolCalls[0].Name, "search")
}

func Test_Message_WithToolCallID_With_ID_Should_HaveToolCallID(t *testing.T) {
	// Arrange
	msg := entities.NewMessage(immutable.RoleTool, "result")

	// Act
	msg = msg.WithToolCallID("tc-1")

	// Assert
	assert.That(t, "message tool call ID must match", msg.ToolCallID, "tc-1")
}
