package entities

import (
	"time"

	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
)

// Task represents a task to be executed by the agent.
// It is an entity within the Agent aggregate.
type Task struct {
	ID          immutable.TaskID     `json:"id"`
	Status      immutable.TaskStatus `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Description string               `json:"description"`
	Error       string               `json:"error,omitempty"`
	Input       string               `json:"input"`
	Name        string               `json:"name"`
	Output      string               `json:"output"`
}

// NewTask creates a new task with the given ID, name, and input.
func NewTask(id immutable.TaskID, name string, input string) Task {
	now := time.Now()
	return Task{
		ID:        id,
		Input:     input,
		Name:      name,
		Status:    immutable.TaskStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Complete marks the task as completed with the given output.
func (t *Task) Complete(output string) {
	t.Status = immutable.TaskStatusCompleted
	t.Output = output
	t.UpdatedAt = time.Now()
}

// Fail marks the task as failed with the given error.
func (t *Task) Fail(err string) {
	t.Status = immutable.TaskStatusFailed
	t.Error = err
	t.UpdatedAt = time.Now()
}

// IsCompleted returns true if the task is completed.
func (t *Task) IsCompleted() bool {
	return t.Status == immutable.TaskStatusCompleted
}

// IsFailed returns true if the task has failed.
func (t *Task) IsFailed() bool {
	return t.Status == immutable.TaskStatusFailed
}

// IsTerminal returns true if the task is in a terminal state (completed or failed).
func (t *Task) IsTerminal() bool {
	return t.IsCompleted() || t.IsFailed()
}

// Start marks the task as in progress.
func (t *Task) Start() {
	t.Status = immutable.TaskStatusInProgress
	t.UpdatedAt = time.Now()
}

// WithDescription sets the description for the task.
func (t *Task) WithDescription(description string) *Task {
	t.Description = description
	t.UpdatedAt = time.Now()
	return t
}
