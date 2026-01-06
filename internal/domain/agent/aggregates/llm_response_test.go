package aggregates_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent/aggregates"
	"github.com/andygeiss/go-agent/internal/domain/agent/entities"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
)

func Test_LLMResponse_NewLLMResponse_With_ValidParams_Should_ReturnResponse(t *testing.T) {
	// Arrange
	msg := entities.NewMessage(immutable.RoleAssistant, "Hello!")
	finishReason := "stop"

	// Act
	resp := aggregates.NewLLMResponse(msg, finishReason)

	// Assert
	assert.That(t, "response message content must match", resp.Message.Content, "Hello!")
	assert.That(t, "response finish reason must match", resp.FinishReason, "stop")
}

func Test_LLMResponse_HasToolCalls_With_NoToolCalls_Should_ReturnFalse(t *testing.T) {
	// Arrange
	msg := entities.NewMessage(immutable.RoleAssistant, "Hello!")
	resp := aggregates.NewLLMResponse(msg, "stop")

	// Act
	hasToolCalls := resp.HasToolCalls()

	// Assert
	assert.That(t, "response must not have tool calls", hasToolCalls, false)
}

func Test_LLMResponse_HasToolCalls_With_ToolCalls_Should_ReturnTrue(t *testing.T) {
	// Arrange
	msg := entities.NewMessage(immutable.RoleAssistant, "")
	toolCalls := []entities.ToolCall{
		entities.NewToolCall("tc-1", "search", `{}`),
	}
	resp := aggregates.NewLLMResponse(msg, "tool_calls").WithToolCalls(toolCalls)

	// Act
	hasToolCalls := resp.HasToolCalls()

	// Assert
	assert.That(t, "response must have tool calls", hasToolCalls, true)
}
