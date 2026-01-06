package entities

import (
	"time"

	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
)

// ToolCall represents a tool call requested by the LLM.
// It is an entity within the Agent aggregate.
type ToolCall struct {
	ID        immutable.ToolCallID     `json:"id"`
	Status    immutable.ToolCallStatus `json:"status"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
	Arguments string                   `json:"arguments"`
	Error     string                   `json:"error,omitempty"`
	Name      string                   `json:"name"`
	Result    string                   `json:"result"`
}

// NewToolCall creates a new tool call with the given ID, name, and arguments.
func NewToolCall(id immutable.ToolCallID, name string, arguments string) ToolCall {
	now := time.Now()
	return ToolCall{
		ID:        id,
		Arguments: arguments,
		Name:      name,
		Status:    immutable.ToolCallStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Complete marks the tool call as completed with the given result.
func (tc *ToolCall) Complete(result string) {
	tc.Status = immutable.ToolCallStatusCompleted
	tc.Result = result
	tc.UpdatedAt = time.Now()
}

// Execute marks the tool call as executing.
func (tc *ToolCall) Execute() {
	tc.Status = immutable.ToolCallStatusExecuting
	tc.UpdatedAt = time.Now()
}

// Fail marks the tool call as failed with the given error.
func (tc *ToolCall) Fail(err string) {
	tc.Status = immutable.ToolCallStatusFailed
	tc.Error = err
	tc.UpdatedAt = time.Now()
}

// IsCompleted returns true if the tool call is completed.
func (tc *ToolCall) IsCompleted() bool {
	return tc.Status == immutable.ToolCallStatusCompleted
}

// IsFailed returns true if the tool call has failed.
func (tc *ToolCall) IsFailed() bool {
	return tc.Status == immutable.ToolCallStatusFailed
}

// ToMessage converts the tool call result to a message for the LLM.
func (tc *ToolCall) ToMessage() Message {
	content := tc.Result
	if tc.IsFailed() {
		content = "Error: " + tc.Error
	}
	return NewMessage(immutable.RoleTool, content).WithToolCallID(string(tc.ID))
}
