package agent_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/pkg/agent"
)

func Test_Message_NewMessage_With_ValidParams_Should_ReturnMessage(t *testing.T) {
	// Arrange
	role := agent.RoleUser
	content := "Hello, world!"

	// Act
	msg := agent.NewMessage(role, content)

	// Assert
	assert.That(t, "message role must match", msg.Role, role)
	assert.That(t, "message content must match", msg.Content, content)
}

func Test_Message_WithToolCalls_Should_AttachToolCalls(t *testing.T) {
	// Arrange
	msg := agent.NewMessage(agent.RoleAssistant, "")
	tc := agent.NewToolCall("tc-1", "search", `{"query":"test"}`)

	// Act
	msg = msg.WithToolCalls([]agent.ToolCall{tc})

	// Assert
	assert.That(t, "message must have one tool call", len(msg.ToolCalls), 1)
	assert.That(t, "tool call name must match", msg.ToolCalls[0].Name, "search")
}

func Test_Message_WithToolCallID_Should_SetToolCallID(t *testing.T) {
	// Arrange
	msg := agent.NewMessage(agent.RoleTool, "result")

	// Act
	msg = msg.WithToolCallID("tc-1")

	// Assert
	assert.That(t, "tool call ID must match", msg.ToolCallID, agent.ToolCallID("tc-1"))
}
