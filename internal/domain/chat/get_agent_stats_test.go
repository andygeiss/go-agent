package chat_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/chat"
	"github.com/andygeiss/go-agent/pkg/agent"
)

func Test_GetAgentStatsUseCase_Execute_Should_ReturnCorrectStats(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("test-agent", "test prompt",
		agent.WithMaxIterations(20),
		agent.WithMaxMessages(100),
		agent.WithMetadata(agent.Metadata{"model": "gpt-4"}),
	)
	ag.AddMessage(agent.NewMessage(agent.RoleUser, "Hello"))
	ag.AddMessage(agent.NewMessage(agent.RoleAssistant, "Hi!"))
	uc := chat.NewGetAgentStatsUseCase(&ag)

	// Act
	stats := uc.Execute()

	// Assert
	assert.That(t, "agent ID must match", stats.AgentID, "test-agent")
	assert.That(t, "message count must be 2", stats.MessageCount, 2)
	assert.That(t, "max iterations must be 20", stats.MaxIterations, 20)
	assert.That(t, "max messages must be 100", stats.MaxMessages, 100)
	assert.That(t, "model must be gpt-4", stats.Model, "gpt-4")
}

func Test_GetAgentStatsUseCase_Execute_With_NoMetadata_Should_ReturnEmptyModel(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("test-agent", "test prompt")
	uc := chat.NewGetAgentStatsUseCase(&ag)

	// Act
	stats := uc.Execute()

	// Assert
	assert.That(t, "model must be empty", stats.Model, "")
}

func Test_GetAgentStatsUseCase_Execute_With_Tasks_Should_ReturnTaskCounts(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("test-agent", "test prompt")

	// Add completed task
	completedTask := agent.NewTask("task-1", "chat", "input")
	completedTask.Complete("output")
	ag.AddTask(completedTask)

	// Add failed task
	failedTask := agent.NewTask("task-2", "chat", "input")
	failedTask.Fail("error")
	ag.AddTask(failedTask)

	uc := chat.NewGetAgentStatsUseCase(&ag)

	// Act
	stats := uc.Execute()

	// Assert
	assert.That(t, "task count must be 2", stats.TaskCount, 2)
	assert.That(t, "completed tasks must be 1", stats.CompletedTasks, 1)
	assert.That(t, "failed tasks must be 1", stats.FailedTasks, 1)
}
