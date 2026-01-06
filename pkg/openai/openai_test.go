package openai_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/pkg/openai"
)

// -----------------------------------------------------------------------------
// Message tests
// -----------------------------------------------------------------------------

func Test_NewMessage_Should_SetRoleAndContent(t *testing.T) {
	// Arrange & Act
	msg := openai.NewMessage("user", "Hello, world!")

	// Assert
	assert.That(t, "role must be user", msg.Role, "user")
	assert.That(t, "content must match", msg.Content, "Hello, world!")
}

func Test_Message_WithToolCallID_Should_SetToolCallID(t *testing.T) {
	// Arrange
	msg := openai.NewMessage("tool", "result")

	// Act
	msg = msg.WithToolCallID("call_123")

	// Assert
	assert.That(t, "tool call ID must be set", msg.ToolCallID, "call_123")
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

// -----------------------------------------------------------------------------
// ChatCompletionRequest tests
// -----------------------------------------------------------------------------

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

// -----------------------------------------------------------------------------
// ChatCompletionResponse tests
// -----------------------------------------------------------------------------

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

// -----------------------------------------------------------------------------
// Tool tests
// -----------------------------------------------------------------------------

func Test_NewTool_Should_CreateFunctionTool(t *testing.T) {
	// Arrange & Act
	tool := openai.NewTool("get_weather", "Get the current weather")

	// Assert
	assert.That(t, "type must be function", tool.Type, "function")
	assert.That(t, "name must match", tool.Function.Name, "get_weather")
	assert.That(t, "description must match", tool.Function.Description, "Get the current weather")
	assert.That(t, "parameters type must be object", tool.Function.Parameters.Type, "object")
}

func Test_Tool_WithParameter_Should_AddRequiredParameter(t *testing.T) {
	// Arrange
	tool := openai.NewTool("get_weather", "Get the current weather")

	// Act
	tool = tool.WithParameter("location", "string", "The city name", true)

	// Assert
	assert.That(t, "must have 1 property", len(tool.Function.Parameters.Properties), 1)
	assert.That(t, "property type must be string", tool.Function.Parameters.Properties["location"].Type, "string")
	assert.That(t, "must have 1 required field", len(tool.Function.Parameters.Required), 1)
	assert.That(t, "required field must match", tool.Function.Parameters.Required[0], "location")
}

func Test_Tool_WithParameter_Should_AddOptionalParameter(t *testing.T) {
	// Arrange
	tool := openai.NewTool("get_weather", "Get the current weather")

	// Act
	tool = tool.WithParameter("units", "string", "Temperature units (celsius/fahrenheit)", false)

	// Assert
	assert.That(t, "must have 1 property", len(tool.Function.Parameters.Properties), 1)
	assert.That(t, "property type must be string", tool.Function.Parameters.Properties["units"].Type, "string")
	assert.That(t, "must have 0 required fields", len(tool.Function.Parameters.Required), 0)
}

func Test_Tool_WithParameter_Should_AddMultipleParameters(t *testing.T) {
	// Arrange
	tool := openai.NewTool("get_weather", "Get the current weather")

	// Act
	tool = tool.
		WithParameter("location", "string", "The city name", true).
		WithParameter("units", "string", "Temperature units", false)

	// Assert
	assert.That(t, "must have 2 properties", len(tool.Function.Parameters.Properties), 2)
	assert.That(t, "must have 1 required field", len(tool.Function.Parameters.Required), 1)
}

// -----------------------------------------------------------------------------
// ToolCall tests
// -----------------------------------------------------------------------------

func Test_NewToolCall_Should_CreateToolCall(t *testing.T) {
	// Arrange & Act
	tc := openai.NewToolCall("call_abc123", "get_weather", `{"location":"London"}`)

	// Assert
	assert.That(t, "ID must match", tc.ID, "call_abc123")
	assert.That(t, "type must be function", tc.Type, "function")
	assert.That(t, "function name must match", tc.Function.Name, "get_weather")
	assert.That(t, "function arguments must match", tc.Function.Arguments, `{"location":"London"}`)
}
