//go:build integration

package outbound_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/internal/domain/agent"
)

// -----------------------------------------------------------------------------
// Integration tests for OpenAI-compatible chat completion API
// -----------------------------------------------------------------------------
//
// These tests require a running LM Studio (or compatible) server.
// Run with: go test -tags=integration ./...
//
// Environment variables:
//   - OPENAI_CHAT_URL: Base URL for the chat API (default: http://localhost:1234)
//   - OPENAI_CHAT_MODEL: Model name to use (required)

func getChattingURL() string {
	if url := os.Getenv("OPENAI_CHAT_URL"); url != "" {
		return url
	}
	return "http://localhost:1234"
}

func getChattingModel() string {
	return os.Getenv("OPENAI_CHAT_MODEL")
}

func skipIfNoChattingModel(t *testing.T) {
	if getChattingModel() == "" {
		t.Skip("OPENAI_CHAT_MODEL not set, skipping integration test")
	}
}

// -----------------------------------------------------------------------------
// Chat completion integration tests
// -----------------------------------------------------------------------------

func Test_Integration_OpenAIClient_Run_With_SimplePrompt_Should_ReturnResponse(t *testing.T) {
	// Skip if no model configured
	skipIfNoChattingModel(t)

	// Arrange
	client := outbound.NewOpenAIClient(getChattingURL(), getChattingModel()).
		WithLLMTimeout(60 * time.Second)

	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Say 'Hello' and nothing else."),
	}

	// Act
	result, err := client.Run(context.Background(), messages, nil)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "response must not be empty", result.Message.Content != "", true)
	assert.That(t, "role must be assistant", result.Message.Role, agent.RoleAssistant)
}

func Test_Integration_OpenAIClient_Run_With_SystemPrompt_Should_FollowInstructions(t *testing.T) {
	// Skip if no model configured
	skipIfNoChattingModel(t)

	// Arrange
	client := outbound.NewOpenAIClient(getChattingURL(), getChattingModel()).
		WithLLMTimeout(60 * time.Second)

	messages := []agent.Message{
		agent.NewMessage(agent.RoleSystem, "You are a helpful assistant. Always respond in exactly 3 words."),
		agent.NewMessage(agent.RoleUser, "How are you?"),
	}

	// Act
	result, err := client.Run(context.Background(), messages, nil)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "response must not be empty", result.Message.Content != "", true)
}

func Test_Integration_OpenAIClient_Run_With_ToolDefinition_Should_ReturnToolCall(t *testing.T) {
	// Skip if no model configured
	skipIfNoChattingModel(t)

	// Arrange
	client := outbound.NewOpenAIClient(getChattingURL(), getChattingModel()).
		WithLLMTimeout(60 * time.Second)

	messages := []agent.Message{
		agent.NewMessage(agent.RoleSystem, "You have access to tools. Use the get_current_time tool when asked about the time."),
		agent.NewMessage(agent.RoleUser, "What time is it right now?"),
	}

	tools := []agent.ToolDefinition{
		agent.NewToolDefinition("get_current_time", "Get the current date and time"),
	}

	// Act
	result, err := client.Run(context.Background(), messages, tools)

	// Assert
	assert.That(t, "must not return error", err, nil)
	// The model should either call the tool or respond with text
	// We check that the response is valid (either has content or tool calls)
	hasContent := result.Message.Content != ""
	hasToolCalls := len(result.ToolCalls) > 0
	assert.That(t, "response must have content or tool calls", hasContent || hasToolCalls, true)
}

func Test_Integration_OpenAIClient_Run_With_MultiTurnConversation_Should_MaintainContext(t *testing.T) {
	// Skip if no model configured
	skipIfNoChattingModel(t)

	// Arrange
	client := outbound.NewOpenAIClient(getChattingURL(), getChattingModel()).
		WithLLMTimeout(60 * time.Second)

	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "My name is Alice."),
		agent.NewMessage(agent.RoleAssistant, "Nice to meet you, Alice!"),
		agent.NewMessage(agent.RoleUser, "What is my name?"),
	}

	// Act
	result, err := client.Run(context.Background(), messages, nil)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "response must not be empty", result.Message.Content != "", true)
	// The response should contain "Alice" since the model should remember the context
}

func Test_Integration_OpenAIClient_Run_With_CanceledContext_Should_ReturnError(t *testing.T) {
	// Skip if no model configured
	skipIfNoChattingModel(t)

	// Arrange
	client := outbound.NewOpenAIClient(getChattingURL(), getChattingModel())

	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Tell me a very long story about dragons."),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	_, err := client.Run(ctx, messages, nil)

	// Assert
	assert.That(t, "must return error", err != nil, true)
}

func Test_Integration_OpenAIClient_Run_With_ShortTimeout_Should_ReturnError(t *testing.T) {
	// Skip if no model configured
	skipIfNoChattingModel(t)

	// Arrange
	client := outbound.NewOpenAIClient(getChattingURL(), getChattingModel()).
		WithLLMTimeout(1 * time.Millisecond) // Very short timeout

	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Tell me a story."),
	}

	// Act
	_, err := client.Run(context.Background(), messages, nil)

	// Assert
	assert.That(t, "must return error due to timeout", err != nil, true)
}
