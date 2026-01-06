package events

import "github.com/andygeiss/go-agent/internal/domain/agent/immutable"

// EventToolCallExecuted represents a tool call executed event.
type EventToolCallExecuted struct {
	AgentID    immutable.AgentID    `json:"agent_id"`
	TaskID     immutable.TaskID     `json:"task_id"`
	ToolCallID immutable.ToolCallID `json:"tool_call_id"`
	ToolName   string               `json:"tool_name"`
	Success    bool                 `json:"success"`
}

// NewEventToolCallExecuted creates a new EventToolCallExecuted instance.
func NewEventToolCallExecuted() *EventToolCallExecuted {
	return &EventToolCallExecuted{}
}

// Topic returns the topic for the event.
func (e *EventToolCallExecuted) Topic() string {
	return EventTopicToolCallExecuted
}

// WithAgentID sets the AgentID field.
func (e *EventToolCallExecuted) WithAgentID(id immutable.AgentID) *EventToolCallExecuted {
	e.AgentID = id
	return e
}

// WithSuccess sets the Success field.
func (e *EventToolCallExecuted) WithSuccess(success bool) *EventToolCallExecuted {
	e.Success = success
	return e
}

// WithTaskID sets the TaskID field.
func (e *EventToolCallExecuted) WithTaskID(id immutable.TaskID) *EventToolCallExecuted {
	e.TaskID = id
	return e
}

// WithToolCallID sets the ToolCallID field.
func (e *EventToolCallExecuted) WithToolCallID(id immutable.ToolCallID) *EventToolCallExecuted {
	e.ToolCallID = id
	return e
}

// WithToolName sets the ToolName field.
func (e *EventToolCallExecuted) WithToolName(name string) *EventToolCallExecuted {
	e.ToolName = name
	return e
}
