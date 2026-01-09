package agent

import "time"

// Task represents a unit of work for the agent to execute.
// It has a defined lifecycle: Pending → Running → Completed/Failed.
type Task struct {
	CompletedAt time.Time
	CreatedAt   time.Time
	StartedAt   time.Time
	Error       string
	Input       string
	Name        string
	Output      string
	ID          TaskID
	Status      TaskStatus
	Iterations  int
}

// NewTask creates a new Task with the given ID, name, and input.
func NewTask(id TaskID, name string, input string) *Task {
	return &Task{
		CreatedAt: time.Now(),
		ID:        id,
		Input:     input,
		Name:      name,
		Status:    TaskStatusPending,
	}
}

// Complete marks the task as successfully completed with the given output.
func (t *Task) Complete(output string) {
	t.CompletedAt = time.Now()
	t.Output = output
	t.Status = TaskStatusCompleted
}

// Duration returns the execution duration of the task.
// Returns 0 if the task hasn't started or hasn't completed.
func (t *Task) Duration() time.Duration {
	if t.StartedAt.IsZero() {
		return 0
	}
	if t.CompletedAt.IsZero() {
		return time.Since(t.StartedAt)
	}
	return t.CompletedAt.Sub(t.StartedAt)
}

// Fail marks the task as failed with the given error message.
func (t *Task) Fail(errMsg string) {
	t.CompletedAt = time.Now()
	t.Error = errMsg
	t.Status = TaskStatusFailed
}

// IncrementIterations increments the iteration counter.
func (t *Task) IncrementIterations() {
	t.Iterations++
}

// IsTerminal returns true if the task is in a terminal state (completed or failed).
func (t *Task) IsTerminal() bool {
	return t.Status == TaskStatusCompleted || t.Status == TaskStatusFailed
}

// Start marks the task as running.
func (t *Task) Start() {
	t.StartedAt = time.Now()
	t.Status = TaskStatusRunning
}

// WaitTime returns how long the task waited before starting.
// Returns 0 if the task hasn't started.
func (t *Task) WaitTime() time.Duration {
	if t.StartedAt.IsZero() {
		return time.Since(t.CreatedAt)
	}
	return t.StartedAt.Sub(t.CreatedAt)
}
