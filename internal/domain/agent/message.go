package agent

// Message represents a single message in a conversation.
// It follows the OpenAI chat completion message format.
type Message struct {
	Content    string     `json:"content"`
	Role       Role       `json:"role"`
	ToolCallID ToolCallID `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// NewMessage creates a new Message with the given role and content.
func NewMessage(role Role, content string) Message {
	return Message{
		Content:   content,
		Role:      role,
		ToolCalls: make([]ToolCall, 0),
	}
}

// WithToolCallID sets the tool call ID for tool response messages.
func (m Message) WithToolCallID(id ToolCallID) Message {
	m.ToolCallID = id
	return m
}

// WithToolCalls attaches tool calls to the message.
func (m Message) WithToolCalls(toolCalls []ToolCall) Message {
	m.ToolCalls = toolCalls
	return m
}

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
