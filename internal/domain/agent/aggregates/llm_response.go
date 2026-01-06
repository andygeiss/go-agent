package aggregates

import "github.com/andygeiss/go-agent/internal/domain/agent/entities"

// LLMResponse represents a response from the LLM.
type LLMResponse struct {
	Message      entities.Message    `json:"message"`
	FinishReason string              `json:"finish_reason"`
	ToolCalls    []entities.ToolCall `json:"tool_calls,omitempty"`
}

// NewLLMResponse creates a new LLM response with the given message.
func NewLLMResponse(message entities.Message, finishReason string) LLMResponse {
	return LLMResponse{
		Message:      message,
		FinishReason: finishReason,
	}
}

// HasToolCalls returns true if the response contains tool calls.
func (r LLMResponse) HasToolCalls() bool {
	return len(r.ToolCalls) > 0
}

// WithToolCalls sets the tool calls for the response.
func (r LLMResponse) WithToolCalls(toolCalls []entities.ToolCall) LLMResponse {
	r.ToolCalls = toolCalls
	return r
}
