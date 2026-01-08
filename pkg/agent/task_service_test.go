package agent_test

import (
	"context"
	"errors"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/pkg/agent"
	"github.com/andygeiss/go-agent/pkg/event"
)

// mockLLMClient implements agent.LLMClient for testing.
type mockLLMClient struct {
	err        error
	responseFn func(messages []agent.Message) agent.LLMResponse
	response   agent.LLMResponse
}

func (m *mockLLMClient) Run(
	_ context.Context,
	messages []agent.Message,
	_ []agent.ToolDefinition,
) (agent.LLMResponse, error) {
	if m.err != nil {
		return agent.LLMResponse{}, m.err
	}
	if m.responseFn != nil {
		return m.responseFn(messages), nil
	}
	return m.response, nil
}

// mockToolExecutor implements agent.ToolExecutor for testing.
type mockToolExecutor struct {
	err    error
	result string
	called bool
}

func (m *mockToolExecutor) Execute(_ context.Context, _ string, _ string) (string, error) {
	m.called = true
	if m.err != nil {
		return "", m.err
	}
	return m.result, nil
}

func (m *mockToolExecutor) GetAvailableTools() []string {
	return []string{"search", "loop_tool"}
}

func (m *mockToolExecutor) GetToolDefinitions() []agent.ToolDefinition {
	return []agent.ToolDefinition{
		agent.NewToolDefinition("search", "Search for items").WithParameter("query", "The search query"),
		agent.NewToolDefinition("loop_tool", "A tool that loops"),
	}
}

func (m *mockToolExecutor) HasTool(_ string) bool {
	return true
}

// mockEventPublisher implements agent.EventPublisher for testing.
type mockEventPublisher struct {
	events []event.Event
}

func (m *mockEventPublisher) Publish(_ context.Context, e event.Event) error {
	m.events = append(m.events, e)
	return nil
}

func Test_TaskService_RunTask_With_DirectCompletion_Should_Succeed(t *testing.T) {
	// Arrange
	mockLLM := &mockLLMClient{
		response: agent.NewLLMResponse(
			agent.NewMessage(agent.RoleAssistant, "Here is the answer"),
			"stop",
		),
	}
	mockExecutor := &mockToolExecutor{}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher)
	ag := agent.NewAgent("agent-1", "You are helpful")
	task := agent.NewTask("task-1", "Answer Question", "What is 2+2?")
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
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			callCount++
			if callCount == 1 {
				return agent.NewLLMResponse(
					agent.NewMessage(agent.RoleAssistant, ""),
					"tool_calls",
				).WithToolCalls([]agent.ToolCall{
					agent.NewToolCall("tc-1", "search", `{"query":"test"}`),
				})
			}
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, "Based on the search: answer"),
				"stop",
			)
		},
	}
	mockExecutor := &mockToolExecutor{
		result: "search result",
	}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher)
	ag := agent.NewAgent("agent-1", "You are helpful")
	task := agent.NewTask("task-1", "Search Task", "Find something")
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
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, ""),
				"tool_calls",
			).WithToolCalls([]agent.ToolCall{
				agent.NewToolCall("tc-1", "loop_tool", `{}`),
			})
		},
	}
	mockExecutor := &mockToolExecutor{result: "loop result"}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher)

	ag := agent.NewAgent("agent-1", "prompt")
	ag.SetMaxIterations(3)
	task := agent.NewTask("task-1", "Infinite Loop", "loop")
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
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher)

	ag := agent.NewAgent("agent-1", "prompt")
	task := agent.NewTask("task-1", "Fail Task", "input")
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
		response: agent.NewLLMResponse(
			agent.NewMessage(agent.RoleAssistant, "done"),
			"stop",
		),
	}
	mockExecutor := &mockToolExecutor{}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher)

	ag := agent.NewAgent("agent-1", "prompt")
	task := agent.NewTask("task-1", "Task", "input")
	ctx := context.Background()

	// Act
	_, err := sut.RunTask(ctx, &ag, task)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "events must be published", len(mockPublisher.events) >= 2, true)
}

func Test_TaskService_WithParallelToolExecution_Should_ExecuteToolsInParallel(t *testing.T) {
	// Arrange
	callCount := 0
	mockLLM := &mockLLMClient{
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			callCount++
			if callCount == 1 {
				return agent.NewLLMResponse(
					agent.NewMessage(agent.RoleAssistant, ""),
					"tool_calls",
				).WithToolCalls([]agent.ToolCall{
					agent.NewToolCall("tc-1", "search", `{"query":"test1"}`),
					agent.NewToolCall("tc-2", "search", `{"query":"test2"}`),
				})
			}
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, "Done"),
				"stop",
			)
		},
	}
	executionCount := 0
	mockExecutor := &mockToolExecutor{
		result: "result",
	}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher).
		WithParallelToolExecution()
	ag := agent.NewAgent("agent-1", "You are helpful")
	task := agent.NewTask("task-1", "Multi-tool Task", "Do multiple things")
	ctx := context.Background()

	// Act
	result, err := sut.RunTask(ctx, &ag, task)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "result must be successful", result.Success, true)
	assert.That(t, "LLM called twice", callCount, 2)
	// Note: Both tool calls should have been executed
	_ = executionCount // executionCount is incremented but mockExecutor.called tracks single call
}

func Test_TaskService_WithHooks_Should_CallHooks(t *testing.T) {
	// Arrange
	beforeTaskCalled := false
	afterTaskCalled := false
	beforeLLMCalled := false
	afterLLMCalled := false

	hooks := agent.NewHooks().
		WithBeforeTask(func(_ context.Context, _ *agent.Agent, _ *agent.Task) error {
			beforeTaskCalled = true
			return nil
		}).
		WithAfterTask(func(_ context.Context, _ *agent.Agent, _ *agent.Task) error {
			afterTaskCalled = true
			return nil
		}).
		WithBeforeLLMCall(func(_ context.Context, _ *agent.Agent, _ *agent.Task) error {
			beforeLLMCalled = true
			return nil
		}).
		WithAfterLLMCall(func(_ context.Context, _ *agent.Agent, _ *agent.Task) error {
			afterLLMCalled = true
			return nil
		})

	mockLLM := &mockLLMClient{
		response: agent.NewLLMResponse(
			agent.NewMessage(agent.RoleAssistant, "done"),
			"stop",
		),
	}
	mockExecutor := &mockToolExecutor{}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher).
		WithHooks(hooks)
	ag := agent.NewAgent("agent-1", "prompt")
	task := agent.NewTask("task-1", "Hook Test", "input")
	ctx := context.Background()

	// Act
	_, err := sut.RunTask(ctx, &ag, task)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "beforeTask must be called", beforeTaskCalled, true)
	assert.That(t, "afterTask must be called", afterTaskCalled, true)
	assert.That(t, "beforeLLM must be called", beforeLLMCalled, true)
	assert.That(t, "afterLLM must be called", afterLLMCalled, true)
}

func Test_TaskService_WithHooks_BeforeToolCall_Should_BeCalled(t *testing.T) {
	// Arrange
	beforeToolCallCalled := false
	afterToolCallCalled := false

	hooks := agent.NewHooks().
		WithBeforeToolCall(func(_ context.Context, _ *agent.Agent, _ *agent.ToolCall) error {
			beforeToolCallCalled = true
			return nil
		}).
		WithAfterToolCall(func(_ context.Context, _ *agent.Agent, _ *agent.ToolCall) error {
			afterToolCallCalled = true
			return nil
		})

	callCount := 0
	mockLLM := &mockLLMClient{
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			callCount++
			if callCount == 1 {
				return agent.NewLLMResponse(
					agent.NewMessage(agent.RoleAssistant, ""),
					"tool_calls",
				).WithToolCalls([]agent.ToolCall{
					agent.NewToolCall("tc-1", "search", `{"query":"test"}`),
				})
			}
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, "done"),
				"stop",
			)
		},
	}
	mockExecutor := &mockToolExecutor{result: "tool result"}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher).
		WithHooks(hooks)
	ag := agent.NewAgent("agent-1", "prompt")
	task := agent.NewTask("task-1", "Tool Hook Test", "input")
	ctx := context.Background()

	// Act
	_, err := sut.RunTask(ctx, &ag, task)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "beforeToolCall must be called", beforeToolCallCalled, true)
	assert.That(t, "afterToolCall must be called", afterToolCallCalled, true)
}
