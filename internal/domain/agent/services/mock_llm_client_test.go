package services_test

import (
	"context"

	"github.com/andygeiss/go-agent/internal/domain/agent/aggregates"
	"github.com/andygeiss/go-agent/internal/domain/agent/entities"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
	"github.com/andygeiss/go-agent/pkg/event"
)

// Mock implementations

type mockLLMClient struct {
	err        error
	responseFn func(messages []entities.Message) aggregates.LLMResponse
	response   aggregates.LLMResponse
}

func (m *mockLLMClient) Run(_ context.Context, messages []entities.Message, _ []immutable.ToolDefinition) (aggregates.LLMResponse, error) {
	if m.err != nil {
		return aggregates.LLMResponse{}, m.err
	}
	if m.responseFn != nil {
		return m.responseFn(messages), nil
	}
	return m.response, nil
}

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

func (m *mockToolExecutor) GetToolDefinitions() []immutable.ToolDefinition {
	return []immutable.ToolDefinition{
		immutable.NewToolDefinition("search", "Search for items").WithParameter("query", "The search query"),
		immutable.NewToolDefinition("loop_tool", "A tool that loops"),
	}
}

func (m *mockToolExecutor) HasTool(_ string) bool {
	return true
}

type mockEventPublisher struct {
	events []event.Event
}

func (m *mockEventPublisher) Publish(_ context.Context, e event.Event) error {
	m.events = append(m.events, e)
	return nil
}
