package events

// Event topic constants for the agent domain.
// These are used by the event publisher to route events to subscribers.
const (
	EventTopicTaskCompleted    = "agent.task_completed"     // Published when a task finishes successfully
	EventTopicTaskFailed       = "agent.task_failed"        // Published when a task terminates with an error
	EventTopicTaskStarted      = "agent.task_started"       // Published when a task begins execution
	EventTopicToolCallExecuted = "agent.tool_call_executed" // Published after each tool call completes
)
