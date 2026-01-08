package agent_test

import (
	"context"
	"testing"

	"github.com/andygeiss/go-agent/pkg/agent"
	"github.com/andygeiss/go-agent/pkg/agent/events"
)

func Benchmark_TaskService_DirectCompletion(b *testing.B) {
	mockLLM := &mockLLMClient{
		response: agent.NewLLMResponse(
			agent.NewMessage(agent.RoleAssistant, "Response"),
			"stop",
		),
	}
	mockExecutor := &mockToolExecutor{}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher)
	ctx := context.Background()

	for b.Loop() {
		ag := agent.NewAgent("agent-1", "prompt")
		task := agent.NewTask("task-1", "Task", "input")
		_, _ = sut.RunTask(ctx, &ag, task)
	}
}

func Benchmark_TaskService_SingleToolCall(b *testing.B) {
	callCount := 0
	mockLLM := &mockLLMClient{
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			callCount++
			if callCount%2 == 1 {
				return agent.NewLLMResponse(
					agent.NewMessage(agent.RoleAssistant, ""),
					"tool_calls",
				).WithToolCalls([]agent.ToolCall{
					agent.NewToolCall("tc-1", "search", `{}`),
				})
			}
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, "done"),
				"stop",
			)
		},
	}
	mockExecutor := &mockToolExecutor{result: "result"}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher)
	ctx := context.Background()

	for b.Loop() {
		callCount = 0
		ag := agent.NewAgent("agent-1", "prompt")
		task := agent.NewTask("task-1", "Task", "input")
		_, _ = sut.RunTask(ctx, &ag, task)
	}
}

func Benchmark_TaskService_MultipleToolCalls_Sequential(b *testing.B) {
	callCount := 0
	mockLLM := &mockLLMClient{
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			callCount++
			if callCount%2 == 1 {
				return agent.NewLLMResponse(
					agent.NewMessage(agent.RoleAssistant, ""),
					"tool_calls",
				).WithToolCalls([]agent.ToolCall{
					agent.NewToolCall("tc-1", "search", `{}`),
					agent.NewToolCall("tc-2", "search", `{}`),
					agent.NewToolCall("tc-3", "search", `{}`),
				})
			}
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, "done"),
				"stop",
			)
		},
	}
	mockExecutor := &mockToolExecutor{result: "result"}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher)
	ctx := context.Background()

	for b.Loop() {
		callCount = 0
		ag := agent.NewAgent("agent-1", "prompt")
		task := agent.NewTask("task-1", "Task", "input")
		_, _ = sut.RunTask(ctx, &ag, task)
	}
}

func Benchmark_TaskService_MultipleToolCalls_Parallel(b *testing.B) {
	callCount := 0
	mockLLM := &mockLLMClient{
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			callCount++
			if callCount%2 == 1 {
				return agent.NewLLMResponse(
					agent.NewMessage(agent.RoleAssistant, ""),
					"tool_calls",
				).WithToolCalls([]agent.ToolCall{
					agent.NewToolCall("tc-1", "search", `{}`),
					agent.NewToolCall("tc-2", "search", `{}`),
					agent.NewToolCall("tc-3", "search", `{}`),
				})
			}
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, "done"),
				"stop",
			)
		},
	}
	mockExecutor := &mockToolExecutor{result: "result"}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher).
		WithParallelToolExecution()
	ctx := context.Background()

	for b.Loop() {
		callCount = 0
		ag := agent.NewAgent("agent-1", "prompt")
		task := agent.NewTask("task-1", "Task", "input")
		_, _ = sut.RunTask(ctx, &ag, task)
	}
}

func Benchmark_Message_Create(b *testing.B) {
	for b.Loop() {
		_ = agent.NewMessage(agent.RoleUser, "Hello, how are you?")
	}
}

func Benchmark_Agent_Create(b *testing.B) {
	for b.Loop() {
		_ = agent.NewAgent("agent-1", "You are a helpful assistant")
	}
}

func Benchmark_Event_Create(b *testing.B) {
	for b.Loop() {
		_ = events.NewEventTaskStarted("task-1", "TaskName")
	}
}
