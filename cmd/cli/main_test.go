package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/chatting"
	"github.com/andygeiss/go-agent/internal/domain/tooling"
)

// =============================================================================
// Benchmarks for Profile-Guided Optimization (PGO)
// =============================================================================
//
// These benchmarks exercise the hot paths of the CLI application to generate
// accurate CPU profiles for PGO builds. Run with: just profile
//
// The benchmarks cover:
// - Full use case execution (SendMessage)
// - Task service with various tool call patterns
// - Message handling with trimming
// - Real tool execution (calculate, time)
// - Parallel vs sequential tool execution
// =============================================================================

// -----------------------------------------------------------------------------
// Mock implementations for benchmarking
// -----------------------------------------------------------------------------

// mockLLMClient implements agent.LLMClient for benchmarking.
type mockLLMClient struct {
	responseFn func(msgs []agent.Message) agent.LLMResponse
	response   agent.LLMResponse
}

func (m *mockLLMClient) Run(_ context.Context, msgs []agent.Message, _ []agent.ToolDefinition) (agent.LLMResponse, error) {
	if m.responseFn != nil {
		return m.responseFn(msgs), nil
	}
	if m.response.Message.Content != "" || m.response.FinishReason != "" {
		return m.response, nil
	}
	return agent.LLMResponse{
		Message:      agent.NewMessage(agent.RoleAssistant, "Mock response for benchmarking"),
		FinishReason: "stop",
	}, nil
}

// mockToolExecutor implements agent.ToolExecutor for benchmarking.
type mockToolExecutor struct {
	result string
}

func (m *mockToolExecutor) Execute(_ context.Context, _ string, _ string) (string, error) {
	if m.result != "" {
		return m.result, nil
	}
	return "mock result", nil
}

func (m *mockToolExecutor) GetAvailableTools() []string {
	return []string{"calculate", "get_current_time"}
}

func (m *mockToolExecutor) GetToolDefinitions() []agent.ToolDefinition {
	return []agent.ToolDefinition{
		agent.NewToolDefinition("calculate", "Perform arithmetic calculation").
			WithParameter("expression", "The arithmetic expression to evaluate"),
		agent.NewToolDefinition("get_current_time", "Get the current date and time"),
	}
}

func (m *mockToolExecutor) HasTool(name string) bool {
	return name == "calculate" || name == "get_current_time"
}

func (m *mockToolExecutor) RegisterTool(_ string, _ agent.ToolFunc) {}

func (m *mockToolExecutor) RegisterToolDefinition(_ agent.ToolDefinition) {}

// mockEventPublisher implements agent.EventPublisher for benchmarking.
type mockEventPublisher struct{}

func (m *mockEventPublisher) Publish(_ context.Context, _ event.Event) error {
	return nil
}

// -----------------------------------------------------------------------------
// Use Case Benchmarks - Full execution path
// -----------------------------------------------------------------------------

// Benchmark_SendMessageUseCase_DirectCompletion benchmarks direct LLM responses.
func Benchmark_SendMessageUseCase_DirectCompletion(b *testing.B) {
	llmClient := &mockLLMClient{
		response: agent.NewLLMResponse(
			agent.NewMessage(agent.RoleAssistant, "Hello! How can I help you today?"),
			"stop",
		),
	}
	toolExecutor := &mockToolExecutor{}
	publisher := &mockEventPublisher{}
	taskService := agent.NewTaskService(llmClient, toolExecutor, publisher)

	ctx := context.Background()
	input := chatting.SendMessageInput{Message: "Hello"}

	b.ResetTimer()
	for b.Loop() {
		ag := agent.NewAgent("bench-agent", "You are a helpful assistant",
			agent.WithMaxIterations(10),
			agent.WithMaxMessages(100),
		)
		useCase := chatting.NewSendMessageUseCase(taskService, &ag)
		_, _ = useCase.Execute(ctx, input)
	}
}

// Benchmark_SendMessageUseCase_SingleToolCall benchmarks single tool execution.
func Benchmark_SendMessageUseCase_SingleToolCall(b *testing.B) {
	callCount := 0
	llmClient := &mockLLMClient{
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			callCount++
			if callCount%2 == 1 {
				return agent.NewLLMResponse(
					agent.NewMessage(agent.RoleAssistant, ""),
					"tool_calls",
				).WithToolCalls([]agent.ToolCall{
					agent.NewToolCall("tc-1", "calculate", `{"expression": "2 + 2"}`),
				})
			}
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, "The result is 4"),
				"stop",
			)
		},
	}
	toolExecutor := &mockToolExecutor{result: "4"}
	publisher := &mockEventPublisher{}
	taskService := agent.NewTaskService(llmClient, toolExecutor, publisher)

	ctx := context.Background()
	input := chatting.SendMessageInput{Message: "What is 2 + 2?"}

	b.ResetTimer()
	for b.Loop() {
		callCount = 0
		ag := agent.NewAgent("bench-agent", "You are a helpful assistant",
			agent.WithMaxIterations(10),
			agent.WithMaxMessages(100),
		)
		useCase := chatting.NewSendMessageUseCase(taskService, &ag)
		_, _ = useCase.Execute(ctx, input)
	}
}

// Benchmark_SendMessageUseCase_MultipleToolCalls benchmarks multiple tool executions.
func Benchmark_SendMessageUseCase_MultipleToolCalls(b *testing.B) {
	callCount := 0
	llmClient := &mockLLMClient{
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			callCount++
			if callCount%2 == 1 {
				return agent.NewLLMResponse(
					agent.NewMessage(agent.RoleAssistant, ""),
					"tool_calls",
				).WithToolCalls([]agent.ToolCall{
					agent.NewToolCall("tc-1", "calculate", `{"expression": "10 * 5"}`),
					agent.NewToolCall("tc-2", "calculate", `{"expression": "50 + 25"}`),
					agent.NewToolCall("tc-3", "get_current_time", `{}`),
				})
			}
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, "The calculations are complete"),
				"stop",
			)
		},
	}
	toolExecutor := &mockToolExecutor{result: "75"}
	publisher := &mockEventPublisher{}
	taskService := agent.NewTaskService(llmClient, toolExecutor, publisher)

	ctx := context.Background()
	input := chatting.SendMessageInput{Message: "Calculate 10*5, 50+25 and get time"}

	b.ResetTimer()
	for b.Loop() {
		callCount = 0
		ag := agent.NewAgent("bench-agent", "You are a helpful assistant",
			agent.WithMaxIterations(10),
			agent.WithMaxMessages(100),
		)
		useCase := chatting.NewSendMessageUseCase(taskService, &ag)
		_, _ = useCase.Execute(ctx, input)
	}
}

// -----------------------------------------------------------------------------
// Task Service Benchmarks - Core agent loop
// -----------------------------------------------------------------------------

// Benchmark_TaskService_DirectCompletion benchmarks direct task completion.
func Benchmark_TaskService_DirectCompletion(b *testing.B) {
	llmClient := &mockLLMClient{
		response: agent.NewLLMResponse(
			agent.NewMessage(agent.RoleAssistant, "Task completed"),
			"stop",
		),
	}
	toolExecutor := &mockToolExecutor{}
	publisher := &mockEventPublisher{}
	taskService := agent.NewTaskService(llmClient, toolExecutor, publisher)

	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		ag := agent.NewAgent("bench-agent", "You are a helpful assistant",
			agent.WithMaxIterations(10),
		)
		task := agent.NewTask("bench-task", "benchmark", "Test input")
		_, _ = taskService.RunTask(ctx, &ag, task)
	}
}

// Benchmark_TaskService_WithHooks benchmarks task execution with hooks enabled.
func Benchmark_TaskService_WithHooks(b *testing.B) {
	llmClient := &mockLLMClient{
		response: agent.NewLLMResponse(
			agent.NewMessage(agent.RoleAssistant, "Task completed"),
			"stop",
		),
	}
	toolExecutor := &mockToolExecutor{}
	publisher := &mockEventPublisher{}

	hooks := agent.NewHooks().
		WithBeforeTask(func(_ context.Context, _ *agent.Agent, _ *agent.Task) error { return nil }).
		WithAfterTask(func(_ context.Context, _ *agent.Agent, _ *agent.Task) error { return nil }).
		WithBeforeLLMCall(func(_ context.Context, _ *agent.Agent, _ *agent.Task) error { return nil }).
		WithAfterLLMCall(func(_ context.Context, _ *agent.Agent, _ *agent.Task) error { return nil })

	taskService := agent.NewTaskService(llmClient, toolExecutor, publisher).WithHooks(hooks)

	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		ag := agent.NewAgent("bench-agent", "You are a helpful assistant",
			agent.WithMaxIterations(10),
		)
		task := agent.NewTask("bench-task", "benchmark", "Test input")
		_, _ = taskService.RunTask(ctx, &ag, task)
	}
}

// Benchmark_TaskService_MultiIteration benchmarks tasks requiring multiple iterations.
func Benchmark_TaskService_MultiIteration(b *testing.B) {
	callCount := 0
	llmClient := &mockLLMClient{
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			callCount++
			// Simulate 3 iterations before completing
			if callCount%3 != 0 {
				return agent.NewLLMResponse(
					agent.NewMessage(agent.RoleAssistant, ""),
					"tool_calls",
				).WithToolCalls([]agent.ToolCall{
					agent.NewToolCall(agent.ToolCallID(fmt.Sprintf("tc-%d", callCount)), "calculate", `{"expression": "1+1"}`),
				})
			}
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, "Done"),
				"stop",
			)
		},
	}
	toolExecutor := &mockToolExecutor{result: "2"}
	publisher := &mockEventPublisher{}
	taskService := agent.NewTaskService(llmClient, toolExecutor, publisher)

	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		callCount = 0
		ag := agent.NewAgent("bench-agent", "You are a helpful assistant",
			agent.WithMaxIterations(10),
		)
		task := agent.NewTask("bench-task", "benchmark", "Multi-step task")
		_, _ = taskService.RunTask(ctx, &ag, task)
	}
}

// Benchmark_TaskService_ParallelToolExecution benchmarks parallel tool execution.
func Benchmark_TaskService_ParallelToolExecution(b *testing.B) {
	callCount := 0
	llmClient := &mockLLMClient{
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			callCount++
			if callCount%2 == 1 {
				return agent.NewLLMResponse(
					agent.NewMessage(agent.RoleAssistant, ""),
					"tool_calls",
				).WithToolCalls([]agent.ToolCall{
					agent.NewToolCall("tc-1", "calculate", `{"expression": "1+1"}`),
					agent.NewToolCall("tc-2", "calculate", `{"expression": "2+2"}`),
					agent.NewToolCall("tc-3", "calculate", `{"expression": "3+3"}`),
					agent.NewToolCall("tc-4", "get_current_time", `{}`),
				})
			}
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, "Done"),
				"stop",
			)
		},
	}
	toolExecutor := &mockToolExecutor{result: "result"}
	publisher := &mockEventPublisher{}
	taskService := agent.NewTaskService(llmClient, toolExecutor, publisher).WithParallelToolExecution()

	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		callCount = 0
		ag := agent.NewAgent("bench-agent", "You are a helpful assistant",
			agent.WithMaxIterations(10),
		)
		task := agent.NewTask("bench-task", "benchmark", "Parallel tools")
		_, _ = taskService.RunTask(ctx, &ag, task)
	}
}

// -----------------------------------------------------------------------------
// Real Tool Benchmarks - Actual tool execution
// -----------------------------------------------------------------------------

// Benchmark_RealToolExecutor_Calculate benchmarks actual calculate tool execution.
func Benchmark_RealToolExecutor_Calculate(b *testing.B) {
	toolExecutor := outbound.NewToolExecutor()
	calculateTool := tooling.NewCalculateTool()
	toolExecutor.RegisterTool("calculate", calculateTool.Func)
	toolExecutor.RegisterToolDefinition(calculateTool.Definition)

	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, _ = toolExecutor.Execute(ctx, "calculate", `{"expression": "((10 + 5) * 3 - 15) / 6"}`)
	}
}

// Benchmark_RealToolExecutor_Time benchmarks actual time tool execution.
func Benchmark_RealToolExecutor_Time(b *testing.B) {
	toolExecutor := outbound.NewToolExecutor()
	timeTool := tooling.NewGetCurrentTimeTool()
	toolExecutor.RegisterTool("get_current_time", timeTool.Func)
	toolExecutor.RegisterToolDefinition(timeTool.Definition)

	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, _ = toolExecutor.Execute(ctx, "get_current_time", `{}`)
	}
}

// Benchmark_FullStack_WithRealTools benchmarks the full stack with real tools.
func Benchmark_FullStack_WithRealTools(b *testing.B) {
	callCount := 0
	llmClient := &mockLLMClient{
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			callCount++
			if callCount%2 == 1 {
				return agent.NewLLMResponse(
					agent.NewMessage(agent.RoleAssistant, ""),
					"tool_calls",
				).WithToolCalls([]agent.ToolCall{
					agent.NewToolCall("tc-1", "calculate", `{"expression": "100 / 4 + 25"}`),
				})
			}
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, "The result is 50"),
				"stop",
			)
		},
	}

	toolExecutor := outbound.NewToolExecutor()
	calculateTool := tooling.NewCalculateTool()
	toolExecutor.RegisterTool("calculate", calculateTool.Func)
	toolExecutor.RegisterToolDefinition(calculateTool.Definition)
	timeTool := tooling.NewGetCurrentTimeTool()
	toolExecutor.RegisterTool("get_current_time", timeTool.Func)
	toolExecutor.RegisterToolDefinition(timeTool.Definition)

	publisher := &mockEventPublisher{}
	taskService := agent.NewTaskService(llmClient, toolExecutor, publisher)

	ctx := context.Background()
	input := chatting.SendMessageInput{Message: "What is 100/4 + 25?"}

	b.ResetTimer()
	for b.Loop() {
		callCount = 0
		ag := agent.NewAgent("bench-agent", "You are a helpful assistant",
			agent.WithMaxIterations(10),
			agent.WithMaxMessages(100),
		)
		useCase := chatting.NewSendMessageUseCase(taskService, &ag)
		_, _ = useCase.Execute(ctx, input)
	}
}

// -----------------------------------------------------------------------------
// Message Handling Benchmarks
// -----------------------------------------------------------------------------

// Benchmark_Agent_AddMessage benchmarks message handling without trimming.
func Benchmark_Agent_AddMessage(b *testing.B) {
	ag := agent.NewAgent("bench-agent", "prompt", agent.WithMaxMessages(0))
	msg := agent.NewMessage(agent.RoleUser, "Benchmark message content")

	b.ResetTimer()
	for b.Loop() {
		ag.AddMessage(msg)
	}
}

// Benchmark_Agent_AddMessage_WithTrimming benchmarks message handling with trimming.
func Benchmark_Agent_AddMessage_WithTrimming(b *testing.B) {
	ag := agent.NewAgent("bench-agent", "prompt", agent.WithMaxMessages(10))
	msg := agent.NewMessage(agent.RoleUser, "Benchmark message content")

	b.ResetTimer()
	for b.Loop() {
		ag.AddMessage(msg)
	}
}

// Benchmark_Agent_ConversationFlow benchmarks realistic conversation flow.
func Benchmark_Agent_ConversationFlow(b *testing.B) {
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello, can you help me?"),
		agent.NewMessage(agent.RoleAssistant, "Of course! What do you need?"),
		agent.NewMessage(agent.RoleUser, "What is 2 + 2?"),
		agent.NewMessage(agent.RoleTool, `{"result": 4}`),
		agent.NewMessage(agent.RoleAssistant, "The answer is 4"),
	}

	b.ResetTimer()
	for b.Loop() {
		ag := agent.NewAgent("bench-agent", "You are a helpful assistant",
			agent.WithMaxMessages(50),
		)
		for _, msg := range messages {
			ag.AddMessage(msg)
		}
	}
}

// -----------------------------------------------------------------------------
// Object Creation Benchmarks
// -----------------------------------------------------------------------------

// Benchmark_NewTask benchmarks task creation.
func Benchmark_NewTask(b *testing.B) {
	for b.Loop() {
		_ = agent.NewTask("task-id", "task-name", "task input for benchmarking")
	}
}

// Benchmark_NewMessage benchmarks message creation.
func Benchmark_NewMessage(b *testing.B) {
	for b.Loop() {
		_ = agent.NewMessage(agent.RoleUser, "Message content for benchmarking")
	}
}

// Benchmark_NewAgent benchmarks agent creation with options.
func Benchmark_NewAgent(b *testing.B) {
	for b.Loop() {
		_ = agent.NewAgent("bench-agent", "You are a helpful AI assistant",
			agent.WithMaxIterations(10),
			agent.WithMaxMessages(100),
			agent.WithMetadata(agent.Metadata{
				"created_by": "benchmark",
				"model":      "test-model",
			}),
		)
	}
}

// Benchmark_NewToolDefinition benchmarks tool definition creation.
func Benchmark_NewToolDefinition(b *testing.B) {
	for b.Loop() {
		_ = agent.NewToolDefinition("calculate", "Perform arithmetic calculation").
			WithParameter("expression", "The arithmetic expression to evaluate")
	}
}
