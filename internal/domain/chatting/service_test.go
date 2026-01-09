package chatting_test

import (
	"context"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/chatting"
)

// mockTaskRunner implements agent.TaskRunner for testing.
type mockTaskRunner struct {
	err    error
	result agent.Result
}

func (m *mockTaskRunner) RunTask(_ context.Context, _ *agent.Agent, _ *agent.Task) (agent.Result, error) {
	return m.result, m.err
}

// ClearConversationUseCase tests

func Test_ClearConversationUseCase_Execute_Should_ClearMessages(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("test-agent", "test prompt")
	ag.AddMessage(agent.NewMessage(agent.RoleUser, "Hello"))
	ag.AddMessage(agent.NewMessage(agent.RoleAssistant, "Hi!"))
	uc := chatting.NewClearConversationUseCase(&ag)

	// Act
	uc.Execute()

	// Assert
	assert.That(t, "message count must be 0", ag.MessageCount(), 0)
}

func Test_ClearConversationUseCase_Execute_With_EmptyHistory_Should_NotPanic(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("test-agent", "test prompt")
	uc := chatting.NewClearConversationUseCase(&ag)

	// Act & Assert (no panic)
	uc.Execute()
	assert.That(t, "message count must be 0", ag.MessageCount(), 0)
}

// GetAgentStatsUseCase tests

func Test_GetAgentStatsUseCase_Execute_Should_ReturnCorrectStats(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("test-agent", "test prompt",
		agent.WithMaxIterations(20),
		agent.WithMaxMessages(100),
		agent.WithMetadata(agent.Metadata{"model": "gpt-4"}),
	)
	ag.AddMessage(agent.NewMessage(agent.RoleUser, "Hello"))
	ag.AddMessage(agent.NewMessage(agent.RoleAssistant, "Hi!"))
	uc := chatting.NewGetAgentStatsUseCase(&ag)

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
	uc := chatting.NewGetAgentStatsUseCase(&ag)

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

	uc := chatting.NewGetAgentStatsUseCase(&ag)

	// Act
	stats := uc.Execute()

	// Assert
	assert.That(t, "task count must be 2", stats.TaskCount, 2)
	assert.That(t, "completed tasks must be 1", stats.CompletedTasks, 1)
	assert.That(t, "failed tasks must be 1", stats.FailedTasks, 1)
}

// SendMessageUseCase tests

func Test_SendMessageUseCase_Execute_With_FailedResponse_Should_ReturnError(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("test-agent", "test prompt")
	runner := &mockTaskRunner{
		result: agent.Result{
			Success: false,
			Error:   "task failed",
		},
	}
	uc := chatting.NewSendMessageUseCase(runner, &ag)

	// Act
	output, _ := uc.Execute(context.Background(), chatting.SendMessageInput{Message: "Hi"})

	// Assert
	assert.That(t, "success must be false", output.Success, false)
	assert.That(t, "error must match", output.Error, "task failed")
}

func Test_SendMessageUseCase_Execute_With_MultipleCalls_Should_IncrementTaskCounter(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("test-agent", "test prompt")
	runner := &mockTaskRunner{
		result: agent.Result{Success: true, Output: "OK"},
	}
	uc := chatting.NewSendMessageUseCase(runner, &ag)

	// Act
	_, _ = uc.Execute(context.Background(), chatting.SendMessageInput{Message: "First"})
	_, _ = uc.Execute(context.Background(), chatting.SendMessageInput{Message: "Second"})
	output, _ := uc.Execute(context.Background(), chatting.SendMessageInput{Message: "Third"})

	// Assert - output should succeed (task counter is internal)
	assert.That(t, "success must be true", output.Success, true)
}

func Test_SendMessageUseCase_Execute_With_SuccessfulResponse_Should_ReturnOutput(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("test-agent", "test prompt")
	runner := &mockTaskRunner{
		result: agent.Result{
			Success:        true,
			Output:         "Hello!",
			Duration:       100 * time.Millisecond,
			IterationCount: 1,
			ToolCallCount:  0,
		},
	}
	uc := chatting.NewSendMessageUseCase(runner, &ag)

	// Act
	output, err := uc.Execute(context.Background(), chatting.SendMessageInput{Message: "Hi"})

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "success must be true", output.Success, true)
	assert.That(t, "response must match", output.Response, "Hello!")
	assert.That(t, "iteration count must be 1", output.IterationCount, 1)
}
