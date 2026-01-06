package events_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable/events"
)

func Test_EventToolCallExecuted_Topic_Should_ReturnCorrectTopic(t *testing.T) {
	// Arrange
	evt := events.NewEventToolCallExecuted()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must match", topic, "agent.tool_call_executed")
}

func Test_EventToolCallExecuted_WithFields_Should_SetFields(t *testing.T) {
	// Arrange & Act
	evt := events.NewEventToolCallExecuted().
		WithAgentID("agent-1").
		WithTaskID("task-1").
		WithToolCallID("tc-1").
		WithToolName("search").
		WithSuccess(true)

	// Assert
	assert.That(t, "agent ID must match", evt.AgentID, immutable.AgentID("agent-1"))
	assert.That(t, "task ID must match", evt.TaskID, immutable.TaskID("task-1"))
	assert.That(t, "tool call ID must match", evt.ToolCallID, immutable.ToolCallID("tc-1"))
	assert.That(t, "tool name must match", evt.ToolName, "search")
	assert.That(t, "success must match", evt.Success, true)
}
