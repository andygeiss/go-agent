package agent

// Task represents a unit of work for the agent to execute.
// It has a defined lifecycle: Pending → Running → Completed/Failed.
type Task struct {
	ID     TaskID     // Unique identifier for the task
	Name   string     // Human-readable task name
	Input  string     // The input/prompt for the task
	Output string     // The result after completion
	Error  string     // Error message if failed
	Status TaskStatus // Current lifecycle state
}

// NewTask creates a new Task with the given ID, name, and input.
func NewTask(id TaskID, name string, input string) *Task {
	return &Task{
		ID:     id,
		Name:   name,
		Input:  input,
		Status: TaskStatusPending,
	}
}

// Complete marks the task as successfully completed with the given output.
func (t *Task) Complete(output string) {
	t.Output = output
	t.Status = TaskStatusCompleted
}

// Fail marks the task as failed with the given error message.
func (t *Task) Fail(errMsg string) {
	t.Error = errMsg
	t.Status = TaskStatusFailed
}

// IsTerminal returns true if the task is in a terminal state (completed or failed).
func (t *Task) IsTerminal() bool {
	return t.Status == TaskStatusCompleted || t.Status == TaskStatusFailed
}

// Start marks the task as running.
func (t *Task) Start() {
	t.Status = TaskStatusRunning
}
