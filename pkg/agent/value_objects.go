// Package agent provides a reusable AI agent framework implementing the
// observe → decide → act → update loop pattern for LLM-based task execution.
package agent

import "time"

// ID Types (alphabetically sorted)

// AgentID is the unique identifier for an agent instance.
// It is a value object that ensures type safety for agent references.
type AgentID string

// TaskID is the unique identifier for a task within an agent.
// It is a value object that ensures type safety for task references.
type TaskID string

// ToolCallID is the unique identifier for a tool call.
// It correlates tool requests from the LLM with their execution results.
type ToolCallID string

// ToolID uniquely identifies a tool.
type ToolID string

// Result represents the outcome of a task execution.
// It indicates success/failure and contains the output or error.
type Result struct {
	Error          string        // Error message if failed
	Output         string        // The output if successful
	TaskID         TaskID        // ID of the task that produced this result
	Tokens         TokenUsage    // Token usage statistics
	Duration       time.Duration // How long the task took to execute
	IterationCount int           // Number of agent loop iterations
	ToolCallCount  int           // Number of tool calls made
	Success        bool          // Whether the task completed successfully
}

// TokenUsage tracks the number of tokens used in an LLM interaction.
type TokenUsage struct {
	CompletionTokens int // Tokens in the output/completion
	PromptTokens     int // Tokens in the input/prompt
	TotalTokens      int // Total tokens used
}

// NewResult creates a new Result for the given task.
func NewResult(taskID TaskID, success bool, output string) Result {
	return Result{
		Output:  output,
		Success: success,
		TaskID:  taskID,
	}
}

// WithDuration sets the execution duration on the result.
func (r Result) WithDuration(d time.Duration) Result {
	r.Duration = d
	return r
}

// WithError sets an error message on the result.
func (r Result) WithError(errMsg string) Result {
	r.Error = errMsg
	return r
}

// WithIterationCount sets the iteration count on the result.
func (r Result) WithIterationCount(count int) Result {
	r.IterationCount = count
	return r
}

// WithTokens sets the token usage on the result.
func (r Result) WithTokens(tokens TokenUsage) Result {
	r.Tokens = tokens
	return r
}

// WithToolCallCount sets the tool call count on the result.
func (r Result) WithToolCallCount(count int) Result {
	r.ToolCallCount = count
	return r
}

// Role represents the role of a message in a conversation.
// It follows the OpenAI chat completion API role convention.
type Role string

// Standard conversation roles for LLM chat completions (alphabetically sorted).
const (
	RoleAssistant Role = "assistant" // LLM response
	RoleSystem    Role = "system"    // System instructions/prompt
	RoleTool      Role = "tool"      // Tool execution result
	RoleUser      Role = "user"      // Human input
)

// TaskStatus represents the lifecycle state of a task.
// Tasks transition: Pending → InProgress → Completed/Failed.
type TaskStatus string

// Task lifecycle states (alphabetically sorted).
const (
	TaskStatusCompleted TaskStatus = "completed" // Finished successfully
	TaskStatusFailed    TaskStatus = "failed"    // Terminated with error
	TaskStatusPending   TaskStatus = "pending"   // Awaiting execution
	TaskStatusRunning   TaskStatus = "running"   // Currently running
)

// ToolCallStatus represents the execution state of a tool call.
// Tool calls transition: Pending → Executing → Completed/Failed.
type ToolCallStatus string

// Tool call execution states (alphabetically sorted).
const (
	ToolCallStatusCompleted ToolCallStatus = "completed" // Finished successfully
	ToolCallStatusExecuting ToolCallStatus = "executing" // Currently running
	ToolCallStatusFailed    ToolCallStatus = "failed"    // Terminated with error
	ToolCallStatusPending   ToolCallStatus = "pending"   // Queued for execution
)

// Tool represents a complete tool aggregate with its function and definition.
// Use this to bundle a tool's implementation with its LLM-facing definition.
type Tool struct {
	ID         ToolID
	Func       ToolFunc
	Definition ToolDefinition
}
