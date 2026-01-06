package events_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable/events"
)

func Test_EventTaskCompleted_Topic_Should_ReturnCorrectTopic(t *testing.T) {
	// Arrange
	evt := events.NewEventTaskCompleted()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must match", topic, "agent.task_completed")
}

func Test_EventTaskCompleted_WithFields_Should_SetFields(t *testing.T) {
	// Arrange & Act
	evt := events.NewEventTaskCompleted().
		WithAgentID("agent-1").
		WithTaskID("task-1").
		WithName("Test Task").
		WithOutput("result").
		WithIterations(3)

	// Assert
	assert.That(t, "agent ID must match", evt.AgentID, immutable.AgentID("agent-1"))
	assert.That(t, "task ID must match", evt.TaskID, immutable.TaskID("task-1"))
	assert.That(t, "name must match", evt.Name, "Test Task")
	assert.That(t, "output must match", evt.Output, "result")
	assert.That(t, "iterations must match", evt.Iterations, 3)
}
