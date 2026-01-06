package agent

// Agent is the aggregate root that coordinates task execution.
// It maintains conversation state and manages the agent loop lifecycle.
type Agent struct {
	ID            AgentID   // Unique identifier for this agent instance
	SystemPrompt  string    // Instructions that define agent behavior
	Messages      []Message // Conversation history
	Tasks         []*Task   // Queue of tasks to execute
	Iteration     int       // Current iteration count within a task
	MaxIterations int       // Safety limit for iterations per task
}

// NewAgent creates a new Agent with the given ID and system prompt.
// Default max iterations is set to 10.
func NewAgent(id AgentID, systemPrompt string) Agent {
	return Agent{
		ID:            id,
		SystemPrompt:  systemPrompt,
		Messages:      make([]Message, 0),
		Tasks:         make([]*Task, 0),
		Iteration:     0,
		MaxIterations: 10,
	}
}

// AddMessage appends a message to the conversation history.
func (a *Agent) AddMessage(msg Message) {
	a.Messages = append(a.Messages, msg)
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

// WithMaxIterations sets the maximum number of iterations per task.
func (a *Agent) WithMaxIterations(maxIter int) {
	a.MaxIterations = maxIter
}
