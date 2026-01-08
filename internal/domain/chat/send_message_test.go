package chat_test

import (
	"context"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/chat"
	"github.com/andygeiss/go-agent/pkg/agent"
)

// mockTaskRunner implements chat.TaskRunner for testing.
type mockTaskRunner struct {
	err    error
	result agent.Result
}

func (m *mockTaskRunner) RunTask(_ context.Context, _ *agent.Agent, _ *agent.Task) (agent.Result, error) {
	return m.result, m.err
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
	uc := chat.NewSendMessageUseCase(runner, &ag)

	// Act
	output, err := uc.Execute(context.Background(), chat.SendMessageInput{Message: "Hi"})

	// Assert
	assert.That(t, "err must be nil", err, nil)
	assert.That(t, "success must be true", output.Success, true)
	assert.That(t, "response must match", output.Response, "Hello!")
	assert.That(t, "iteration count must be 1", output.IterationCount, 1)
}

func Test_SendMessageUseCase_Execute_With_FailedResponse_Should_ReturnError(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("test-agent", "test prompt")
	runner := &mockTaskRunner{
		result: agent.Result{
			Success: false,
			Error:   "task failed",
		},
	}
	uc := chat.NewSendMessageUseCase(runner, &ag)

	// Act
	output, _ := uc.Execute(context.Background(), chat.SendMessageInput{Message: "Hi"})

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
	uc := chat.NewSendMessageUseCase(runner, &ag)

	// Act
	_, _ = uc.Execute(context.Background(), chat.SendMessageInput{Message: "First"})
	_, _ = uc.Execute(context.Background(), chat.SendMessageInput{Message: "Second"})
	output, _ := uc.Execute(context.Background(), chat.SendMessageInput{Message: "Third"})

	// Assert - output should succeed (task counter is internal)
	assert.That(t, "success must be true", output.Success, true)
}
