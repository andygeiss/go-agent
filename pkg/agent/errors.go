package agent

import "errors"

// Sentinel errors for common failure conditions.
var (
	// ErrMaxIterationsReached is returned when the agent exceeds the maximum allowed iterations.
	ErrMaxIterationsReached = errors.New("max iterations reached")

	// ErrContextCanceled is returned when the context is canceled during execution.
	ErrContextCanceled = errors.New("context canceled")

	// ErrNoResponse is returned when the LLM returns an empty response.
	ErrNoResponse = errors.New("no response from LLM")

	// ErrToolNotFound is returned when trying to execute an unknown tool.
	ErrToolNotFound = errors.New("tool not found")

	// ErrInvalidArguments is returned when tool arguments are malformed.
	ErrInvalidArguments = errors.New("invalid tool arguments")
)

// LLMError wraps errors from the LLM client with additional context.
type LLMError struct {
	Cause   error
	Message string
}

// NewLLMError creates a new LLMError with the given message and cause.
func NewLLMError(message string, cause error) *LLMError {
	return &LLMError{
		Message: message,
		Cause:   cause,
	}
}

// Error implements the error interface.
func (e *LLMError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Unwrap returns the underlying cause for errors.Is/As support.
func (e *LLMError) Unwrap() error {
	return e.Cause
}

// ToolError wraps errors from tool execution with additional context.
type ToolError struct {
	Cause    error
	ToolName string
	Message  string
}

// NewToolError creates a new ToolError with the given tool name, message, and cause.
func NewToolError(toolName, message string, cause error) *ToolError {
	return &ToolError{
		ToolName: toolName,
		Message:  message,
		Cause:    cause,
	}
}

// Error implements the error interface.
func (e *ToolError) Error() string {
	base := "tool " + e.ToolName + ": " + e.Message
	if e.Cause != nil {
		return base + ": " + e.Cause.Error()
	}
	return base
}

// Unwrap returns the underlying cause for errors.Is/As support.
func (e *ToolError) Unwrap() error {
	return e.Cause
}

// TaskError wraps errors from task execution with additional context.
type TaskError struct {
	Cause  error
	TaskID TaskID
	Reason string
}

// NewTaskError creates a new TaskError with the given task ID, reason, and cause.
func NewTaskError(taskID TaskID, reason string, cause error) *TaskError {
	return &TaskError{
		TaskID: taskID,
		Reason: reason,
		Cause:  cause,
	}
}

// Error implements the error interface.
func (e *TaskError) Error() string {
	base := "task " + string(e.TaskID) + ": " + e.Reason
	if e.Cause != nil {
		return base + ": " + e.Cause.Error()
	}
	return base
}

// Unwrap returns the underlying cause for errors.Is/As support.
func (e *TaskError) Unwrap() error {
	return e.Cause
}
