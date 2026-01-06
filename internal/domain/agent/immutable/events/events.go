package events

// This file contains the event topic constants for the agent domain.

const (
	EventTopicTaskCompleted    = "agent.task_completed"
	EventTopicTaskFailed       = "agent.task_failed"
	EventTopicTaskStarted      = "agent.task_started"
	EventTopicToolCallExecuted = "agent.tool_call_executed"
)
