package immutable

// TaskStatus represents the lifecycle state of a task.
// Tasks transition: Pending → InProgress → Completed/Failed.
type TaskStatus string

// Task lifecycle states.
const (
	TaskStatusPending    TaskStatus = "pending"     // Awaiting execution
	TaskStatusInProgress TaskStatus = "in_progress" // Currently running
	TaskStatusCompleted  TaskStatus = "completed"   // Finished successfully
	TaskStatusFailed     TaskStatus = "failed"      // Terminated with error
)
