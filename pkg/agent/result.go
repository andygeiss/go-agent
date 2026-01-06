package agent

// Result represents the outcome of a task execution.
// It indicates success/failure and contains the output or error.
type Result struct {
	TaskID  TaskID // ID of the task that produced this result
	Output  string // The output if successful
	Error   string // Error message if failed
	Success bool   // Whether the task completed successfully
}

// NewResult creates a new Result for the given task.
func NewResult(taskID TaskID, success bool, output string) Result {
	return Result{
		TaskID:  taskID,
		Success: success,
		Output:  output,
	}
}

// WithError sets an error message on the result.
func (r Result) WithError(errMsg string) Result {
	r.Error = errMsg
	return r
}
