package agent

// LLMResponse represents the response from an LLM.
// It contains the assistant message and any tool calls requested.
type LLMResponse struct {
	FinishReason string     // Why the LLM stopped (e.g., "stop", "tool_calls")
	Message      Message    // The response message from the LLM
	ToolCalls    []ToolCall // Tool calls requested by the LLM
}

// NewLLMResponse creates a new LLMResponse with the given message and finish reason.
func NewLLMResponse(message Message, finishReason string) LLMResponse {
	return LLMResponse{
		FinishReason: finishReason,
		Message:      message,
		ToolCalls:    make([]ToolCall, 0),
	}
}

// HasToolCalls returns true if the response contains tool calls.
func (r LLMResponse) HasToolCalls() bool {
	return len(r.ToolCalls) > 0
}

// WithToolCalls sets the tool calls on the response.
func (r LLMResponse) WithToolCalls(toolCalls []ToolCall) LLMResponse {
	r.ToolCalls = toolCalls
	return r
}
