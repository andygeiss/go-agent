package agent

import "context"

// LLMClient is the interface for communicating with a language model.
// Implementations translate between domain types and LLM-specific APIs.
type LLMClient interface {
	// Run sends messages to the LLM and returns its response.
	Run(ctx context.Context, messages []Message, tools []ToolDefinition) (LLMResponse, error)
}
