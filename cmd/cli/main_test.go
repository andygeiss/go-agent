package main

import (
	"context"
	"testing"

	"github.com/andygeiss/go-agent/internal/domain/chatting"
	"github.com/andygeiss/go-agent/pkg/agent"
	"github.com/andygeiss/go-agent/pkg/event"
)

// Benchmarks for Profile-Guided Optimization (PGO).
// Run with: just profile
// This generates cpuprofile.pprof for optimized builds.

// mockLLMClient implements agent.LLMClient for benchmarking.
type mockLLMClient struct{}

func (m *mockLLMClient) Run(_ context.Context, _ []agent.Message, _ []agent.ToolDefinition) (agent.LLMResponse, error) {
	return agent.LLMResponse{
		Message: agent.NewMessage(agent.RoleAssistant, "Mock response for benchmarking"),
	}, nil
}

// mockToolExecutor implements agent.ToolExecutor for benchmarking.
type mockToolExecutor struct{}

func (m *mockToolExecutor) Execute(_ context.Context, _ string, _ string) (string, error) {
	return "mock result", nil
}

func (m *mockToolExecutor) GetAvailableTools() []string {
	return []string{"mock_tool"}
}

func (m *mockToolExecutor) GetToolDefinitions() []agent.ToolDefinition {
	return []agent.ToolDefinition{
		agent.NewToolDefinition("mock_tool", "A mock tool for benchmarking"),
	}
}

func (m *mockToolExecutor) HasTool(name string) bool {
	return name == "mock_tool"
}

func (m *mockToolExecutor) RegisterTool(_ string, _ agent.ToolFunc) {}

func (m *mockToolExecutor) RegisterToolDefinition(_ agent.ToolDefinition) {}

// mockEventPublisher implements agent.EventPublisher for benchmarking.
type mockEventPublisher struct{}

func (m *mockEventPublisher) Publish(_ context.Context, _ event.Event) error {
	return nil
}

// Benchmark_SendMessageUseCase benchmarks the full domain use case execution path.
func Benchmark_SendMessageUseCase_Execute(b *testing.B) {
	// Setup mocks
	llmClient := &mockLLMClient{}
	toolExecutor := &mockToolExecutor{}
	publisher := &mockEventPublisher{}

	// Create task service with mocks
	taskService := agent.NewTaskService(llmClient, toolExecutor, publisher)

	// Create agent
	ag := agent.NewAgent("bench-agent", "You are a benchmark assistant",
		agent.WithMaxIterations(10),
		agent.WithMaxMessages(100),
	)

	// Create use case
	useCase := chatting.NewSendMessageUseCase(taskService, &ag)
	ctx := context.Background()
	input := chatting.SendMessageInput{Message: "Benchmark message"}

	b.ResetTimer()
	for b.Loop() {
		_, _ = useCase.Execute(ctx, input)
	}
}

// Benchmark_TaskService_RunTask benchmarks the core agent loop.
func Benchmark_TaskService_RunTask(b *testing.B) {
	// Setup mocks
	llmClient := &mockLLMClient{}
	toolExecutor := &mockToolExecutor{}
	publisher := &mockEventPublisher{}

	// Create task service with mocks
	taskService := agent.NewTaskService(llmClient, toolExecutor, publisher)

	// Create agent
	ag := agent.NewAgent("bench-agent", "You are a benchmark assistant",
		agent.WithMaxIterations(10),
	)

	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		task := agent.NewTask("bench-task", "benchmark", "Test input")
		_, _ = taskService.RunTask(ctx, &ag, task)
	}
}

// Benchmark_Agent_AddMessage benchmarks message handling.
func Benchmark_Agent_AddMessage(b *testing.B) {
	ag := agent.NewAgent("bench-agent", "prompt", agent.WithMaxMessages(1000))
	msg := agent.NewMessage(agent.RoleUser, "Benchmark message content")

	b.ResetTimer()
	for b.Loop() {
		ag.AddMessage(msg)
	}
}

// Benchmark_Agent_WithMaxMessages benchmarks message trimming.
func Benchmark_Agent_AddMessage_WithTrimming(b *testing.B) {
	ag := agent.NewAgent("bench-agent", "prompt", agent.WithMaxMessages(10))
	msg := agent.NewMessage(agent.RoleUser, "Benchmark message content")

	b.ResetTimer()
	for b.Loop() {
		ag.AddMessage(msg)
	}
}

// Benchmark_NewTask benchmarks task creation.
func Benchmark_NewTask(b *testing.B) {
	for b.Loop() {
		_ = agent.NewTask("task-id", "task-name", "task input")
	}
}

// Benchmark_NewMessage benchmarks message creation.
func Benchmark_NewMessage(b *testing.B) {
	for b.Loop() {
		_ = agent.NewMessage(agent.RoleUser, "Message content for benchmarking")
	}
}
