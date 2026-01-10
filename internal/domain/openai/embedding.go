package openai

// EmbeddingRequest represents a request to the embeddings endpoint.
type EmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// NewEmbeddingRequest creates a new embedding request.
func NewEmbeddingRequest(model, input string) EmbeddingRequest {
	return EmbeddingRequest{
		Input: input,
		Model: model,
	}
}

// EmbeddingResponse represents a response from the embeddings endpoint.
type EmbeddingResponse struct {
	Model  string          `json:"model"`
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Usage  EmbeddingUsage  `json:"usage"`
}

// EmbeddingData contains the embedding vector.
type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingUsage contains token usage information.
type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// GetFirstEmbedding returns the first embedding from the response, or nil if empty.
func (r EmbeddingResponse) GetFirstEmbedding() []float32 {
	if len(r.Data) == 0 {
		return nil
	}
	return r.Data[0].Embedding
}
