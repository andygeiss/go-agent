package events

import "github.com/andygeiss/go-agent/internal/domain/agent/immutable"

// EventTaskFailed represents a task failed event.
type EventTaskFailed struct {
	AgentID    immutable.AgentID `json:"agent_id"`
	Error      string            `json:"error"`
	Name       string            `json:"name"`
	TaskID     immutable.TaskID  `json:"task_id"`
	Iterations int               `json:"iterations"`
}

// NewEventTaskFailed creates a new EventTaskFailed instance.
func NewEventTaskFailed() *EventTaskFailed {
	return &EventTaskFailed{}
}

// Topic returns the topic for the event.
func (e *EventTaskFailed) Topic() string {
	return EventTopicTaskFailed
}

// WithAgentID sets the AgentID field.
func (e *EventTaskFailed) WithAgentID(id immutable.AgentID) *EventTaskFailed {
	e.AgentID = id
	return e
}

// WithError sets the Error field.
func (e *EventTaskFailed) WithError(err string) *EventTaskFailed {
	e.Error = err
	return e
}

// WithIterations sets the Iterations field.
func (e *EventTaskFailed) WithIterations(iterations int) *EventTaskFailed {
	e.Iterations = iterations
	return e
}

// WithName sets the Name field.
func (e *EventTaskFailed) WithName(name string) *EventTaskFailed {
	e.Name = name
	return e
}

// WithTaskID sets the TaskID field.
func (e *EventTaskFailed) WithTaskID(id immutable.TaskID) *EventTaskFailed {
	e.TaskID = id
	return e
}
