package agent

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
