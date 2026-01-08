package agent

import "time"

// Task represents a unit of work for the agent to execute.
// It has a defined lifecycle: Pending → Running → Completed/Failed.
type Task struct {
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	ID          TaskID
	Name        string
	Input       string
	Output      string
	Error       string
	Status      TaskStatus
	Iterations  int
}

// NewTask creates a new Task with the given ID, name, and input.
func NewTask(id TaskID, name string, input string) *Task {
	return &Task{
		ID:        id,
		Name:      name,
		Input:     input,
		Status:    TaskStatusPending,
		CreatedAt: time.Now(),
	}
}

// Complete marks the task as successfully completed with the given output.
func (t *Task) Complete(output string) {
	t.Output = output
	t.Status = TaskStatusCompleted
	t.CompletedAt = time.Now()
}

// Fail marks the task as failed with the given error message.
func (t *Task) Fail(errMsg string) {
	t.Error = errMsg
	t.Status = TaskStatusFailed
	t.CompletedAt = time.Now()
}

// IsTerminal returns true if the task is in a terminal state (completed or failed).
func (t *Task) IsTerminal() bool {
	return t.Status == TaskStatusCompleted || t.Status == TaskStatusFailed
}

// Start marks the task as running.
func (t *Task) Start() {
	t.Status = TaskStatusRunning
	t.StartedAt = time.Now()
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

// WaitTime returns how long the task waited before starting.
// Returns 0 if the task hasn't started.
func (t *Task) WaitTime() time.Duration {
	if t.StartedAt.IsZero() {
		return time.Since(t.CreatedAt)
	}
	return t.StartedAt.Sub(t.CreatedAt)
}

// IncrementIterations increments the iteration counter.
func (t *Task) IncrementIterations() {
	t.Iterations++
}
