package ports

import (
	"context"

	"github.com/andygeiss/go-agent/internal/domain/agent/aggregates"
	"github.com/andygeiss/go-agent/internal/domain/agent/entities"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
)

// LLMClient represents the interface for interacting with a Large Language Model.
// It abstracts the communication with external LLM services (e.g., LM Studio, OpenAI).
// The domain layer uses this interface without knowing the implementation details.
type LLMClient interface {
	// Run sends the conversation messages to the LLM and returns the response.
	// The messages should include the system prompt, user messages, assistant responses,
	// and any tool call results from previous iterations.
	// The tools parameter provides the available tool definitions for the LLM to use.
	Run(ctx context.Context, messages []entities.Message, tools []immutable.ToolDefinition) (aggregates.LLMResponse, error)
}
