package outbound_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/internal/domain/openai"
)

// -----------------------------------------------------------------------------
// NewOpenAIEmbeddingClient tests
// -----------------------------------------------------------------------------

func Test_NewOpenAIEmbeddingClient_Should_CreateClient(t *testing.T) {
	// Arrange & Act
	client := outbound.NewOpenAIEmbeddingClient("http://localhost:1234")

	// Assert
	assert.That(t, "client must not be nil", client != nil, true)
}

func Test_OpenAIEmbeddingClient_WithHTTPClient_Should_SetCustomClient(t *testing.T) {
	// Arrange
	client := outbound.NewOpenAIEmbeddingClient("http://localhost:1234")
	customHTTPClient := &http.Client{}

	// Act
	result := client.WithHTTPClient(customHTTPClient)

	// Assert
	assert.That(t, "must return client for chaining", result != nil, true)
}

func Test_OpenAIEmbeddingClient_WithModel_Should_SetModel(t *testing.T) {
	// Arrange
	client := outbound.NewOpenAIEmbeddingClient("http://localhost:1234")

	// Act
	result := client.WithModel("text-embedding-ada-002")

	// Assert
	assert.That(t, "must return client for chaining", result != nil, true)
}

func Test_OpenAIEmbeddingClient_WithTimeout_Should_SetTimeout(t *testing.T) {
	// Arrange
	client := outbound.NewOpenAIEmbeddingClient("http://localhost:1234")

	// Act
	result := client.WithTimeout(10 * time.Second)

	// Assert
	assert.That(t, "must return client for chaining", result != nil, true)
}

// -----------------------------------------------------------------------------
// Embed tests with mock HTTP server
// -----------------------------------------------------------------------------

func Test_OpenAIEmbeddingClient_Embed_With_ValidResponse_Should_ReturnEmbedding(t *testing.T) {
	// Arrange
	expectedEmbedding := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	response := openai.EmbeddingResponse{
		Object: "list",
		Model:  "text-embedding-3-small",
		Data: []openai.EmbeddingData{
			{
				Object:    "embedding",
				Index:     0,
				Embedding: expectedEmbedding,
			},
		},
		Usage: openai.EmbeddingUsage{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := outbound.NewOpenAIEmbeddingClient(server.URL)

	// Act
	result, err := client.Embed(context.Background(), "Hello, world!")

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "must return correct length", len(result), len(expectedEmbedding))
	assert.That(t, "first value must match", result[0], expectedEmbedding[0])
}

func Test_OpenAIEmbeddingClient_Embed_With_EmptyResponse_Should_ReturnError(t *testing.T) {
	// Arrange
	response := openai.EmbeddingResponse{
		Object: "list",
		Model:  "text-embedding-3-small",
		Data:   []openai.EmbeddingData{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := outbound.NewOpenAIEmbeddingClient(server.URL).
		WithRetry(0, 0) // Disable retries for fast test execution

	// Act
	result, err := client.Embed(context.Background(), "Hello, world!")

	// Assert
	assert.That(t, "must return error", err != nil, true)
	assert.That(t, "result must be nil", result == nil, true)
}

func Test_OpenAIEmbeddingClient_Embed_With_HTTPError_Should_ReturnError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := outbound.NewOpenAIEmbeddingClient(server.URL).
		WithRetry(1, 10*time.Millisecond) // Fast retry for tests

	// Act
	result, err := client.Embed(context.Background(), "Hello, world!")

	// Assert
	assert.That(t, "must return error", err != nil, true)
	assert.That(t, "result must be nil", result == nil, true)
}

func Test_OpenAIEmbeddingClient_Embed_With_InvalidJSON_Should_ReturnError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	client := outbound.NewOpenAIEmbeddingClient(server.URL).
		WithRetry(1, 10*time.Millisecond)

	// Act
	result, err := client.Embed(context.Background(), "Hello, world!")

	// Assert
	assert.That(t, "must return error", err != nil, true)
	assert.That(t, "result must be nil", result == nil, true)
}

func Test_OpenAIEmbeddingClient_Embed_With_ContextCanceled_Should_ReturnError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond) // Simulate slow response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := outbound.NewOpenAIEmbeddingClient(server.URL).
		WithTimeout(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	result, err := client.Embed(ctx, "Hello, world!")

	// Assert
	assert.That(t, "must return error", err != nil, true)
	assert.That(t, "result must be nil", result == nil, true)
}

func Test_OpenAIEmbeddingClient_Embed_Should_SendCorrectRequest(t *testing.T) {
	// Arrange
	var receivedRequest openai.EmbeddingRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedRequest)

		response := openai.EmbeddingResponse{
			Data: []openai.EmbeddingData{
				{Embedding: []float32{0.1}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := outbound.NewOpenAIEmbeddingClient(server.URL).
		WithModel("custom-model")

	// Act
	_, _ = client.Embed(context.Background(), "test input")

	// Assert
	assert.That(t, "input must match", receivedRequest.Input, "test input")
	assert.That(t, "model must match", receivedRequest.Model, "custom-model")
}
