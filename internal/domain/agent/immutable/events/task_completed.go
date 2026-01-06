package events

import "github.com/andygeiss/go-agent/internal/domain/agent/immutable"

// EventTaskCompleted is published when a task finishes successfully.
// It includes the final output and iteration count for metrics/logging.
type EventTaskCompleted struct {
	Name       string            `json:"name"`
	Output     string            `json:"output"`
	AgentID    immutable.AgentID `json:"agent_id"`
	TaskID     immutable.TaskID  `json:"task_id"`
	Iterations int               `json:"iterations"`
}

// NewEventTaskCompleted creates a new EventTaskCompleted instance.
func NewEventTaskCompleted() *EventTaskCompleted {
	return &EventTaskCompleted{}
}

// Topic returns the topic for the event.
func (e *EventTaskCompleted) Topic() string {
	return EventTopicTaskCompleted
}

// WithAgentID sets the AgentID field.
func (e *EventTaskCompleted) WithAgentID(id immutable.AgentID) *EventTaskCompleted {
	e.AgentID = id
	return e
}

// WithIterations sets the Iterations field.
func (e *EventTaskCompleted) WithIterations(iterations int) *EventTaskCompleted {
	e.Iterations = iterations
	return e
}

// WithName sets the Name field.
func (e *EventTaskCompleted) WithName(name string) *EventTaskCompleted {
	e.Name = name
	return e
}

// WithOutput sets the Output field.
func (e *EventTaskCompleted) WithOutput(output string) *EventTaskCompleted {
	e.Output = output
	return e
}

// WithTaskID sets the TaskID field.
func (e *EventTaskCompleted) WithTaskID(id immutable.TaskID) *EventTaskCompleted {
	e.TaskID = id
	return e
}
