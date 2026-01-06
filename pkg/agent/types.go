// Package agent provides a reusable AI agent framework implementing the
// observe → decide → act → update loop pattern for LLM-based task execution.
package agent

// AgentID is the unique identifier for an agent instance.
// It is a value object that ensures type safety for agent references.
type AgentID string

// TaskID is the unique identifier for a task within an agent.
// It is a value object that ensures type safety for task references.
type TaskID string

// ToolCallID is the unique identifier for a tool call.
// It correlates tool requests from the LLM with their execution results.
type ToolCallID string

// Role represents the role of a message in a conversation.
// It follows the OpenAI chat completion API role convention.
type Role string

// Standard conversation roles for LLM chat completions.
const (
	RoleSystem    Role = "system"    // System instructions/prompt
	RoleUser      Role = "user"      // Human input
	RoleAssistant Role = "assistant" // LLM response
	RoleTool      Role = "tool"      // Tool execution result
)

// TaskStatus represents the lifecycle state of a task.
// Tasks transition: Pending → InProgress → Completed/Failed.
type TaskStatus string

// Task lifecycle states.
const (
	TaskStatusPending   TaskStatus = "pending"   // Awaiting execution
	TaskStatusRunning   TaskStatus = "running"   // Currently running
	TaskStatusCompleted TaskStatus = "completed" // Finished successfully
	TaskStatusFailed    TaskStatus = "failed"    // Terminated with error
)

// ToolCallStatus represents the execution state of a tool call.
// Tool calls transition: Pending → Executing → Completed/Failed.
type ToolCallStatus string

// Tool call execution states.
const (
	ToolCallStatusPending   ToolCallStatus = "pending"   // Queued for execution
	ToolCallStatusExecuting ToolCallStatus = "executing" // Currently running
	ToolCallStatusCompleted ToolCallStatus = "completed" // Finished successfully
	ToolCallStatusFailed    ToolCallStatus = "failed"    // Terminated with error
)
