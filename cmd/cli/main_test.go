package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/chatting"
	"github.com/andygeiss/go-agent/internal/domain/memorizing"
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

const unusedIDGenerator = "unused"

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

// -----------------------------------------------------------------------------
// Memory System Benchmarks - memorizing context
// -----------------------------------------------------------------------------

// generateNotes creates n memory notes with varied content for benchmarking.
func generateNotes(ctx context.Context, store agent.MemoryStore, n int) {
	sourceTypes := []agent.SourceType{
		agent.SourceTypeFact,
		agent.SourceTypePlanStep,
		agent.SourceTypePreference,
		agent.SourceTypeSummary,
		agent.SourceTypeToolResult,
		agent.SourceTypeUserMessage,
	}
	tags := [][]string{
		{"config", "important"},
		{"preference", "user"},
		{"task", "result"},
		{"fact", "codebase"},
		{"summary", "session"},
	}
	for i := range n {
		noteID := agent.NoteID(fmt.Sprintf("note-%d", i))
		sourceType := sourceTypes[i%len(sourceTypes)]
		note := agent.NewMemoryNote(noteID, sourceType).
			WithRawContent(fmt.Sprintf("This is the raw content for note number %d with some searchable text like apple, banana, cherry", i)).
			WithSummary(fmt.Sprintf("Summary for note %d about various topics including programming and testing", i)).
			WithContextDescription(fmt.Sprintf("Context: Note created during benchmark iteration %d", i)).
			WithKeywords("benchmark", "test", fmt.Sprintf("keyword-%d", i%100)).
			WithTags(tags[i%len(tags)]...).
			WithImportance((i % 5) + 1).
			WithUserID("bench-user").
			WithSessionID("bench-session")
		_ = store.Write(ctx, note)
	}
}

// Benchmark_MemoryStore_Write_100 benchmarks writing 100 notes.
func Benchmark_MemoryStore_Write_100(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		store := outbound.NewInMemoryMemoryStore()
		for i := range 100 {
			noteID := agent.NoteID(fmt.Sprintf("note-%d", i))
			note := agent.NewMemoryNote(noteID, agent.SourceTypeFact).
				WithRawContent(fmt.Sprintf("Content for note %d", i)).
				WithSummary(fmt.Sprintf("Summary %d", i)).
				WithImportance(3)
			_ = store.Write(ctx, note)
		}
	}
}

// Benchmark_MemoryStore_Write_1000 benchmarks writing 1000 notes.
func Benchmark_MemoryStore_Write_1000(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		store := outbound.NewInMemoryMemoryStore()
		for i := range 1000 {
			noteID := agent.NoteID(fmt.Sprintf("note-%d", i))
			note := agent.NewMemoryNote(noteID, agent.SourceTypeFact).
				WithRawContent(fmt.Sprintf("Content for note %d", i)).
				WithSummary(fmt.Sprintf("Summary %d", i)).
				WithImportance(3)
			_ = store.Write(ctx, note)
		}
	}
}

// Benchmark_MemoryStore_Search_100 benchmarks searching 100 notes.
func Benchmark_MemoryStore_Search_100(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotes(ctx, store, 100)

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.Search(ctx, "programming", 10, nil)
	}
}

// Benchmark_MemoryStore_Search_1000 benchmarks searching 1000 notes.
func Benchmark_MemoryStore_Search_1000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotes(ctx, store, 1000)

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.Search(ctx, "programming", 10, nil)
	}
}

// Benchmark_MemoryStore_Search_10000 benchmarks searching 10000 notes.
func Benchmark_MemoryStore_Search_10000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotes(ctx, store, 10000)

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.Search(ctx, "programming", 10, nil)
	}
}

// Benchmark_MemoryStore_Search_WithFilters_1000 benchmarks filtered search on 1000 notes.
func Benchmark_MemoryStore_Search_WithFilters_1000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotes(ctx, store, 1000)

	opts := &agent.MemorySearchOptions{
		UserID:    "bench-user",
		SessionID: "bench-session",
		Tags:      []string{"preference"},
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.Search(ctx, "programming", 10, opts)
	}
}

// Benchmark_MemoryStore_Search_WithFilters_10000 benchmarks filtered search on 10000 notes.
func Benchmark_MemoryStore_Search_WithFilters_10000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotes(ctx, store, 10000)

	opts := &agent.MemorySearchOptions{
		UserID:    "bench-user",
		SessionID: "bench-session",
		Tags:      []string{"preference"},
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.Search(ctx, "programming", 10, opts)
	}
}

// Benchmark_MemoryStore_Get_1000 benchmarks getting a note from 1000 notes.
func Benchmark_MemoryStore_Get_1000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotes(ctx, store, 1000)

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.Get(ctx, "note-500")
	}
}

// Benchmark_MemoryStore_Get_10000 benchmarks getting a note from 10000 notes.
func Benchmark_MemoryStore_Get_10000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotes(ctx, store, 10000)

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.Get(ctx, "note-5000")
	}
}

// -----------------------------------------------------------------------------
// Memorizing Use Case Benchmarks
// -----------------------------------------------------------------------------

// Benchmark_WriteNoteUseCase benchmarks the WriteNote use case.
func Benchmark_WriteNoteUseCase(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	useCase := memorizing.NewWriteNoteUseCase(store)

	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		noteID := agent.NoteID(fmt.Sprintf("note-%d", i))
		note := agent.NewMemoryNote(noteID, agent.SourceTypePreference).
			WithRawContent("User prefers dark mode").
			WithSummary("Dark mode preference").
			WithImportance(4)
		_ = useCase.Execute(ctx, note)
	}
}

// Benchmark_SearchNotesUseCase_1000 benchmarks the SearchNotes use case with 1000 notes.
func Benchmark_SearchNotesUseCase_1000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotes(ctx, store, 1000)
	useCase := memorizing.NewSearchNotesUseCase(store)

	b.ResetTimer()
	for b.Loop() {
		_, _ = useCase.Execute(ctx, "apple banana", 10, nil)
	}
}

// Benchmark_SearchNotesUseCase_10000 benchmarks the SearchNotes use case with 10000 notes.
func Benchmark_SearchNotesUseCase_10000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotes(ctx, store, 10000)
	useCase := memorizing.NewSearchNotesUseCase(store)

	b.ResetTimer()
	for b.Loop() {
		_, _ = useCase.Execute(ctx, "apple banana", 10, nil)
	}
}

// Benchmark_GetNoteUseCase benchmarks the GetNote use case.
func Benchmark_GetNoteUseCase(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotes(ctx, store, 1000)
	useCase := memorizing.NewGetNoteUseCase(store)

	b.ResetTimer()
	for b.Loop() {
		_, _ = useCase.Execute(ctx, "note-500")
	}
}

// Benchmark_DeleteNoteUseCase benchmarks the DeleteNote use case.
func Benchmark_DeleteNoteUseCase(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		store := outbound.NewInMemoryMemoryStore()
		// Write a note first
		noteID := agent.NoteID(fmt.Sprintf("note-%d", i))
		note := agent.NewMemoryNote(noteID, agent.SourceTypeFact).
			WithRawContent("Temporary note").
			WithSummary("To be deleted")
		_ = store.Write(ctx, note)

		// Delete it
		useCase := memorizing.NewDeleteNoteUseCase(store)
		_ = useCase.Execute(ctx, noteID)
	}
}

// Benchmark_MemorizingService_FullWorkflow benchmarks a complete memory workflow.
func Benchmark_MemorizingService_FullWorkflow(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		store := outbound.NewInMemoryMemoryStore()
		svc := memorizing.NewService(store)

		// Write some notes
		for j := range 10 {
			noteID := agent.NoteID(fmt.Sprintf("note-%d-%d", i, j))
			note := agent.NewMemoryNote(noteID, agent.SourceTypeFact).
				WithRawContent(fmt.Sprintf("Content %d", j)).
				WithSummary(fmt.Sprintf("Summary %d", j)).
				WithKeywords("workflow", "test").
				WithImportance(3)
			_ = svc.WriteNote(ctx, note)
		}

		// Search notes
		_, _ = svc.SearchNotes(ctx, "Content", 5, nil)

		// Get a specific note
		_, _ = svc.GetNote(ctx, agent.NoteID(fmt.Sprintf("note-%d-5", i)))

		// Delete a note
		_ = svc.DeleteNote(ctx, agent.NoteID(fmt.Sprintf("note-%d-0", i)))
	}
}

// -----------------------------------------------------------------------------
// Memory Tool Benchmarks
// -----------------------------------------------------------------------------

// Benchmark_MemoryTools_Write benchmarks the memory_write tool.
func Benchmark_MemoryTools_Write(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	idCounter := 0
	idGen := func() string {
		idCounter++
		return fmt.Sprintf("note-%d", idCounter)
	}
	svc := tooling.NewMemoryToolService(store, idGen)

	b.ResetTimer()
	for b.Loop() {
		args := `{"source_type": "preference", "raw_content": "User prefers Go", "summary": "Go preference", "importance": 4}`
		_, _ = svc.MemoryWrite(ctx, args)
	}
}

// Benchmark_MemoryTools_Search_1000 benchmarks the memory_search tool with 1000 notes.
func Benchmark_MemoryTools_Search_1000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotes(ctx, store, 1000)
	idGen := func() string { return unusedIDGenerator }
	svc := tooling.NewMemoryToolService(store, idGen)

	b.ResetTimer()
	for b.Loop() {
		args := `{"query": "programming", "limit": 10}`
		_, _ = svc.MemorySearch(ctx, args)
	}
}

// Benchmark_MemoryTools_Search_10000 benchmarks the memory_search tool with 10000 notes.
func Benchmark_MemoryTools_Search_10000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotes(ctx, store, 10000)
	idGen := func() string { return unusedIDGenerator }
	svc := tooling.NewMemoryToolService(store, idGen)

	b.ResetTimer()
	for b.Loop() {
		args := `{"query": "programming", "limit": 10}`
		_, _ = svc.MemorySearch(ctx, args)
	}
}

// Benchmark_MemoryTools_Get benchmarks the memory_get tool.
func Benchmark_MemoryTools_Get(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotes(ctx, store, 1000)
	idGen := func() string { return unusedIDGenerator }
	svc := tooling.NewMemoryToolService(store, idGen)

	b.ResetTimer()
	for b.Loop() {
		args := `{"id": "note-500"}`
		_, _ = svc.MemoryGet(ctx, args)
	}
}

// -----------------------------------------------------------------------------
// Full Stack Memory Benchmarks
// -----------------------------------------------------------------------------

// Benchmark_FullStack_WithMemoryTools benchmarks full agent with memory tools.
func Benchmark_FullStack_WithMemoryTools(b *testing.B) {
	ctx := context.Background()

	// Setup memory store with some existing notes
	store := outbound.NewInMemoryMemoryStore()
	generateNotes(ctx, store, 100)

	idCounter := 100
	idGen := func() string {
		idCounter++
		return fmt.Sprintf("note-%d", idCounter)
	}
	memoryToolSvc := tooling.NewMemoryToolService(store, idGen)

	// Create tool executor with memory tools
	toolExecutor := outbound.NewToolExecutor()

	memoryGetTool := tooling.NewMemoryGetTool(memoryToolSvc)
	toolExecutor.RegisterTool(string(memoryGetTool.ID), memoryGetTool.Func)
	toolExecutor.RegisterToolDefinition(memoryGetTool.Definition)

	memorySearchTool := tooling.NewMemorySearchTool(memoryToolSvc)
	toolExecutor.RegisterTool(string(memorySearchTool.ID), memorySearchTool.Func)
	toolExecutor.RegisterToolDefinition(memorySearchTool.Definition)

	memoryWriteTool := tooling.NewMemoryWriteTool(memoryToolSvc)
	toolExecutor.RegisterTool(string(memoryWriteTool.ID), memoryWriteTool.Func)
	toolExecutor.RegisterToolDefinition(memoryWriteTool.Definition)

	// Mock LLM that uses memory tools
	callCount := 0
	llmClient := &mockLLMClient{
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			callCount++
			if callCount%2 == 1 {
				return agent.NewLLMResponse(
					agent.NewMessage(agent.RoleAssistant, ""),
					"tool_calls",
				).WithToolCalls([]agent.ToolCall{
					agent.NewToolCall("tc-1", "memory_search", `{"query": "programming", "limit": 5}`),
				})
			}
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, "I found relevant memories about programming."),
				"stop",
			)
		},
	}

	publisher := &mockEventPublisher{}
	taskService := agent.NewTaskService(llmClient, toolExecutor, publisher)
	input := chatting.SendMessageInput{Message: "What do you remember about programming?"}

	b.ResetTimer()
	for b.Loop() {
		callCount = 0
		ag := agent.NewAgent("bench-agent", "You are a helpful assistant with memory",
			agent.WithMaxIterations(10),
			agent.WithMaxMessages(100),
		)
		useCase := chatting.NewSendMessageUseCase(taskService, &ag)
		_, _ = useCase.Execute(ctx, input)
	}
}

// Benchmark_FullStack_MemoryWriteAndSearch benchmarks write then search pattern.
func Benchmark_FullStack_MemoryWriteAndSearch(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		benchmarkWriteAndSearchIteration(ctx, i)
	}
}

// benchmarkWriteAndSearchIteration runs a single iteration of the write-and-search benchmark.
func benchmarkWriteAndSearchIteration(ctx context.Context, iteration int) {
	store := outbound.NewInMemoryMemoryStore()

	idCounter := 0
	idGen := func() string {
		idCounter++
		return fmt.Sprintf("note-%d-%d", iteration, idCounter)
	}
	memoryToolSvc := tooling.NewMemoryToolService(store, idGen)

	toolExecutor := outbound.NewToolExecutor()
	memorySearchTool := tooling.NewMemorySearchTool(memoryToolSvc)
	toolExecutor.RegisterTool(string(memorySearchTool.ID), memorySearchTool.Func)
	toolExecutor.RegisterToolDefinition(memorySearchTool.Definition)
	memoryWriteTool := tooling.NewMemoryWriteTool(memoryToolSvc)
	toolExecutor.RegisterTool(string(memoryWriteTool.ID), memoryWriteTool.Func)
	toolExecutor.RegisterToolDefinition(memoryWriteTool.Definition)

	callCount := 0
	llmClient := &mockLLMClient{
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			callCount++
			return buildWriteSearchResponse(callCount)
		},
	}

	publisher := &mockEventPublisher{}
	taskService := agent.NewTaskService(llmClient, toolExecutor, publisher)
	ag := agent.NewAgent("bench-agent", "You are a helpful assistant",
		agent.WithMaxIterations(15),
		agent.WithMaxMessages(100),
	)
	useCase := chatting.NewSendMessageUseCase(taskService, &ag)
	_, _ = useCase.Execute(ctx, chatting.SendMessageInput{Message: "Remember these facts and search"})
}

// buildWriteSearchResponse builds the appropriate LLM response for write-search benchmark.
func buildWriteSearchResponse(callCount int) agent.LLMResponse {
	if callCount <= 10 {
		return agent.NewLLMResponse(
			agent.NewMessage(agent.RoleAssistant, ""),
			"tool_calls",
		).WithToolCalls([]agent.ToolCall{
			agent.NewToolCall(
				agent.ToolCallID(fmt.Sprintf("tc-%d", callCount)),
				"memory_write",
				fmt.Sprintf(`{"source_type": "fact", "raw_content": "Fact %d about Go programming", "summary": "Go fact %d", "importance": 3}`, callCount, callCount),
			),
		})
	}
	if callCount == 11 {
		return agent.NewLLMResponse(
			agent.NewMessage(agent.RoleAssistant, ""),
			"tool_calls",
		).WithToolCalls([]agent.ToolCall{
			agent.NewToolCall("tc-search", "memory_search", `{"query": "Go programming", "limit": 5}`),
		})
	}
	return agent.NewLLMResponse(
		agent.NewMessage(agent.RoleAssistant, "I've stored and retrieved the facts."),
		"stop",
	)
}

// -----------------------------------------------------------------------------
// MemoryNote Object Benchmarks
// -----------------------------------------------------------------------------

// Benchmark_NewMemoryNote benchmarks memory note creation.
func Benchmark_NewMemoryNote(b *testing.B) {
	for b.Loop() {
		_ = agent.NewMemoryNote("note-1", agent.SourceTypeFact)
	}
}

// Benchmark_MemoryNote_WithBuilders benchmarks memory note with all builders.
func Benchmark_MemoryNote_WithBuilders(b *testing.B) {
	for b.Loop() {
		_ = agent.NewMemoryNote("note-1", agent.SourceTypePreference).
			WithRawContent("User prefers dark mode for better readability").
			WithSummary("Dark mode preference").
			WithContextDescription("User expressed preference during UI customization").
			WithKeywords("dark", "mode", "theme", "preference", "ui").
			WithTags("preference", "ui", "config").
			WithImportance(4).
			WithUserID("user-123").
			WithSessionID("session-456").
			WithTaskID("task-789")
	}
}

// Benchmark_MemoryNote_SearchableText benchmarks searchable text generation.
func Benchmark_MemoryNote_SearchableText(b *testing.B) {
	note := agent.NewMemoryNote("note-1", agent.SourceTypeFact).
		WithRawContent("This is the raw content of the note").
		WithSummary("This is the summary").
		WithContextDescription("This is the context description")

	b.ResetTimer()
	for b.Loop() {
		_ = note.SearchableText()
	}
}

// Benchmark_MemoryNote_HasTag benchmarks tag checking.
func Benchmark_MemoryNote_HasTag(b *testing.B) {
	note := agent.NewMemoryNote("note-1", agent.SourceTypeFact).
		WithTags("preference", "ui", "config", "important", "user")

	b.ResetTimer()
	for b.Loop() {
		_ = note.HasTag("config")
	}
}

// Benchmark_MemoryNote_HasKeyword benchmarks keyword checking.
func Benchmark_MemoryNote_HasKeyword(b *testing.B) {
	note := agent.NewMemoryNote("note-1", agent.SourceTypeFact).
		WithKeywords("dark", "mode", "theme", "preference", "ui", "customization")

	b.ResetTimer()
	for b.Loop() {
		_ = note.HasKeyword("preference")
	}
}
