package events_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable/events"
)

func Test_EventTaskFailed_Topic_Should_ReturnCorrectTopic(t *testing.T) {
	// Arrange
	evt := events.NewEventTaskFailed()

	// Act
	topic := evt.Topic()

	// Assert
	assert.That(t, "topic must match", topic, "agent.task_failed")
}

func Test_EventTaskFailed_WithFields_Should_SetFields(t *testing.T) {
	// Arrange & Act
	evt := events.NewEventTaskFailed().
		WithAgentID("agent-1").
		WithTaskID("task-1").
		WithName("Test Task").
		WithError("something went wrong").
		WithIterations(5)

	// Assert
	assert.That(t, "agent ID must match", evt.AgentID, immutable.AgentID("agent-1"))
	assert.That(t, "task ID must match", evt.TaskID, immutable.TaskID("task-1"))
	assert.That(t, "name must match", evt.Name, "Test Task")
	assert.That(t, "error must match", evt.Error, "something went wrong")
	assert.That(t, "iterations must match", evt.Iterations, 5)
}
