package openai_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/openai"
)

// -----------------------------------------------------------------------------
// EmbeddingRequest tests
// -----------------------------------------------------------------------------

func Test_NewEmbeddingRequest_Should_SetModelAndInput(t *testing.T) {
	// Arrange & Act
	request := openai.NewEmbeddingRequest("text-embedding-3-small", "Hello, world!")

	// Assert
	assert.That(t, "model must be set", request.Model, "text-embedding-3-small")
	assert.That(t, "input must be set", request.Input, "Hello, world!")
}

// -----------------------------------------------------------------------------
// EmbeddingResponse tests
// -----------------------------------------------------------------------------

func Test_EmbeddingResponse_GetFirstEmbedding_With_Data_Should_ReturnFirst(t *testing.T) {
	// Arrange
	response := openai.EmbeddingResponse{
		Data: []openai.EmbeddingData{
			{Embedding: []float32{0.1, 0.2, 0.3}},
			{Embedding: []float32{0.4, 0.5, 0.6}},
		},
	}

	// Act
	result := response.GetFirstEmbedding()

	// Assert
	assert.That(t, "result must not be nil", result != nil, true)
	assert.That(t, "first value must match", result[0], float32(0.1))
	assert.That(t, "length must match", len(result), 3)
}

func Test_EmbeddingResponse_GetFirstEmbedding_With_EmptyData_Should_ReturnNil(t *testing.T) {
	// Arrange
	response := openai.EmbeddingResponse{
		Data: []openai.EmbeddingData{},
	}

	// Act
	result := response.GetFirstEmbedding()

	// Assert
	assert.That(t, "result must be nil", result == nil, true)
}
