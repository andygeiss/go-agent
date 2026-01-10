package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/event"
	"github.com/andygeiss/go-agent/internal/adapters/inbound"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/chatting"
	"github.com/andygeiss/go-agent/internal/domain/indexing"
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
// - Memory system with text search and embedding-based semantic search
// - Cosine similarity computation at various embedding dimensions (128, 512, 1536)
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
// Embedding Search Benchmarks - Semantic similarity search
// -----------------------------------------------------------------------------

// generateEmbedding creates a deterministic embedding vector for benchmarking.
// Uses a simple pattern based on the seed to create varied but reproducible vectors.
func generateEmbedding(seed int, dim int) agent.Embedding {
	embedding := make(agent.Embedding, dim)
	for i := range dim {
		// Create varied but deterministic values between -1 and 1
		embedding[i] = float32((seed+i)%1000)/500.0 - 1.0
	}
	return embedding
}

// generateNotesWithEmbeddings creates n memory notes with embeddings for benchmarking.
func generateNotesWithEmbeddings(ctx context.Context, store *outbound.MemoryStore, n int, embeddingDim int) {
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
			WithSessionID("bench-session").
			WithEmbedding(generateEmbedding(i, embeddingDim))
		_ = store.Write(ctx, note)
	}
}

// Benchmark_MemoryStore_SearchWithEmbedding_1000_Dim128 benchmarks embedding search with 1000 notes and 128-dim vectors.
func Benchmark_MemoryStore_SearchWithEmbedding_1000_Dim128(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotesWithEmbeddings(ctx, store, 1000, 128)
	queryEmbedding := generateEmbedding(42, 128)

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.SearchWithEmbedding(ctx, "programming", queryEmbedding, 10, nil)
	}
}

// Benchmark_MemoryStore_SearchWithEmbedding_1000_Dim512 benchmarks embedding search with 1000 notes and 512-dim vectors.
func Benchmark_MemoryStore_SearchWithEmbedding_1000_Dim512(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotesWithEmbeddings(ctx, store, 1000, 512)
	queryEmbedding := generateEmbedding(42, 512)

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.SearchWithEmbedding(ctx, "programming", queryEmbedding, 10, nil)
	}
}

// Benchmark_MemoryStore_SearchWithEmbedding_1000_Dim1536 benchmarks embedding search with 1000 notes and 1536-dim vectors (OpenAI ada-002).
func Benchmark_MemoryStore_SearchWithEmbedding_1000_Dim1536(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotesWithEmbeddings(ctx, store, 1000, 1536)
	queryEmbedding := generateEmbedding(42, 1536)

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.SearchWithEmbedding(ctx, "programming", queryEmbedding, 10, nil)
	}
}

// Benchmark_MemoryStore_SearchWithEmbedding_10000_Dim128 benchmarks embedding search with 10000 notes and 128-dim vectors.
func Benchmark_MemoryStore_SearchWithEmbedding_10000_Dim128(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotesWithEmbeddings(ctx, store, 10000, 128)
	queryEmbedding := generateEmbedding(42, 128)

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.SearchWithEmbedding(ctx, "programming", queryEmbedding, 10, nil)
	}
}

// Benchmark_MemoryStore_SearchWithEmbedding_10000_Dim512 benchmarks embedding search with 10000 notes and 512-dim vectors.
func Benchmark_MemoryStore_SearchWithEmbedding_10000_Dim512(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotesWithEmbeddings(ctx, store, 10000, 512)
	queryEmbedding := generateEmbedding(42, 512)

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.SearchWithEmbedding(ctx, "programming", queryEmbedding, 10, nil)
	}
}

// Benchmark_MemoryStore_SearchWithEmbedding_10000_Dim1536 benchmarks embedding search with 10000 notes and 1536-dim vectors (OpenAI ada-002).
func Benchmark_MemoryStore_SearchWithEmbedding_10000_Dim1536(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotesWithEmbeddings(ctx, store, 10000, 1536)
	queryEmbedding := generateEmbedding(42, 1536)

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.SearchWithEmbedding(ctx, "programming", queryEmbedding, 10, nil)
	}
}

// Benchmark_MemoryStore_SearchWithEmbedding_WithFilters_10000 benchmarks filtered embedding search on 10000 notes.
func Benchmark_MemoryStore_SearchWithEmbedding_WithFilters_10000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotesWithEmbeddings(ctx, store, 10000, 512)
	queryEmbedding := generateEmbedding(42, 512)

	opts := &agent.MemorySearchOptions{
		UserID:    "bench-user",
		SessionID: "bench-session",
		Tags:      []string{"preference"},
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.SearchWithEmbedding(ctx, "programming", queryEmbedding, 10, opts)
	}
}

// Benchmark_MemoryStore_SearchWithEmbedding_NilEmbedding_Fallback benchmarks fallback to importance sorting.
func Benchmark_MemoryStore_SearchWithEmbedding_NilEmbedding_Fallback(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateNotesWithEmbeddings(ctx, store, 1000, 512)

	b.ResetTimer()
	for b.Loop() {
		// nil embedding should fallback to importance-based sorting
		_, _ = store.SearchWithEmbedding(ctx, "programming", nil, 10, nil)
	}
}

// Benchmark_MemoryStore_SearchWithEmbedding_MixedEmbeddings benchmarks search where some notes lack embeddings.
func Benchmark_MemoryStore_SearchWithEmbedding_MixedEmbeddings(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()

	// Create a mix of notes with and without embeddings
	for i := range 1000 {
		noteID := agent.NoteID(fmt.Sprintf("note-%d", i))
		note := agent.NewMemoryNote(noteID, agent.SourceTypeFact).
			WithRawContent(fmt.Sprintf("Content %d with programming topics", i)).
			WithSummary(fmt.Sprintf("Summary %d", i)).
			WithImportance((i % 5) + 1)

		// Only half the notes have embeddings
		if i%2 == 0 {
			note = note.WithEmbedding(generateEmbedding(i, 512))
		}
		_ = store.Write(ctx, note)
	}
	queryEmbedding := generateEmbedding(42, 512)

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.SearchWithEmbedding(ctx, "programming", queryEmbedding, 10, nil)
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

// Benchmark_MemoryTools_Search_WithSourceTypes benchmarks memory_search with source_types filter.
func Benchmark_MemoryTools_Search_WithSourceTypes(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateTypedNotes(ctx, store, 1000)
	idGen := func() string { return unusedIDGenerator }
	svc := tooling.NewMemoryToolService(store, idGen)

	b.ResetTimer()
	for b.Loop() {
		args := `{"query": "architecture", "limit": 10, "source_types": ["decision", "fact"]}`
		_, _ = svc.MemorySearch(ctx, args)
	}
}

// Benchmark_MemoryTools_Search_WithMinImportance benchmarks memory_search with min_importance filter.
func Benchmark_MemoryTools_Search_WithMinImportance(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateTypedNotes(ctx, store, 1000)
	idGen := func() string { return unusedIDGenerator }
	svc := tooling.NewMemoryToolService(store, idGen)

	b.ResetTimer()
	for b.Loop() {
		args := `{"query": "programming", "limit": 10, "min_importance": 4}`
		_, _ = svc.MemorySearch(ctx, args)
	}
}

// Benchmark_MemoryTools_Search_WithAllFilters benchmarks memory_search with all filters.
func Benchmark_MemoryTools_Search_WithAllFilters(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateTypedNotes(ctx, store, 1000)
	idGen := func() string { return unusedIDGenerator }
	svc := tooling.NewMemoryToolService(store, idGen)

	b.ResetTimer()
	for b.Loop() {
		args := `{"query": "programming", "limit": 10, "source_types": ["decision", "requirement"], "min_importance": 4, "scope": "session", "tags": ["architecture"]}`
		_, _ = svc.MemorySearch(ctx, args)
	}
}

// Benchmark_MemoryTools_Write_TypedNote benchmarks memory_write with source type.
func Benchmark_MemoryTools_Write_TypedNote(b *testing.B) {
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
		args := `{"source_type": "decision", "raw_content": "Use PostgreSQL for database", "summary": "Database decision", "importance": 4, "tags": ["architecture", "database"]}`
		_, _ = svc.MemoryWrite(ctx, args)
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

// -----------------------------------------------------------------------------
// Typed Memory Note Factory Benchmarks
// -----------------------------------------------------------------------------

// Benchmark_NewDecisionNote benchmarks decision note factory.
func Benchmark_NewDecisionNote(b *testing.B) {
	for b.Loop() {
		_ = agent.NewDecisionNote("note-1", "Use PostgreSQL for persistence", "database", "architecture")
	}
}

// Benchmark_NewExperimentNote benchmarks experiment note factory.
func Benchmark_NewExperimentNote(b *testing.B) {
	for b.Loop() {
		_ = agent.NewExperimentNote("note-1", "Caching reduces latency", "50% reduction observed", "performance")
	}
}

// Benchmark_NewExternalSourceNote benchmarks external source note factory.
func Benchmark_NewExternalSourceNote(b *testing.B) {
	for b.Loop() {
		_ = agent.NewExternalSourceNote("note-1", "https://go.dev/doc", "Official Go documentation", "docs", "reference")
	}
}

// Benchmark_NewFactNote benchmarks fact note factory.
func Benchmark_NewFactNote(b *testing.B) {
	for b.Loop() {
		_ = agent.NewFactNote("note-1", "The system uses hexagonal architecture", "architecture")
	}
}

// Benchmark_NewIssueNote benchmarks issue note factory.
func Benchmark_NewIssueNote(b *testing.B) {
	for b.Loop() {
		_ = agent.NewIssueNote("note-1", "Memory leak in connection pool", "bug", "critical")
	}
}

// Benchmark_NewPlanStepNote benchmarks plan step note factory.
func Benchmark_NewPlanStepNote(b *testing.B) {
	for b.Loop() {
		_ = agent.NewPlanStepNote("note-1", "Implement caching layer", "plan-123", 3, "optimization")
	}
}

// Benchmark_NewPreferenceNote benchmarks preference note factory.
func Benchmark_NewPreferenceNote(b *testing.B) {
	for b.Loop() {
		_ = agent.NewPreferenceNote("note-1", "Use verbose logging in development", "config", "dev")
	}
}

// Benchmark_NewRequirementNote benchmarks requirement note factory.
func Benchmark_NewRequirementNote(b *testing.B) {
	for b.Loop() {
		_ = agent.NewRequirementNote("note-1", "Must support 1000 concurrent users", "scalability", "nfr")
	}
}

// Benchmark_NewRetrospectiveNote benchmarks retrospective note factory.
func Benchmark_NewRetrospectiveNote(b *testing.B) {
	for b.Loop() {
		_ = agent.NewRetrospectiveNote("note-1", "Early optimization was unnecessary", "lessons")
	}
}

// Benchmark_NewSummaryNote benchmarks summary note factory.
func Benchmark_NewSummaryNote(b *testing.B) {
	sourceIDs := []string{"note-1", "note-2", "note-3", "note-4", "note-5"}
	b.ResetTimer()
	for b.Loop() {
		_ = agent.NewSummaryNote("note-100", "Combined summary of database decisions", sourceIDs, "database")
	}
}

// Benchmark_AllFactoryNotes benchmarks creating one note of each factory type.
func Benchmark_AllFactoryNotes(b *testing.B) {
	sourceIDs := []string{"src-1", "src-2"}
	b.ResetTimer()
	for b.Loop() {
		_ = agent.NewDecisionNote("d-1", "decision content", "tag1")
		_ = agent.NewExperimentNote("e-1", "hypothesis", "result", "tag1")
		_ = agent.NewExternalSourceNote("x-1", "https://example.com", "annotation", "tag1")
		_ = agent.NewFactNote("f-1", "fact content", "tag1")
		_ = agent.NewIssueNote("i-1", "issue description", "tag1")
		_ = agent.NewPlanStepNote("p-1", "plan step", "plan-1", 1, "tag1")
		_ = agent.NewPreferenceNote("pr-1", "preference content", "tag1")
		_ = agent.NewRequirementNote("r-1", "requirement content", "tag1")
		_ = agent.NewRetrospectiveNote("rt-1", "retrospective content", "tag1")
		_ = agent.NewSummaryNote("s-1", "summary content", sourceIDs, "tag1")
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

// -----------------------------------------------------------------------------
// SourceType Validation Benchmarks
// -----------------------------------------------------------------------------

// Benchmark_IsValidSourceType benchmarks source type validation.
func Benchmark_IsValidSourceType(b *testing.B) {
	for b.Loop() {
		_ = agent.IsValidSourceType("decision")
		_ = agent.IsValidSourceType("invalid")
		_ = agent.IsValidSourceType("fact")
	}
}

// Benchmark_ParseSourceType benchmarks source type parsing.
func Benchmark_ParseSourceType(b *testing.B) {
	for b.Loop() {
		_ = agent.ParseSourceType("decision")
		_ = agent.ParseSourceType("requirement")
		_ = agent.ParseSourceType("experiment")
	}
}

// Benchmark_ValidSourceTypes benchmarks getting valid source types list.
func Benchmark_ValidSourceTypes(b *testing.B) {
	for b.Loop() {
		_ = agent.ValidSourceTypes()
	}
}

// -----------------------------------------------------------------------------
// MemorySearchOptions Benchmarks
// -----------------------------------------------------------------------------

// Benchmark_MemoryStore_Search_WithSourceTypes_1000 benchmarks search with source type filter.
func Benchmark_MemoryStore_Search_WithSourceTypes_1000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateTypedNotes(ctx, store, 1000)

	opts := &agent.MemorySearchOptions{
		SourceTypes: []agent.SourceType{agent.SourceTypeDecision, agent.SourceTypeFact},
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.Search(ctx, "programming", 10, opts)
	}
}

// Benchmark_MemoryStore_Search_WithMinImportance_1000 benchmarks search with importance filter.
func Benchmark_MemoryStore_Search_WithMinImportance_1000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateTypedNotes(ctx, store, 1000)

	opts := &agent.MemorySearchOptions{
		MinImportance: 4,
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.Search(ctx, "programming", 10, opts)
	}
}

// Benchmark_MemoryStore_Search_WithCombinedFilters_1000 benchmarks search with all filters.
func Benchmark_MemoryStore_Search_WithCombinedFilters_1000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateTypedNotes(ctx, store, 1000)

	opts := &agent.MemorySearchOptions{
		SourceTypes:   []agent.SourceType{agent.SourceTypeDecision, agent.SourceTypeRequirement},
		MinImportance: 3,
		Tags:          []string{"architecture"},
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.Search(ctx, "programming", 10, opts)
	}
}

// Benchmark_MemoryStore_Search_WithCombinedFilters_10000 benchmarks search with all filters at scale.
func Benchmark_MemoryStore_Search_WithCombinedFilters_10000(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateTypedNotes(ctx, store, 10000)

	opts := &agent.MemorySearchOptions{
		SourceTypes:   []agent.SourceType{agent.SourceTypeDecision, agent.SourceTypeRequirement, agent.SourceTypeFact},
		MinImportance: 4,
		Tags:          []string{"architecture", "database"},
		UserID:        "bench-user",
	}

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.Search(ctx, "programming", 10, opts)
	}
}

// generateTypedNotes creates n typed memory notes with varied source types for benchmarking.
func generateTypedNotes(ctx context.Context, store agent.MemoryStore, n int) {
	sourceTypes := []agent.SourceType{
		agent.SourceTypeDecision,
		agent.SourceTypeExperiment,
		agent.SourceTypeExternalSource,
		agent.SourceTypeFact,
		agent.SourceTypeIssue,
		agent.SourceTypePlanStep,
		agent.SourceTypePreference,
		agent.SourceTypeRequirement,
		agent.SourceTypeRetrospective,
		agent.SourceTypeSummary,
	}
	tags := [][]string{
		{"architecture", "database"},
		{"performance", "optimization"},
		{"api", "design"},
		{"testing", "quality"},
		{"deployment", "devops"},
	}
	for i := range n {
		noteID := agent.NoteID(fmt.Sprintf("note-%d", i))
		sourceType := sourceTypes[i%len(sourceTypes)]
		note := agent.NewMemoryNote(noteID, sourceType).
			WithRawContent(fmt.Sprintf("Content for typed note %d about programming and architecture patterns", i)).
			WithSummary(fmt.Sprintf("Typed summary for note %d about programming topics", i)).
			WithContextDescription(fmt.Sprintf("Context: Note created during typed benchmark iteration %d", i)).
			WithKeywords("benchmark", "typed", fmt.Sprintf("keyword-%d", i%100)).
			WithTags(tags[i%len(tags)]...).
			WithImportance((i % 5) + 1).
			WithUserID("bench-user").
			WithSessionID("bench-session")
		_ = store.Write(ctx, note)
	}
}

// -----------------------------------------------------------------------------
// Service-Level Convenience Method Benchmarks
// -----------------------------------------------------------------------------

// Benchmark_MemorizingService_SearchDecisions benchmarks SearchDecisions.
func Benchmark_MemorizingService_SearchDecisions(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateTypedNotes(ctx, store, 1000)
	svc := memorizing.NewService(store)

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.SearchDecisions(ctx, "architecture", 10)
	}
}

// Benchmark_MemorizingService_SearchFacts benchmarks SearchFacts.
func Benchmark_MemorizingService_SearchFacts(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateTypedNotes(ctx, store, 1000)
	svc := memorizing.NewService(store)

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.SearchFacts(ctx, "programming", 10)
	}
}

// Benchmark_MemorizingService_SearchRequirements benchmarks SearchRequirements.
func Benchmark_MemorizingService_SearchRequirements(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateTypedNotes(ctx, store, 1000)
	svc := memorizing.NewService(store)

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.SearchRequirements(ctx, "scalability", 10)
	}
}

// Benchmark_MemorizingService_SearchBySourceTypes benchmarks SearchBySourceTypes with multiple types.
func Benchmark_MemorizingService_SearchBySourceTypes(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryMemoryStore()
	generateTypedNotes(ctx, store, 1000)
	svc := memorizing.NewService(store)
	sourceTypes := []agent.SourceType{agent.SourceTypeDecision, agent.SourceTypeFact, agent.SourceTypeRequirement}

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.SearchBySourceTypes(ctx, "architecture", sourceTypes, 10)
	}
}

// Benchmark_MemorizingService_WriteTypedNote benchmarks WriteTypedNote.
func Benchmark_MemorizingService_WriteTypedNote(b *testing.B) {
	ctx := context.Background()
	opts := &memorizing.TypedNoteOptions{
		Tags:       []string{"architecture", "database"},
		Importance: 4,
		Keywords:   []string{"postgresql", "orm"},
	}

	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		store := outbound.NewInMemoryMemoryStore()
		svc := memorizing.NewService(store)
		noteID := agent.NoteID(fmt.Sprintf("note-%d", i))
		_ = svc.WriteTypedNote(ctx, noteID, agent.SourceTypeDecision, "Use PostgreSQL for persistence", opts)
	}
}

// Benchmark_MemorizingService_TypedWorkflow benchmarks a complete typed memory workflow.
func Benchmark_MemorizingService_TypedWorkflow(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		store := outbound.NewInMemoryMemoryStore()
		svc := memorizing.NewService(store)

		// Write various typed notes
		_ = svc.WriteNote(ctx, agent.NewDecisionNote(agent.NoteID(fmt.Sprintf("d-%d", i)), "Use hexagonal architecture", "architecture"))
		_ = svc.WriteNote(ctx, agent.NewFactNote(agent.NoteID(fmt.Sprintf("f-%d", i)), "Go supports concurrency", "golang"))
		_ = svc.WriteNote(ctx, agent.NewRequirementNote(agent.NoteID(fmt.Sprintf("r-%d", i)), "Must handle 1000 TPS", "performance"))
		_ = svc.WriteNote(ctx, agent.NewPreferenceNote(agent.NoteID(fmt.Sprintf("p-%d", i)), "Prefer explicit error handling", "style"))
		_ = svc.WriteNote(ctx, agent.NewIssueNote(agent.NoteID(fmt.Sprintf("i-%d", i)), "Memory leak detected", "bug"))

		// Search by type
		_, _ = svc.SearchDecisions(ctx, "architecture", 5)
		_, _ = svc.SearchFacts(ctx, "concurrency", 5)
		_, _ = svc.SearchRequirements(ctx, "performance", 5)
	}
}

// =============================================================================
// Indexing Benchmarks - File system indexing and snapshot operations
// =============================================================================

// mockFileWalker implements indexing.FileWalker for benchmarking.
type mockFileWalker struct {
	files []indexing.FileInfo
}

func (m *mockFileWalker) Walk(_ context.Context, _ []string, _ []string) ([]indexing.FileInfo, error) {
	return m.files, nil
}

// generateMockFiles creates a slice of mock FileInfo for benchmarking.
func generateMockFiles(count int) []indexing.FileInfo {
	files := make([]indexing.FileInfo, count)
	now := time.Now()
	for i := range count {
		files[i] = indexing.NewFileInfo(
			fmt.Sprintf("/mock/path/to/file_%d.go", i),
			now.Add(-time.Duration(i)*time.Minute),
			int64(i*100+1024),
		)
	}
	return files
}

// generateMockSnapshot creates a mock snapshot with the given file count.
func generateMockSnapshot(id string, fileCount int) indexing.Snapshot {
	return indexing.NewSnapshot(indexing.SnapshotID(id), generateMockFiles(fileCount))
}

// snapshotIDGen creates a simple ID generator for benchmarks.
func snapshotIDGen() func() string {
	counter := 0
	return func() string {
		counter++
		return fmt.Sprintf("snap-%d", counter)
	}
}

// Benchmark_IndexingService_Scan_100Files benchmarks scanning 100 files.
func Benchmark_IndexingService_Scan_100Files(b *testing.B) {
	ctx := context.Background()
	walker := &mockFileWalker{files: generateMockFiles(100)}
	store := outbound.NewInMemoryIndexStore()
	svc := indexing.NewService(walker, store, snapshotIDGen())

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.Scan(ctx, []string{"/mock"}, nil)
	}
}

// Benchmark_IndexingService_Scan_1000Files benchmarks scanning 1000 files.
func Benchmark_IndexingService_Scan_1000Files(b *testing.B) {
	ctx := context.Background()
	walker := &mockFileWalker{files: generateMockFiles(1000)}
	store := outbound.NewInMemoryIndexStore()
	svc := indexing.NewService(walker, store, snapshotIDGen())

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.Scan(ctx, []string{"/mock"}, nil)
	}
}

// Benchmark_IndexingService_Scan_10000Files benchmarks scanning 10000 files.
func Benchmark_IndexingService_Scan_10000Files(b *testing.B) {
	ctx := context.Background()
	walker := &mockFileWalker{files: generateMockFiles(10000)}
	store := outbound.NewInMemoryIndexStore()
	svc := indexing.NewService(walker, store, snapshotIDGen())

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.Scan(ctx, []string{"/mock"}, nil)
	}
}

// Benchmark_IndexingService_ChangedSince_100Files benchmarks filtering changed files.
func Benchmark_IndexingService_ChangedSince_100Files(b *testing.B) {
	ctx := context.Background()
	walker := &mockFileWalker{files: generateMockFiles(100)}
	store := outbound.NewInMemoryIndexStore()
	svc := indexing.NewService(walker, store, snapshotIDGen())

	// Create a snapshot first
	_, _ = svc.Scan(ctx, []string{"/mock"}, nil)
	since := time.Now().Add(-30 * time.Minute)

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.ChangedSince(ctx, since)
	}
}

// Benchmark_IndexingService_ChangedSince_1000Files benchmarks filtering changed files.
func Benchmark_IndexingService_ChangedSince_1000Files(b *testing.B) {
	ctx := context.Background()
	walker := &mockFileWalker{files: generateMockFiles(1000)}
	store := outbound.NewInMemoryIndexStore()
	svc := indexing.NewService(walker, store, snapshotIDGen())

	// Create a snapshot first
	_, _ = svc.Scan(ctx, []string{"/mock"}, nil)
	since := time.Now().Add(-30 * time.Minute)

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.ChangedSince(ctx, since)
	}
}

// Benchmark_IndexingService_ChangedSince_10000Files benchmarks filtering changed files.
func Benchmark_IndexingService_ChangedSince_10000Files(b *testing.B) {
	ctx := context.Background()
	walker := &mockFileWalker{files: generateMockFiles(10000)}
	store := outbound.NewInMemoryIndexStore()
	svc := indexing.NewService(walker, store, snapshotIDGen())

	// Create a snapshot first
	_, _ = svc.Scan(ctx, []string{"/mock"}, nil)
	since := time.Now().Add(-30 * time.Minute)

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.ChangedSince(ctx, since)
	}
}

// Benchmark_IndexingService_DiffSnapshots_100Files benchmarks diffing snapshots.
func Benchmark_IndexingService_DiffSnapshots_100Files(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryIndexStore()

	// Create two snapshots with some differences
	snapshot1 := generateMockSnapshot("snap-1", 100)
	snapshot2 := generateMockSnapshot("snap-2", 100)
	_ = store.SaveSnapshot(ctx, snapshot1)
	_ = store.SaveSnapshot(ctx, snapshot2)

	walker := &mockFileWalker{}
	svc := indexing.NewService(walker, store, snapshotIDGen())

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.DiffSnapshots(ctx, "snap-1", "snap-2")
	}
}

// Benchmark_IndexingService_DiffSnapshots_1000Files benchmarks diffing snapshots.
func Benchmark_IndexingService_DiffSnapshots_1000Files(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryIndexStore()

	// Create two snapshots with some differences
	snapshot1 := generateMockSnapshot("snap-1", 1000)
	snapshot2 := generateMockSnapshot("snap-2", 1000)
	_ = store.SaveSnapshot(ctx, snapshot1)
	_ = store.SaveSnapshot(ctx, snapshot2)

	walker := &mockFileWalker{}
	svc := indexing.NewService(walker, store, snapshotIDGen())

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.DiffSnapshots(ctx, "snap-1", "snap-2")
	}
}

// Benchmark_IndexingService_DiffSnapshots_10000Files benchmarks diffing snapshots.
func Benchmark_IndexingService_DiffSnapshots_10000Files(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryIndexStore()

	// Create two snapshots with some differences
	snapshot1 := generateMockSnapshot("snap-1", 10000)
	snapshot2 := generateMockSnapshot("snap-2", 10000)
	_ = store.SaveSnapshot(ctx, snapshot1)
	_ = store.SaveSnapshot(ctx, snapshot2)

	walker := &mockFileWalker{}
	svc := indexing.NewService(walker, store, snapshotIDGen())

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.DiffSnapshots(ctx, "snap-1", "snap-2")
	}
}

// Benchmark_IndexStore_SaveSnapshot_100Files benchmarks saving snapshots.
func Benchmark_IndexStore_SaveSnapshot_100Files(b *testing.B) {
	ctx := context.Background()
	snapshot := generateMockSnapshot("snap-1", 100)

	b.ResetTimer()
	for b.Loop() {
		store := outbound.NewInMemoryIndexStore()
		_ = store.SaveSnapshot(ctx, snapshot)
	}
}

// Benchmark_IndexStore_SaveSnapshot_1000Files benchmarks saving snapshots.
func Benchmark_IndexStore_SaveSnapshot_1000Files(b *testing.B) {
	ctx := context.Background()
	snapshot := generateMockSnapshot("snap-1", 1000)

	b.ResetTimer()
	for b.Loop() {
		store := outbound.NewInMemoryIndexStore()
		_ = store.SaveSnapshot(ctx, snapshot)
	}
}

// Benchmark_IndexStore_GetSnapshot benchmarks retrieving snapshots.
func Benchmark_IndexStore_GetSnapshot(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryIndexStore()
	snapshot := generateMockSnapshot("snap-1", 1000)
	_ = store.SaveSnapshot(ctx, snapshot)

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.GetSnapshot(ctx, "snap-1")
	}
}

// Benchmark_IndexStore_GetLatestSnapshot benchmarks retrieving latest snapshot.
func Benchmark_IndexStore_GetLatestSnapshot(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryIndexStore()
	snapshot := generateMockSnapshot("snap-1", 1000)
	_ = store.SaveSnapshot(ctx, snapshot)

	b.ResetTimer()
	for b.Loop() {
		_, _ = store.GetLatestSnapshot(ctx)
	}
}

// Benchmark_FSWalker_Walk_RealFS benchmarks real file system walking.
// This benchmark uses a temp directory with synthetic files.
func Benchmark_FSWalker_Walk_RealFS(b *testing.B) {
	// Create temp directory with test files
	tempDir := b.TempDir()

	// Create 100 test files
	for i := range 100 {
		path := filepath.Join(tempDir, fmt.Sprintf("file_%d.txt", i))
		if err := os.WriteFile(path, []byte("test content"), 0600); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}

	ctx := context.Background()
	walker := inbound.NewFSWalker()

	b.ResetTimer()
	for b.Loop() {
		_, _ = walker.Walk(ctx, []string{tempDir}, nil)
	}
}

// Benchmark_FSWalker_Walk_WithIgnore benchmarks walking with ignore patterns.
func Benchmark_FSWalker_Walk_WithIgnore(b *testing.B) {
	// Create temp directory with test files
	tempDir := b.TempDir()

	// Create mixed test files
	for i := range 100 {
		path := filepath.Join(tempDir, fmt.Sprintf("file_%d.go", i))
		if err := os.WriteFile(path, []byte("test content"), 0600); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
		logPath := filepath.Join(tempDir, fmt.Sprintf("file_%d.log", i))
		if err := os.WriteFile(logPath, []byte("log content"), 0600); err != nil {
			b.Fatalf("Failed to create log file: %v", err)
		}
	}

	ctx := context.Background()
	walker := inbound.NewFSWalker()
	ignore := []string{"*.log"}

	b.ResetTimer()
	for b.Loop() {
		_, _ = walker.Walk(ctx, []string{tempDir}, ignore)
	}
}

// Benchmark_IndexToolService_IndexScan benchmarks the index.scan tool.
func Benchmark_IndexToolService_IndexScan(b *testing.B) {
	ctx := context.Background()
	walker := &mockFileWalker{files: generateMockFiles(100)}
	store := outbound.NewInMemoryIndexStore()
	svc := indexing.NewService(walker, store, snapshotIDGen())
	toolSvc := tooling.NewIndexToolService(svc)

	args := `{"paths": ["/mock"]}`

	b.ResetTimer()
	for b.Loop() {
		_, _ = toolSvc.IndexScan(ctx, args)
	}
}

// Benchmark_IndexToolService_IndexChangedSince benchmarks the index.changed_since tool.
func Benchmark_IndexToolService_IndexChangedSince(b *testing.B) {
	ctx := context.Background()
	walker := &mockFileWalker{files: generateMockFiles(100)}
	store := outbound.NewInMemoryIndexStore()
	svc := indexing.NewService(walker, store, snapshotIDGen())
	toolSvc := tooling.NewIndexToolService(svc)

	// Create initial snapshot
	_, _ = svc.Scan(ctx, []string{"/mock"}, nil)
	args := `{"since": "2024-01-01T00:00:00Z"}`

	b.ResetTimer()
	for b.Loop() {
		_, _ = toolSvc.IndexChangedSince(ctx, args)
	}
}

// Benchmark_IndexToolService_IndexDiffSnapshot benchmarks the index.diff_snapshot tool.
func Benchmark_IndexToolService_IndexDiffSnapshot(b *testing.B) {
	ctx := context.Background()
	store := outbound.NewInMemoryIndexStore()

	// Create two snapshots
	snapshot1 := generateMockSnapshot("snap-1", 100)
	snapshot2 := generateMockSnapshot("snap-2", 100)
	_ = store.SaveSnapshot(ctx, snapshot1)
	_ = store.SaveSnapshot(ctx, snapshot2)

	walker := &mockFileWalker{}
	svc := indexing.NewService(walker, store, snapshotIDGen())
	toolSvc := tooling.NewIndexToolService(svc)

	args := `{"from_id": "snap-1", "to_id": "snap-2"}`

	b.ResetTimer()
	for b.Loop() {
		_, _ = toolSvc.IndexDiffSnapshot(ctx, args)
	}
}

// Benchmark_Snapshot_FileCount benchmarks FileCount method.
func Benchmark_Snapshot_FileCount(b *testing.B) {
	snapshot := generateMockSnapshot("snap-1", 10000)

	b.ResetTimer()
	for b.Loop() {
		_ = snapshot.FileCount()
	}
}

// Benchmark_Snapshot_GetFileByPath benchmarks GetFileByPath method.
func Benchmark_Snapshot_GetFileByPath(b *testing.B) {
	snapshot := generateMockSnapshot("snap-1", 10000)
	targetPath := "/mock/path/to/file_5000.go"

	b.ResetTimer()
	for b.Loop() {
		_ = snapshot.GetFileByPath(targetPath)
	}
}

// Benchmark_NewFileInfo benchmarks FileInfo creation.
func Benchmark_NewFileInfo(b *testing.B) {
	now := time.Now()

	b.ResetTimer()
	for b.Loop() {
		_ = indexing.NewFileInfo("/path/to/file.go", now, 1024)
	}
}

// Benchmark_NewSnapshot benchmarks Snapshot creation.
func Benchmark_NewSnapshot(b *testing.B) {
	files := generateMockFiles(1000)

	b.ResetTimer()
	for b.Loop() {
		_ = indexing.NewSnapshot("snap-1", files)
	}
}

// =============================================================================
// Unit Tests for CLI Helper Functions
// =============================================================================

// Test_parseIndexScanArgs tests the parseIndexScanArgs function.
func Test_parseIndexScanArgs_With_Paths_Should_ReturnPaths(t *testing.T) {
	args := []string{"./src", "./lib"}
	paths, ignore := parseIndexScanArgs(args)

	if len(paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(paths))
	}
	if paths[0] != "./src" || paths[1] != "./lib" {
		t.Errorf("Unexpected paths: %v", paths)
	}
	// Should have default ignore patterns
	if len(ignore) == 0 {
		t.Error("Expected default ignore patterns")
	}
}

// Test_parseIndexScanArgs tests arguments with ignore separator.
func Test_parseIndexScanArgs_With_IgnoreSeparator_Should_SplitCorrectly(t *testing.T) {
	args := []string{"./src", "--", ".git", "*.log"}
	paths, ignore := parseIndexScanArgs(args)

	if len(paths) != 1 || paths[0] != "./src" {
		t.Errorf("Unexpected paths: %v", paths)
	}
	if len(ignore) != 2 || ignore[0] != ".git" || ignore[1] != "*.log" {
		t.Errorf("Unexpected ignore patterns: %v", ignore)
	}
}

// Test_parseIndexScanArgs tests empty arguments.
func Test_parseIndexScanArgs_With_EmptyArgs_Should_UseDefaults(t *testing.T) {
	args := []string{}
	paths, ignore := parseIndexScanArgs(args)

	if len(paths) != 1 {
		t.Errorf("Expected 1 default path, got %d", len(paths))
	}
	if len(ignore) == 0 {
		t.Error("Expected default ignore patterns")
	}
}

// Test_parseSinceTime tests RFC3339 parsing.
func Test_parseSinceTime_With_RFC3339_Should_ParseCorrectly(t *testing.T) {
	args := []string{"2024-01-15T10:00:00Z"}
	result := parseSinceTime(args)

	if result.IsZero() {
		t.Error("Expected non-zero time")
	}
	if result.Year() != 2024 || result.Month() != 1 || result.Day() != 15 {
		t.Errorf("Unexpected date: %v", result)
	}
}

// Test_parseSinceTime tests duration parsing.
func Test_parseSinceTime_With_Duration_Should_ParseCorrectly(t *testing.T) {
	args := []string{"1h"}
	result := parseSinceTime(args)

	if result.IsZero() {
		t.Error("Expected non-zero time")
	}
}

// Test_parseSinceTime tests empty args default.
func Test_parseSinceTime_With_EmptyArgs_Should_DefaultTo24Hours(t *testing.T) {
	args := []string{}
	result := parseSinceTime(args)

	if result.IsZero() {
		t.Error("Expected non-zero time")
	}
}

// Test_parseSinceTime tests invalid format.
func Test_parseSinceTime_With_InvalidFormat_Should_ReturnZero(t *testing.T) {
	// Capture stdout to suppress error message during test
	oldStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldStdout }()

	args := []string{"invalid-time"}
	result := parseSinceTime(args)

	if !result.IsZero() {
		t.Errorf("Expected zero time for invalid format, got %v", result)
	}
}

// Test_truncate tests string truncation.
func Test_truncate_With_ShortString_Should_ReturnUnchanged(t *testing.T) {
	result := truncate("short", 10)
	if result != "short" {
		t.Errorf("Expected 'short', got '%s'", result)
	}
}

// Test_truncate tests string truncation with long string.
func Test_truncate_With_LongString_Should_Truncate(t *testing.T) {
	result := truncate("this is a very long string", 10)
	if len(result) > 10 {
		t.Errorf("Expected max 10 chars, got %d", len(result))
	}
	if result != "this is..." {
		t.Errorf("Expected 'this is...', got '%s'", result)
	}
}

// Test_truncate tests newline removal.
func Test_truncate_With_Newlines_Should_RemoveThem(t *testing.T) {
	result := truncate("line1\nline2", 20)
	if result != "line1 line2" {
		t.Errorf("Expected newlines removed, got '%s'", result)
	}
}

// Test_generateNoteID tests ID generation.
func Test_generateNoteID_Should_ReturnValidFormat(t *testing.T) {
	id := generateNoteID()

	// ID should have correct prefix
	if id[:5] != "note-" {
		t.Error("Expected ID to start with 'note-'")
	}

	// Verify format: note-<nanoseconds>
	if len(id) <= 5 {
		t.Error("Expected ID to have timestamp suffix")
	}
}

// Test_generateSnapshotID tests snapshot ID generation.
func Test_generateSnapshotID_Should_ReturnUniqueIDs(t *testing.T) {
	id1 := generateSnapshotID()

	// IDs should have the correct prefix
	if id1[:5] != "snap-" {
		t.Error("Expected ID to start with 'snap-'")
	}

	// Verify format: snap-<nanoseconds>
	if len(id1) <= 5 {
		t.Error("Expected ID to have timestamp suffix")
	}
}

// Test_parseMemoryFlags_WithNoArgs_Should_DefaultImportanceToZero verifies
// that search operations don't filter by importance by default.
func Test_parseMemoryFlags_WithNoArgs_Should_DefaultImportanceToZero(t *testing.T) {
	flags := parseMemoryFlags([]string{})

	if flags.importance != 0 {
		t.Errorf("Expected default importance 0, got %d", flags.importance)
	}
}

// Test_parseMemoryFlags_WithQuery_Should_PutQueryInRemaining verifies
// that non-flag arguments are collected in remaining.
func Test_parseMemoryFlags_WithQuery_Should_PutQueryInRemaining(t *testing.T) {
	flags := parseMemoryFlags([]string{"banana", "apple"})

	if len(flags.remaining) != 2 {
		t.Fatalf("Expected 2 remaining args, got %d", len(flags.remaining))
	}
	if flags.remaining[0] != "banana" || flags.remaining[1] != "apple" {
		t.Errorf("Expected [banana apple], got %v", flags.remaining)
	}
}

// Test_parseMemoryFlags_WithImportanceFlag_Should_SetImportance verifies
// that --min-importance flag is correctly parsed.
func Test_parseMemoryFlags_WithImportanceFlag_Should_SetImportance(t *testing.T) {
	flags := parseMemoryFlags([]string{"--min-importance", "5", "query"})

	if flags.importance != 5 {
		t.Errorf("Expected importance 5, got %d", flags.importance)
	}
	if len(flags.remaining) != 1 || flags.remaining[0] != "query" {
		t.Errorf("Expected remaining [query], got %v", flags.remaining)
	}
}

// Test_buildSearchOptions_WithNoFilters_Should_ReturnNil verifies that
// search without filters returns nil (no filtering applied).
func Test_buildSearchOptions_WithNoFilters_Should_ReturnNil(t *testing.T) {
	flags := memoryFlags{
		importance: 0,
	}

	opts := buildSearchOptions(flags)

	if opts != nil {
		t.Error("Expected nil options when no filters specified")
	}
}

// Test_buildSearchOptions_WithImportance_Should_ReturnOptions verifies
// that specifying importance creates filter options.
func Test_buildSearchOptions_WithImportance_Should_ReturnOptions(t *testing.T) {
	flags := memoryFlags{
		importance: 3,
	}

	opts := buildSearchOptions(flags)

	if opts == nil {
		t.Fatal("Expected non-nil options")
	}
	if opts.MinImportance != 3 {
		t.Errorf("Expected MinImportance 3, got %d", opts.MinImportance)
	}
}

// Test_buildSearchOptions_WithSourceTypes_Should_ReturnOptions verifies
// that specifying source types creates filter options.
func Test_buildSearchOptions_WithSourceTypes_Should_ReturnOptions(t *testing.T) {
	flags := memoryFlags{
		sourceTypes: []agent.SourceType{agent.SourceTypePreference},
	}

	opts := buildSearchOptions(flags)

	if opts == nil {
		t.Fatal("Expected non-nil options")
	}
	if len(opts.SourceTypes) != 1 || opts.SourceTypes[0] != agent.SourceTypePreference {
		t.Errorf("Expected [preference], got %v", opts.SourceTypes)
	}
}

// Test_buildSearchOptions_WithTags_Should_ReturnOptions verifies
// that specifying tags creates filter options.
func Test_buildSearchOptions_WithTags_Should_ReturnOptions(t *testing.T) {
	flags := memoryFlags{
		tags: []string{"food", "preferences"},
	}

	opts := buildSearchOptions(flags)

	if opts == nil {
		t.Fatal("Expected non-nil options")
	}
	if len(opts.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(opts.Tags))
	}
}
