package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent/aggregates"
	"github.com/andygeiss/go-agent/internal/domain/agent/entities"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
	"github.com/andygeiss/go-agent/internal/domain/agent/services"
)

func Test_TaskService_RunTask_With_DirectCompletion_Should_Succeed(t *testing.T) {
	// Arrange
	mockLLM := &mockLLMClient{
		response: aggregates.NewLLMResponse(
			entities.NewMessage(immutable.RoleAssistant, "Here is the answer"),
			"stop",
		),
	}
	mockExecutor := &mockToolExecutor{}
	mockPublisher := &mockEventPublisher{}
	sut := services.NewTaskService(mockLLM, mockExecutor, mockPublisher)
	ag := aggregates.NewAgent("agent-1", "You are helpful")
	task := entities.NewTask("task-1", "Answer Question", "What is 2+2?")
	ctx := context.Background()

	// Act
	result, err := sut.RunTask(ctx, &ag, task)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "result must be successful", result.Success, true)
	assert.That(t, "result output must match", result.Output, "Here is the answer")
}

func Test_TaskService_RunTask_With_ToolCall_Should_ExecuteToolAndComplete(t *testing.T) {
	// Arrange
	callCount := 0
	mockLLM := &mockLLMClient{
		responseFn: func(_ []entities.Message) aggregates.LLMResponse {
			callCount++
			if callCount == 1 {
				return aggregates.NewLLMResponse(
					entities.NewMessage(immutable.RoleAssistant, ""),
					"tool_calls",
				).WithToolCalls([]entities.ToolCall{
					entities.NewToolCall("tc-1", "search", `{"query":"test"}`),
				})
			}
			return aggregates.NewLLMResponse(
				entities.NewMessage(immutable.RoleAssistant, "Based on the search: answer"),
				"stop",
			)
		},
	}
	mockExecutor := &mockToolExecutor{
		result: "search result",
	}
	mockPublisher := &mockEventPublisher{}
	sut := services.NewTaskService(mockLLM, mockExecutor, mockPublisher)
	ag := aggregates.NewAgent("agent-1", "You are helpful")
	task := entities.NewTask("task-1", "Search Task", "Find something")
	ctx := context.Background()

	// Act
	result, err := sut.RunTask(ctx, &ag, task)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "result must be successful", result.Success, true)
	assert.That(t, "tool executor must be called", mockExecutor.called, true)
	assert.That(t, "LLM must be called twice", callCount, 2)
}

func Test_TaskService_RunTask_With_MaxIterations_Should_Fail(t *testing.T) {
	// Arrange
	mockLLM := &mockLLMClient{
		responseFn: func(_ []entities.Message) aggregates.LLMResponse {
			return aggregates.NewLLMResponse(
				entities.NewMessage(immutable.RoleAssistant, ""),
				"tool_calls",
			).WithToolCalls([]entities.ToolCall{
				entities.NewToolCall("tc-1", "loop_tool", `{}`),
			})
		},
	}
	mockExecutor := &mockToolExecutor{result: "loop result"}
	mockPublisher := &mockEventPublisher{}
	sut := services.NewTaskService(mockLLM, mockExecutor, mockPublisher)

	ag := aggregates.NewAgent("agent-1", "prompt")
	ag.WithMaxIterations(3)
	task := entities.NewTask("task-1", "Infinite Loop", "loop")
	ctx := context.Background()

	// Act
	result, err := sut.RunTask(ctx, &ag, task)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "result must not be successful", result.Success, false)
	assert.That(t, "error must indicate max iterations", result.Error, "max iterations reached")
}

func Test_TaskService_RunTask_With_LLMError_Should_FailTask(t *testing.T) {
	// Arrange
	mockLLM := &mockLLMClient{
		err: errors.New("LLM connection failed"),
	}
	mockExecutor := &mockToolExecutor{}
	mockPublisher := &mockEventPublisher{}
	sut := services.NewTaskService(mockLLM, mockExecutor, mockPublisher)

	ag := aggregates.NewAgent("agent-1", "prompt")
	task := entities.NewTask("task-1", "Fail Task", "input")
	ctx := context.Background()

	// Act
	result, err := sut.RunTask(ctx, &ag, task)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "result must not be successful", result.Success, false)
	assert.That(t, "error must contain LLM error", result.Error, "LLM connection failed")
}

func Test_TaskService_RunTask_Should_PublishEvents(t *testing.T) {
	// Arrange
	mockLLM := &mockLLMClient{
		response: aggregates.NewLLMResponse(
			entities.NewMessage(immutable.RoleAssistant, "done"),
			"stop",
		),
	}
	mockExecutor := &mockToolExecutor{}
	mockPublisher := &mockEventPublisher{}
	sut := services.NewTaskService(mockLLM, mockExecutor, mockPublisher)

	ag := aggregates.NewAgent("agent-1", "prompt")
	task := entities.NewTask("task-1", "Task", "input")
	ctx := context.Background()

	// Act
	_, err := sut.RunTask(ctx, &ag, task)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "events must be published", len(mockPublisher.events) >= 2, true)
}
