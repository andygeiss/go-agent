package agent

// Event topic constants for messaging (alphabetically sorted).
const (
	TopicTaskCompleted    = "agent.task.completed"
	TopicTaskFailed       = "agent.task.failed"
	TopicTaskStarted      = "agent.task.started"
	TopicToolCallExecuted = "agent.toolcall.executed"
)

// EventTaskCompleted is emitted when a task finishes successfully.
type EventTaskCompleted struct {
	Output string `json:"output"`
	TaskID string `json:"task_id"`
}

// NewEventTaskCompleted creates a new task completed event.
func NewEventTaskCompleted(taskID, output string) EventTaskCompleted {
	return EventTaskCompleted{
		Output: output,
		TaskID: taskID,
	}
}

// Topic returns the event topic for messaging.
func (e EventTaskCompleted) Topic() string {
	return TopicTaskCompleted
}

// EventTaskFailed is emitted when a task terminates with an error.
type EventTaskFailed struct {
	Error  string `json:"error"`
	TaskID string `json:"task_id"`
}

// NewEventTaskFailed creates a new task failed event.
func NewEventTaskFailed(taskID, errMsg string) EventTaskFailed {
	return EventTaskFailed{
		Error:  errMsg,
		TaskID: taskID,
	}
}

// Topic returns the event topic for messaging.
func (e EventTaskFailed) Topic() string {
	return TopicTaskFailed
}

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

// EventToolCallExecuted is emitted after a tool call completes.
type EventToolCallExecuted struct {
	Error      string `json:"error,omitempty"`
	Result     string `json:"result"`
	ToolCallID string `json:"tool_call_id"`
	ToolName   string `json:"tool_name"`
}

// NewEventToolCallExecuted creates a new tool call executed event.
func NewEventToolCallExecuted(toolCallID, toolName, result, errMsg string) EventToolCallExecuted {
	return EventToolCallExecuted{
		Error:      errMsg,
		Result:     result,
		ToolCallID: toolCallID,
		ToolName:   toolName,
	}
}

// Topic returns the event topic for messaging.
func (e EventToolCallExecuted) Topic() string {
	return TopicToolCallExecuted
}
