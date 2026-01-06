package entities

import "github.com/andygeiss/go-agent/internal/domain/agent/immutable"

// Result represents the final result of a task execution.
type Result struct {
	Output  string           `json:"output"`
	Error   string           `json:"error,omitempty"`
	TaskID  immutable.TaskID `json:"task_id"`
	Success bool             `json:"success"`
}

// NewResult creates a new result for the given task.
func NewResult(taskID immutable.TaskID, success bool, output string) Result {
	return Result{
		TaskID:  taskID,
		Success: success,
		Output:  output,
	}
}

// WithError sets the error message for the result.
func (r Result) WithError(err string) Result {
	r.Error = err
	return r
}
