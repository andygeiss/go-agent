package outbound_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/pkg/agent"
	"github.com/andygeiss/go-agent/pkg/openai"
)

// -----------------------------------------------------------------------------
// NewOpenAIClient tests
// -----------------------------------------------------------------------------

func Test_NewOpenAIClient_Should_CreateClient(t *testing.T) {
	// Arrange & Act
	client := outbound.NewOpenAIClient("http://localhost:1234", "test-model")

	// Assert
	assert.That(t, "client must not be nil", client != nil, true)
}

func Test_OpenAIClient_WithHTTPClient_Should_SetCustomClient(t *testing.T) {
	// Arrange
	client := outbound.NewOpenAIClient("http://localhost:1234", "test-model")
	customHTTPClient := &http.Client{}

	// Act
	result := client.WithHTTPClient(customHTTPClient)

	// Assert
	assert.That(t, "must return client for chaining", result != nil, true)
}

// -----------------------------------------------------------------------------
// Run tests with mock HTTP server
// -----------------------------------------------------------------------------

func Test_OpenAIClient_Run_With_SimpleResponse_Should_ReturnMessage(t *testing.T) {
	// Arrange
	response := openai.ChatCompletionResponse{
		ID:    "chatcmpl-123",
		Model: "test-model",
		Choices: []openai.ChatCompletionChoice{
			{
				Index:        0,
				FinishReason: "stop",
				Message: openai.Message{
					Role:    "assistant",
					Content: "Hello! How can I help you?",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := outbound.NewOpenAIClient(server.URL, "test-model")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
	}

	// Act
	result, err := client.Run(context.Background(), messages, nil)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "message content must match", result.Message.Content, "Hello! How can I help you?")
	assert.That(t, "finish reason must be stop", result.FinishReason, "stop")
}

func Test_OpenAIClient_Run_With_ToolCalls_Should_ReturnToolCalls(t *testing.T) {
	// Arrange
	response := openai.ChatCompletionResponse{
		ID:    "chatcmpl-456",
		Model: "test-model",
		Choices: []openai.ChatCompletionChoice{
			{
				Index:        0,
				FinishReason: "tool_calls",
				Message: openai.Message{
					Role:    "assistant",
					Content: "",
					ToolCalls: []openai.ToolCall{
						{
							ID:   "call_123",
							Type: "function",
							Function: openai.FunctionCall{
								Name:      "get_time",
								Arguments: "{}",
							},
						},
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := outbound.NewOpenAIClient(server.URL, "test-model")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "What time is it?"),
	}
	tools := []agent.ToolDefinition{
		agent.NewToolDefinition("get_time", "Get the current time"),
	}

	// Act
	result, err := client.Run(context.Background(), messages, tools)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "must have 1 tool call", len(result.ToolCalls), 1)
	assert.That(t, "tool call name must match", result.ToolCalls[0].Name, "get_time")
	assert.That(t, "finish reason must be tool_calls", result.FinishReason, "tool_calls")
}

func Test_OpenAIClient_Run_With_EmptyChoices_Should_ReturnError(t *testing.T) {
	// Arrange
	response := openai.ChatCompletionResponse{
		ID:      "chatcmpl-789",
		Model:   "test-model",
		Choices: []openai.ChatCompletionChoice{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := outbound.NewOpenAIClient(server.URL, "test-model")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
	}

	// Act
	_, err := client.Run(context.Background(), messages, nil)

	// Assert
	assert.That(t, "must return error", err != nil, true)
}

func Test_OpenAIClient_Run_With_HTTPError_Should_ReturnError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := outbound.NewOpenAIClient(server.URL, "test-model")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
	}

	// Act
	_, err := client.Run(context.Background(), messages, nil)

	// Assert
	assert.That(t, "must return error", err != nil, true)
}

func Test_OpenAIClient_Run_With_InvalidJSON_Should_ReturnError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := outbound.NewOpenAIClient(server.URL, "test-model")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
	}

	// Act
	_, err := client.Run(context.Background(), messages, nil)

	// Assert
	assert.That(t, "must return error", err != nil, true)
}

func Test_OpenAIClient_Run_With_MessageContainingToolCallID_Should_ConvertCorrectly(t *testing.T) {
	// Arrange
	response := openai.ChatCompletionResponse{
		ID:    "chatcmpl-123",
		Model: "test-model",
		Choices: []openai.ChatCompletionChoice{
			{
				Index:        0,
				FinishReason: "stop",
				Message: openai.Message{
					Role:    "assistant",
					Content: "The time is 12:00 PM",
				},
			},
		},
	}

	var receivedRequest openai.ChatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedRequest)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := outbound.NewOpenAIClient(server.URL, "test-model")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "What time is it?"),
		agent.NewMessage(agent.RoleAssistant, "").WithToolCalls([]agent.ToolCall{
			agent.NewToolCall("call_123", "get_time", "{}"),
		}),
		agent.NewMessage(agent.RoleTool, "12:00 PM").WithToolCallID("call_123"),
	}

	// Act
	_, err := client.Run(context.Background(), messages, nil)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "must have 3 messages in request", len(receivedRequest.Messages), 3)
	assert.That(t, "tool message must have tool_call_id", receivedRequest.Messages[2].ToolCallID, "call_123")
}

func Test_OpenAIClient_Run_With_Tools_Should_IncludeToolsInRequest(t *testing.T) {
	// Arrange
	response := openai.ChatCompletionResponse{
		ID:    "chatcmpl-123",
		Model: "test-model",
		Choices: []openai.ChatCompletionChoice{
			{
				Index:        0,
				FinishReason: "stop",
				Message:      openai.Message{Role: "assistant", Content: "OK"},
			},
		},
	}

	var receivedRequest openai.ChatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedRequest)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := outbound.NewOpenAIClient(server.URL, "test-model")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "What time is it?"),
	}
	tools := []agent.ToolDefinition{
		agent.NewToolDefinition("get_time", "Get the current time"),
		agent.NewToolDefinition("calculate", "Perform a calculation").WithParameter("expression", "The math expression"),
	}

	// Act
	_, err := client.Run(context.Background(), messages, tools)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "must have 2 tools in request", len(receivedRequest.Tools), 2)
	assert.That(t, "first tool name must match", receivedRequest.Tools[0].Function.Name, "get_time")
	assert.That(t, "second tool name must match", receivedRequest.Tools[1].Function.Name, "calculate")
}

func Test_OpenAIClient_Run_With_ContextCanceled_Should_ReturnError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Slow response
		select {}
	}))
	defer server.Close()

	client := outbound.NewOpenAIClient(server.URL, "test-model")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	_, err := client.Run(ctx, messages, nil)

	// Assert
	assert.That(t, "must return error", err != nil, true)
}
