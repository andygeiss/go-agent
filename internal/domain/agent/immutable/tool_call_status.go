package immutable

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
