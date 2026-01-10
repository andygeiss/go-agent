package openai_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/openai"
)

// ---------------------------------------------------------------------------
// ChatCompletionChoice tests
// ---------------------------------------------------------------------------

func Test_ChatCompletionChoice_Should_HaveCorrectFields(t *testing.T) {
	// Arrange & Act
	choice := openai.ChatCompletionChoice{
		Index:        0,
		FinishReason: "stop",
		Message:      openai.NewMessage("assistant", "Hello!"),
	}

	// Assert
	assert.That(t, "index must be 0", choice.Index, 0)
	assert.That(t, "finish reason must be stop", choice.FinishReason, "stop")
	assert.That(t, "message content must match", choice.Message.Content, "Hello!")
}

// ---------------------------------------------------------------------------
// ChatCompletionRequest tests
// ---------------------------------------------------------------------------

func Test_NewChatCompletionRequest_Should_SetModelAndMessages(t *testing.T) {
	// Arrange
	messages := []openai.Message{
		openai.NewMessage("system", "You are a helpful assistant."),
		openai.NewMessage("user", "Hello"),
	}

	// Act
	req := openai.NewChatCompletionRequest("gpt-4", messages)

	// Assert
	assert.That(t, "model must be set", req.Model, "gpt-4")
	assert.That(t, "must have 2 messages", len(req.Messages), 2)
}

func Test_NewChatCompletionRequest_Should_SetEmptyTools(t *testing.T) {
	// Arrange
	messages := []openai.Message{openai.NewMessage("user", "Hello")}

	// Act
	req := openai.NewChatCompletionRequest("gpt-4", messages)

	// Assert
	assert.That(t, "tools must be nil", req.Tools == nil, true)
}

func Test_ChatCompletionRequest_WithTools_Should_SetTools(t *testing.T) {
	// Arrange
	messages := []openai.Message{openai.NewMessage("user", "What time is it?")}
	req := openai.NewChatCompletionRequest("gpt-4", messages)
	tools := []openai.Tool{
		openai.NewTool("get_time", "Get the current time"),
	}

	// Act
	req = req.WithTools(tools)

	// Assert
	assert.That(t, "must have 1 tool", len(req.Tools), 1)
	assert.That(t, "tool name must match", req.Tools[0].Function.Name, "get_time")
}

func Test_ChatCompletionRequest_WithTools_Should_ReplaceExistingTools(t *testing.T) {
	// Arrange
	messages := []openai.Message{openai.NewMessage("user", "Hello")}
	req := openai.NewChatCompletionRequest("gpt-4", messages)
	req = req.WithTools([]openai.Tool{openai.NewTool("old_tool", "Old")})
	newTools := []openai.Tool{
		openai.NewTool("new_tool", "New"),
	}

	// Act
	req = req.WithTools(newTools)

	// Assert
	assert.That(t, "must have 1 tool", len(req.Tools), 1)
	assert.That(t, "tool name must be new", req.Tools[0].Function.Name, "new_tool")
}

// ---------------------------------------------------------------------------
// ChatCompletionResponse tests
// ---------------------------------------------------------------------------

func Test_ChatCompletionResponse_GetFirstChoice_With_EmptyChoices_Should_ReturnNil(t *testing.T) {
	// Arrange
	resp := openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{},
	}

	// Act
	choice := resp.GetFirstChoice()

	// Assert
	assert.That(t, "must return nil", choice == nil, true)
}

func Test_ChatCompletionResponse_GetFirstChoice_With_Choices_Should_ReturnFirst(t *testing.T) {
	// Arrange
	resp := openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Index:        0,
				FinishReason: "stop",
				Message:      openai.Message{Role: "assistant", Content: "Hello!"},
			},
			{
				Index:        1,
				FinishReason: "stop",
				Message:      openai.Message{Role: "assistant", Content: "Hi there!"},
			},
		},
	}

	// Act
	choice := resp.GetFirstChoice()

	// Assert
	assert.That(t, "must not return nil", choice != nil, true)
	assert.That(t, "must return first choice", choice.Index, 0)
	assert.That(t, "content must match", choice.Message.Content, "Hello!")
}

func Test_ChatCompletionResponse_GetFirstChoice_With_NilChoices_Should_ReturnNil(t *testing.T) {
	// Arrange
	resp := openai.ChatCompletionResponse{}

	// Act
	choice := resp.GetFirstChoice()

	// Assert
	assert.That(t, "must return nil", choice == nil, true)
}

func Test_ChatCompletionResponse_Should_HaveAllFields(t *testing.T) {
	// Arrange & Act
	resp := openai.ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1677652288,
		Model:   "gpt-4",
		Choices: []openai.ChatCompletionChoice{
			{Index: 0, FinishReason: "stop", Message: openai.NewMessage("assistant", "Hi")},
		},
		Usage: openai.ChatCompletionUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	// Assert
	assert.That(t, "ID must match", resp.ID, "chatcmpl-123")
	assert.That(t, "Object must match", resp.Object, "chat.completion")
	assert.That(t, "Created must match", resp.Created, int64(1677652288))
	assert.That(t, "Model must match", resp.Model, "gpt-4")
	assert.That(t, "must have 1 choice", len(resp.Choices), 1)
	assert.That(t, "total tokens must match", resp.Usage.TotalTokens, 15)
}

// ---------------------------------------------------------------------------
// ChatCompletionUsage tests
// ---------------------------------------------------------------------------

func Test_ChatCompletionUsage_Should_HaveCorrectFields(t *testing.T) {
	// Arrange & Act
	usage := openai.ChatCompletionUsage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
	}

	// Assert
	assert.That(t, "prompt tokens must match", usage.PromptTokens, 100)
	assert.That(t, "completion tokens must match", usage.CompletionTokens, 50)
	assert.That(t, "total tokens must match", usage.TotalTokens, 150)
}

// ---------------------------------------------------------------------------
// Message tests
// ---------------------------------------------------------------------------

func Test_NewMessage_Should_SetRoleAndContent(t *testing.T) {
	// Arrange & Act
	msg := openai.NewMessage("user", "Hello, world!")

	// Assert
	assert.That(t, "role must be user", msg.Role, "user")
	assert.That(t, "content must match", msg.Content, "Hello, world!")
}

func Test_NewMessage_Should_HaveEmptyToolCalls(t *testing.T) {
	// Arrange & Act
	msg := openai.NewMessage("assistant", "Response")

	// Assert
	assert.That(t, "tool calls must be nil", msg.ToolCalls == nil, true)
	assert.That(t, "tool call ID must be empty", msg.ToolCallID, "")
}

func Test_Message_WithToolCallID_Should_SetToolCallID(t *testing.T) {
	// Arrange
	msg := openai.NewMessage("tool", "result")

	// Act
	msg = msg.WithToolCallID("call_123")

	// Assert
	assert.That(t, "tool call ID must be set", msg.ToolCallID, "call_123")
}

func Test_Message_WithToolCallID_Should_PreserveOtherFields(t *testing.T) {
	// Arrange
	msg := openai.NewMessage("tool", "result content")

	// Act
	msg = msg.WithToolCallID("call_abc")

	// Assert
	assert.That(t, "role must be preserved", msg.Role, "tool")
	assert.That(t, "content must be preserved", msg.Content, "result content")
	assert.That(t, "tool call ID must be set", msg.ToolCallID, "call_abc")
}

func Test_Message_WithToolCalls_Should_SetToolCalls(t *testing.T) {
	// Arrange
	msg := openai.NewMessage("assistant", "")
	toolCalls := []openai.ToolCall{
		openai.NewToolCall("call_1", "get_time", "{}"),
		openai.NewToolCall("call_2", "calculate", `{"expression":"2+2"}`),
	}

	// Act
	msg = msg.WithToolCalls(toolCalls)

	// Assert
	assert.That(t, "must have 2 tool calls", len(msg.ToolCalls), 2)
	assert.That(t, "first tool call ID must match", msg.ToolCalls[0].ID, "call_1")
	assert.That(t, "second tool call ID must match", msg.ToolCalls[1].ID, "call_2")
}

func Test_Message_WithToolCalls_Should_PreserveOtherFields(t *testing.T) {
	// Arrange
	msg := openai.NewMessage("assistant", "I'll call some tools")
	toolCalls := []openai.ToolCall{openai.NewToolCall("call_1", "tool", "{}")}

	// Act
	msg = msg.WithToolCalls(toolCalls)

	// Assert
	assert.That(t, "role must be preserved", msg.Role, "assistant")
	assert.That(t, "content must be preserved", msg.Content, "I'll call some tools")
	assert.That(t, "must have 1 tool call", len(msg.ToolCalls), 1)
}

func Test_Message_Chaining_Should_Work(t *testing.T) {
	// Arrange & Act
	msg := openai.NewMessage("tool", "result").
		WithToolCallID("call_123")

	// Assert
	assert.That(t, "role must be tool", msg.Role, "tool")
	assert.That(t, "content must match", msg.Content, "result")
	assert.That(t, "tool call ID must be set", msg.ToolCallID, "call_123")
}
