package events

import "github.com/andygeiss/go-agent/internal/domain/agent/immutable"

// EventTaskStarted represents a task started event.
type EventTaskStarted struct {
	AgentID immutable.AgentID `json:"agent_id"`
	TaskID  immutable.TaskID  `json:"task_id"`
	Name    string            `json:"name"`
}

// NewEventTaskStarted creates a new EventTaskStarted instance.
func NewEventTaskStarted() *EventTaskStarted {
	return &EventTaskStarted{}
}

// Topic returns the topic for the event.
func (e *EventTaskStarted) Topic() string {
	return EventTopicTaskStarted
}

// WithAgentID sets the AgentID field.
func (e *EventTaskStarted) WithAgentID(id immutable.AgentID) *EventTaskStarted {
	e.AgentID = id
	return e
}

// WithName sets the Name field.
func (e *EventTaskStarted) WithName(name string) *EventTaskStarted {
	e.Name = name
	return e
}

// WithTaskID sets the TaskID field.
func (e *EventTaskStarted) WithTaskID(id immutable.TaskID) *EventTaskStarted {
	e.TaskID = id
	return e
}
