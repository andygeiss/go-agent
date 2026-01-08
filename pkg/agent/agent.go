package agent

import "github.com/andygeiss/cloud-native-utils/slices"

// Agent is the aggregate root that coordinates task execution.
// It maintains conversation state and manages the agent loop lifecycle.
type Agent struct {
	Metadata      Metadata
	ID            AgentID
	SystemPrompt  string
	Messages      []Message
	Tasks         []*Task
	Iteration     int
	MaxIterations int
	MaxMessages   int
}

// Metadata holds arbitrary key-value pairs for agent context.
type Metadata map[string]string

// Option is a functional option for configuring an Agent.
type Option func(*Agent)

// NewAgent creates a new Agent with the given ID and system prompt.
// Default max iterations is set to 10.
func NewAgent(id AgentID, systemPrompt string, opts ...Option) Agent {
	ag := Agent{
		ID:            id,
		SystemPrompt:  systemPrompt,
		Messages:      make([]Message, 0),
		Tasks:         make([]*Task, 0),
		Iteration:     0,
		MaxIterations: 10,
		MaxMessages:   0,
		Metadata:      make(Metadata),
	}
	for _, opt := range opts {
		opt(&ag)
	}
	return ag
}

// WithMaxIterations returns an Option that sets the maximum iterations per task.
func WithMaxIterations(maxIter int) Option {
	return func(a *Agent) {
		a.MaxIterations = maxIter
	}
}

// WithMaxMessages returns an Option that sets the maximum messages to retain.
// When exceeded, older messages are trimmed (keeping system prompt context).
func WithMaxMessages(maxMsg int) Option {
	return func(a *Agent) {
		a.MaxMessages = maxMsg
	}
}

// WithMetadata returns an Option that sets initial metadata.
func WithMetadata(meta Metadata) Option {
	return func(a *Agent) {
		a.Metadata = meta
	}
}

// AddMessage appends a message to the conversation history.
// If MaxMessages is set and exceeded, older messages are trimmed.
func (a *Agent) AddMessage(msg Message) {
	a.Messages = append(a.Messages, msg)
	a.trimMessagesIfNeeded()
}

// AddTask adds a task to the queue.
func (a *Agent) AddTask(task *Task) {
	a.Tasks = append(a.Tasks, task)
}

// CanContinue returns true if the agent has not exceeded max iterations.
func (a *Agent) CanContinue() bool {
	return a.Iteration < a.MaxIterations
}

// ClearMessages removes all messages from the conversation history.
func (a *Agent) ClearMessages() {
	a.Messages = make([]Message, 0)
}

// GetCurrentTask returns the first non-terminal task, or nil if none exist.
func (a *Agent) GetCurrentTask() *Task {
	for _, task := range a.Tasks {
		if !task.IsTerminal() {
			return task
		}
	}
	return nil
}

// GetMessages returns a copy of the conversation history.
func (a *Agent) GetMessages() []Message {
	return a.Messages
}

// GetMetadata returns the value for a metadata key, or empty string if not found.
func (a *Agent) GetMetadata(key string) string {
	return a.Metadata[key]
}

// SetMetadata sets a metadata key-value pair.
func (a *Agent) SetMetadata(key, value string) {
	a.Metadata[key] = value
}

// HasPendingTasks returns true if there are non-terminal tasks in the queue.
func (a *Agent) HasPendingTasks() bool {
	return a.GetCurrentTask() != nil
}

// IncrementIteration increases the iteration counter by one.
func (a *Agent) IncrementIteration() {
	a.Iteration++
}

// ResetIteration sets the iteration counter back to zero.
func (a *Agent) ResetIteration() {
	a.Iteration = 0
}

// SetMaxIterations sets the maximum number of iterations per task.
//
// Deprecated: Use WithMaxIterations option in NewAgent instead.
func (a *Agent) SetMaxIterations(maxIter int) {
	a.MaxIterations = maxIter
}

// TaskCount returns the number of tasks in the queue.
func (a *Agent) TaskCount() int {
	return len(a.Tasks)
}

// MessageCount returns the number of messages in the conversation history.
func (a *Agent) MessageCount() int {
	return len(a.Messages)
}

// CompletedTaskCount returns the number of completed tasks.
func (a *Agent) CompletedTaskCount() int {
	return len(slices.Filter(a.Tasks, func(t *Task) bool {
		return t.Status == TaskStatusCompleted
	}))
}

// FailedTaskCount returns the number of failed tasks.
func (a *Agent) FailedTaskCount() int {
	return len(slices.Filter(a.Tasks, func(t *Task) bool {
		return t.Status == TaskStatusFailed
	}))
}

// trimMessagesIfNeeded removes oldest messages if MaxMessages limit is exceeded.
// It preserves the most recent messages to maintain conversation context.
func (a *Agent) trimMessagesIfNeeded() {
	if a.MaxMessages <= 0 || len(a.Messages) <= a.MaxMessages {
		return
	}
	// Keep only the most recent MaxMessages
	excess := len(a.Messages) - a.MaxMessages
	a.Messages = a.Messages[excess:]
}
