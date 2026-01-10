//go:build integration

package outbound_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
)

// -----------------------------------------------------------------------------
// Integration tests for OpenAI-compatible embedding API
// -----------------------------------------------------------------------------
//
// These tests require a running LM Studio (or compatible) server with an embedding model.
// Run with: go test -tags=integration ./...
//
// Environment variables:
//   - OPENAI_EMBED_URL: Base URL for the embedding API (default: http://localhost:1234)
//   - OPENAI_EMBED_MODEL: Embedding model name (required)

func getEmbeddingURL() string {
	if url := os.Getenv("OPENAI_EMBED_URL"); url != "" {
		return url
	}
	return "http://localhost:1234"
}

func getEmbeddingModel() string {
	return os.Getenv("OPENAI_EMBED_MODEL")
}

func skipIfNoEmbeddingModel(t *testing.T) {
	if getEmbeddingModel() == "" {
		t.Skip("OPENAI_EMBED_MODEL not set, skipping integration test")
	}
}

// -----------------------------------------------------------------------------
// Embedding integration tests
// -----------------------------------------------------------------------------

func Test_Integration_OpenAIEmbeddingClient_Embed_With_SimpleText_Should_ReturnEmbedding(t *testing.T) {
	// Skip if no model configured
	skipIfNoEmbeddingModel(t)

	// Arrange
	client := outbound.NewOpenAIEmbeddingClient(getEmbeddingURL()).
		WithModel(getEmbeddingModel()).
		WithTimeout(30 * time.Second)

	// Act
	result, err := client.Embed(context.Background(), "Hello, world!")

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "embedding must not be empty", len(result) > 0, true)
}

func Test_Integration_OpenAIEmbeddingClient_Embed_With_LongText_Should_ReturnEmbedding(t *testing.T) {
	// Skip if no model configured
	skipIfNoEmbeddingModel(t)

	// Arrange
	client := outbound.NewOpenAIEmbeddingClient(getEmbeddingURL()).
		WithModel(getEmbeddingModel()).
		WithTimeout(30 * time.Second)

	longText := `This is a longer piece of text that should still be embedded correctly.
	It contains multiple sentences and spans several lines.
	The embedding model should be able to process this without any issues.
	We're testing that the client handles larger inputs properly.`

	// Act
	result, err := client.Embed(context.Background(), longText)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "embedding must not be empty", len(result) > 0, true)
}

func Test_Integration_OpenAIEmbeddingClient_Embed_Should_ReturnConsistentDimensions(t *testing.T) {
	// Skip if no model configured
	skipIfNoEmbeddingModel(t)

	// Arrange
	client := outbound.NewOpenAIEmbeddingClient(getEmbeddingURL()).
		WithModel(getEmbeddingModel()).
		WithTimeout(30 * time.Second)

	texts := []string{
		"Short text",
		"A medium length piece of text for embedding",
		"This is a much longer piece of text that contains significantly more words and should test that the embedding dimensions remain consistent regardless of input length",
	}

	// Act
	var dimensions []int
	for _, text := range texts {
		result, err := client.Embed(context.Background(), text)
		assert.That(t, "must not return error", err, nil)
		dimensions = append(dimensions, len(result))
	}

	// Assert
	// All embeddings should have the same dimensions
	for i := 1; i < len(dimensions); i++ {
		assert.That(t, "dimensions must be consistent", dimensions[i], dimensions[0])
	}
}

func Test_Integration_OpenAIEmbeddingClient_Embed_SimilarTexts_Should_HaveSimilarEmbeddings(t *testing.T) {
	// Skip if no model configured
	skipIfNoEmbeddingModel(t)

	// Arrange
	client := outbound.NewOpenAIEmbeddingClient(getEmbeddingURL()).
		WithModel(getEmbeddingModel()).
		WithTimeout(30 * time.Second)

	// Similar texts should have similar embeddings
	text1 := "The quick brown fox jumps over the lazy dog"
	text2 := "A fast brown fox leaps over a sleepy dog"

	// A completely different text
	text3 := "Machine learning is transforming the software industry"

	// Act
	emb1, err1 := client.Embed(context.Background(), text1)
	emb2, err2 := client.Embed(context.Background(), text2)
	emb3, err3 := client.Embed(context.Background(), text3)

	// Assert
	assert.That(t, "must not return error for text1", err1, nil)
	assert.That(t, "must not return error for text2", err2, nil)
	assert.That(t, "must not return error for text3", err3, nil)

	// Calculate cosine similarities
	sim12 := cosineSimilarity(emb1, emb2)
	sim13 := cosineSimilarity(emb1, emb3)

	// Similar texts should have higher similarity than dissimilar texts
	assert.That(t, "similar texts should have higher similarity", sim12 > sim13, true)
}

func Test_Integration_OpenAIEmbeddingClient_Embed_With_CanceledContext_Should_ReturnError(t *testing.T) {
	// Skip if no model configured
	skipIfNoEmbeddingModel(t)

	// Arrange
	client := outbound.NewOpenAIEmbeddingClient(getEmbeddingURL()).
		WithModel(getEmbeddingModel())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	_, err := client.Embed(ctx, "Hello, world!")

	// Assert
	assert.That(t, "must return error", err != nil, true)
}

func Test_Integration_OpenAIEmbeddingClient_Embed_With_EmptyText_Should_ReturnEmbedding(t *testing.T) {
	// Skip if no model configured
	skipIfNoEmbeddingModel(t)

	// Arrange
	client := outbound.NewOpenAIEmbeddingClient(getEmbeddingURL()).
		WithModel(getEmbeddingModel()).
		WithTimeout(30 * time.Second)

	// Act
	result, err := client.Embed(context.Background(), "")

	// Assert
	// Some models may return an embedding for empty text, others may error
	// We just verify the call completes (either way is valid behavior)
	if err == nil {
		assert.That(t, "if no error, embedding should exist", len(result) > 0, true)
	}
	// If err != nil, that's also acceptable behavior for empty input
}

// cosineSimilarity calculates the cosine similarity between two embedding vectors.
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

// sqrt is a simple square root implementation to avoid importing math.
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 100; i++ {
		z = (z + x/z) / 2
	}
	return z
}
