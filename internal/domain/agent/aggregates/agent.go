package aggregates

import (
	"time"

	"github.com/andygeiss/go-agent/internal/domain/agent/entities"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
)

// Agent represents the aggregate root for the agent bounded context.
// It manages the agent's state, tasks, and conversation history.
// The agent follows the observe → decide → act → update loop.
//
//nolint:govet // fieldalignment: struct fields ordered for readability over memory layout
type Agent struct {
	Messages         []entities.Message `json:"messages"`
	Tasks            []entities.Task    `json:"tasks"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	ID               immutable.AgentID  `json:"id"`
	SystemPrompt     string             `json:"system_prompt"`
	CurrentIteration int                `json:"current_iteration"`
	MaxIterations    int                `json:"max_iterations"`
}

// NewAgent creates a new agent with the given ID and system prompt.
func NewAgent(id immutable.AgentID, systemPrompt string) Agent {
	now := time.Now()
	return Agent{
		ID:            id,
		MaxIterations: 10,
		Messages:      make([]entities.Message, 0),
		SystemPrompt:  systemPrompt,
		Tasks:         make([]entities.Task, 0),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// AddMessage adds a message to the conversation history.
func (a *Agent) AddMessage(message entities.Message) {
	a.Messages = append(a.Messages, message)
	a.UpdatedAt = time.Now()
}

// AddTask adds a new task to the agent.
func (a *Agent) AddTask(task entities.Task) {
	a.Tasks = append(a.Tasks, task)
	a.UpdatedAt = time.Now()
}

// CanContinue returns true if the agent can continue the loop.
func (a *Agent) CanContinue() bool {
	return a.CurrentIteration < a.MaxIterations
}

// ClearMessages clears the conversation history.
func (a *Agent) ClearMessages() {
	a.Messages = make([]entities.Message, 0)
	a.UpdatedAt = time.Now()
}

// GetCurrentTask returns the current active task (first non-terminal task).
func (a *Agent) GetCurrentTask() *entities.Task {
	for i := range a.Tasks {
		if !a.Tasks[i].IsTerminal() {
			return &a.Tasks[i]
		}
	}
	return nil
}

// GetMessages returns the full conversation history including the system prompt.
func (a *Agent) GetMessages() []entities.Message {
	messages := make([]entities.Message, 0, len(a.Messages)+1)
	if a.SystemPrompt != "" {
		messages = append(messages, entities.NewMessage(immutable.RoleSystem, a.SystemPrompt))
	}
	messages = append(messages, a.Messages...)
	return messages
}

// HasPendingTasks returns true if there are tasks that are not yet completed or failed.
func (a *Agent) HasPendingTasks() bool {
	for _, task := range a.Tasks {
		if !task.IsTerminal() {
			return true
		}
	}
	return false
}

// IncrementIteration increments the current iteration counter.
func (a *Agent) IncrementIteration() {
	a.CurrentIteration++
	a.UpdatedAt = time.Now()
}

// ResetIteration resets the iteration counter to zero.
func (a *Agent) ResetIteration() {
	a.CurrentIteration = 0
	a.UpdatedAt = time.Now()
}

// WithMaxIterations sets the maximum number of iterations for the agent loop.
func (a *Agent) WithMaxIterations(maxIter int) *Agent {
	a.MaxIterations = maxIter
	a.UpdatedAt = time.Now()
	return a
}
