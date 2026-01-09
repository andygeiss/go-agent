package agent

import (
	"context"

	"github.com/andygeiss/cloud-native-utils/event"
)

// EventPublisher is the interface for publishing domain events.
type EventPublisher interface {
	// Publish sends an event to subscribers.
	Publish(ctx context.Context, e event.Event) error
}

// Event topic constants for messaging.
const (
	TopicTaskCompleted    = "agent.task.completed"
	TopicTaskFailed       = "agent.task.failed"
	TopicTaskStarted      = "agent.task.started"
	TopicToolCallExecuted = "agent.toolcall.executed"
)

// EventTaskStarted is emitted when a task begins execution.
type EventTaskStarted struct {
	TaskID   string `json:"task_id"`
	TaskName string `json:"task_name"`
}

// NewEventTaskStarted creates a new task started event.
func NewEventTaskStarted(taskID, taskName string) EventTaskStarted {
	return EventTaskStarted{
		TaskID:   taskID,
		TaskName: taskName,
	}
}

// Topic returns the event topic for messaging.
func (e EventTaskStarted) Topic() string {
	return TopicTaskStarted
}

// EventTaskCompleted is emitted when a task finishes successfully.
type EventTaskCompleted struct {
	TaskID string `json:"task_id"`
	Output string `json:"output"`
}

// NewEventTaskCompleted creates a new task completed event.
func NewEventTaskCompleted(taskID, output string) EventTaskCompleted {
	return EventTaskCompleted{
		TaskID: taskID,
		Output: output,
	}
}

// Topic returns the event topic for messaging.
func (e EventTaskCompleted) Topic() string {
	return TopicTaskCompleted
}

// EventTaskFailed is emitted when a task terminates with an error.
type EventTaskFailed struct {
	TaskID string `json:"task_id"`
	Error  string `json:"error"`
}

// NewEventTaskFailed creates a new task failed event.
func NewEventTaskFailed(taskID, errMsg string) EventTaskFailed {
	return EventTaskFailed{
		TaskID: taskID,
		Error:  errMsg,
	}
}

// Topic returns the event topic for messaging.
func (e EventTaskFailed) Topic() string {
	return TopicTaskFailed
}

// EventToolCallExecuted is emitted after a tool call completes.
type EventToolCallExecuted struct {
	ToolCallID string `json:"tool_call_id"`
	ToolName   string `json:"tool_name"`
	Result     string `json:"result"`
	Error      string `json:"error,omitempty"`
}

// NewEventToolCallExecuted creates a new tool call executed event.
func NewEventToolCallExecuted(toolCallID, toolName, result, errMsg string) EventToolCallExecuted {
	return EventToolCallExecuted{
		ToolCallID: toolCallID,
		ToolName:   toolName,
		Result:     result,
		Error:      errMsg,
	}
}

// Topic returns the event topic for messaging.
func (e EventToolCallExecuted) Topic() string {
	return TopicToolCallExecuted
}
