package events_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/pkg/agent/events"
)

func Test_EventTaskStarted_Topic_Should_ReturnCorrectTopic(t *testing.T) {
	// Arrange
	event := events.NewEventTaskStarted("task-1", "Test Task")

	// Act
	topic := event.Topic()

	// Assert
	assert.That(t, "topic must match", topic, events.TopicTaskStarted)
}

func Test_EventTaskStarted_Fields_Should_MatchInputs(t *testing.T) {
	// Arrange & Act
	event := events.NewEventTaskStarted("task-1", "Test Task")

	// Assert
	assert.That(t, "task ID must match", event.TaskID, "task-1")
	assert.That(t, "task name must match", event.TaskName, "Test Task")
}

func Test_EventTaskCompleted_Topic_Should_ReturnCorrectTopic(t *testing.T) {
	// Arrange
	event := events.NewEventTaskCompleted("task-1", "output")

	// Act
	topic := event.Topic()

	// Assert
	assert.That(t, "topic must match", topic, events.TopicTaskCompleted)
}

func Test_EventTaskCompleted_Fields_Should_MatchInputs(t *testing.T) {
	// Arrange & Act
	event := events.NewEventTaskCompleted("task-1", "output")

	// Assert
	assert.That(t, "task ID must match", event.TaskID, "task-1")
	assert.That(t, "output must match", event.Output, "output")
}

func Test_EventTaskFailed_Topic_Should_ReturnCorrectTopic(t *testing.T) {
	// Arrange
	event := events.NewEventTaskFailed("task-1", "error message")

	// Act
	topic := event.Topic()

	// Assert
	assert.That(t, "topic must match", topic, events.TopicTaskFailed)
}

func Test_EventTaskFailed_Fields_Should_MatchInputs(t *testing.T) {
	// Arrange & Act
	event := events.NewEventTaskFailed("task-1", "error message")

	// Assert
	assert.That(t, "task ID must match", event.TaskID, "task-1")
	assert.That(t, "error must match", event.Error, "error message")
}

func Test_EventToolCallExecuted_Topic_Should_ReturnCorrectTopic(t *testing.T) {
	// Arrange
	event := events.NewEventToolCallExecuted("tc-1", "search", "result", "")

	// Act
	topic := event.Topic()

	// Assert
	assert.That(t, "topic must match", topic, events.TopicToolCallExecuted)
}

func Test_EventToolCallExecuted_Fields_Should_MatchInputs(t *testing.T) {
	// Arrange & Act
	event := events.NewEventToolCallExecuted("tc-1", "search", "result", "error")

	// Assert
	assert.That(t, "tool call ID must match", event.ToolCallID, "tc-1")
	assert.That(t, "tool name must match", event.ToolName, "search")
	assert.That(t, "result must match", event.Result, "result")
	assert.That(t, "error must match", event.Error, "error")
}
