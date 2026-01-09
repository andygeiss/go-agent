package agent_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent"
)

// Message tests

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

func Test_Message_WithToolCallID_Should_SetToolCallID(t *testing.T) {
	// Arrange
	msg := agent.NewMessage(agent.RoleTool, "result")

	// Act
	msg = msg.WithToolCallID("tc-1")

	// Assert
	assert.That(t, "tool call ID must match", msg.ToolCallID, agent.ToolCallID("tc-1"))
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

// ToolCall tests

func Test_ToolCall_Complete_With_Result_Should_BeCompleted(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)
	tc.Execute()

	// Act
	tc.Complete("search result")

	// Assert
	assert.That(t, "tool call status must be completed", tc.Status, agent.ToolCallStatusCompleted)
	assert.That(t, "tool call result must match", tc.Result, "search result")
}

func Test_ToolCall_Execute_With_PendingToolCall_Should_BeExecuting(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)

	// Act
	tc.Execute()

	// Assert
	assert.That(t, "tool call status must be executing", tc.Status, agent.ToolCallStatusExecuting)
}

func Test_ToolCall_Fail_With_Error_Should_BeFailed(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)
	tc.Execute()

	// Act
	tc.Fail("tool error")

	// Assert
	assert.That(t, "tool call status must be failed", tc.Status, agent.ToolCallStatusFailed)
	assert.That(t, "tool call error must match", tc.Error, "tool error")
}

func Test_ToolCall_NewToolCall_With_ValidParams_Should_ReturnPendingToolCall(t *testing.T) {
	// Arrange
	id := agent.ToolCallID("tc-1")
	name := "search"
	args := `{"query": "test"}`

	// Act
	tc := agent.NewToolCall(id, name, args)

	// Assert
	assert.That(t, "tool call ID must match", tc.ID, id)
	assert.That(t, "tool call name must match", tc.Name, name)
	assert.That(t, "tool call arguments must match", tc.Arguments, args)
	assert.That(t, "tool call status must be pending", tc.Status, agent.ToolCallStatusPending)
}

func Test_ToolCall_ToMessage_With_CompletedToolCall_Should_ReturnToolMessage(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)
	tc.Complete("result")

	// Act
	msg := tc.ToMessage()

	// Assert
	assert.That(t, "message role must be tool", msg.Role, agent.RoleTool)
	assert.That(t, "message content must be result", msg.Content, "result")
	assert.That(t, "message tool call ID must match", msg.ToolCallID, agent.ToolCallID("tc-1"))
}

func Test_ToolCall_ToMessage_With_FailedToolCall_Should_ReturnErrorMessage(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)
	tc.Fail("tool failed")

	// Act
	msg := tc.ToMessage()

	// Assert
	assert.That(t, "message role must be tool", msg.Role, agent.RoleTool)
	assert.That(t, "message content must contain error", msg.Content, "Error: tool failed")
}
