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

// ToolCall represents a tool invocation requested by the LLM.
// It tracks the tool name, arguments, and execution result.
type ToolCall struct {
	Arguments string         `json:"arguments"`        // JSON-encoded arguments
	Error     string         `json:"error,omitempty"`  // Error message if failed
	ID        ToolCallID     `json:"id"`               // Unique identifier for this call
	Name      string         `json:"name"`             // Name of the tool to execute
	Result    string         `json:"result,omitempty"` // Execution result
	Status    ToolCallStatus `json:"status,omitempty"` // Current execution state
}

// NewToolCall creates a new ToolCall with the given ID, name, and arguments.
func NewToolCall(id ToolCallID, name string, arguments string) ToolCall {
	return ToolCall{
		Arguments: arguments,
		ID:        id,
		Name:      name,
		Status:    ToolCallStatusPending,
	}
}

// Complete marks the tool call as successfully completed with the given result.
func (tc *ToolCall) Complete(result string) {
	tc.Result = result
	tc.Status = ToolCallStatusCompleted
}

// Execute marks the tool call as currently executing.
func (tc *ToolCall) Execute() {
	tc.Status = ToolCallStatusExecuting
}

// Fail marks the tool call as failed with the given error message.
func (tc *ToolCall) Fail(errMsg string) {
	tc.Error = errMsg
	tc.Status = ToolCallStatusFailed
}

// ToMessage converts the tool call result to a tool response message.
func (tc *ToolCall) ToMessage() Message {
	content := tc.Result
	if tc.Status == ToolCallStatusFailed {
		content = "Error: " + tc.Error
	}
	return NewMessage(RoleTool, content).WithToolCallID(tc.ID)
}
