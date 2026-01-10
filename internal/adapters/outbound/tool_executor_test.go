package outbound_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/internal/domain/agent"
)

// mockTool is a simple mock tool function for testing.
func mockTool(_ context.Context, args string) (string, error) {
	if args == `{"fail": true}` {
		return "", errors.New("mock error")
	}
	return "mock_result", nil
}

// newToolExecutorWithMockTools creates a ToolExecutor with mock tools registered.
func newToolExecutorWithMockTools() *outbound.ToolExecutor {
	executor := outbound.NewToolExecutor()

	executor.RegisterTool("mock_tool", mockTool)
	executor.RegisterToolDefinition(agent.NewToolDefinition("mock_tool", "A mock tool for testing"))

	executor.RegisterTool("another_tool", mockTool)
	executor.RegisterToolDefinition(agent.NewToolDefinition("another_tool", "Another mock tool for testing"))

	return executor
}

func Test_ToolExecutor_Execute_With_MockTool_Should_ReturnResult(t *testing.T) {
	// Arrange
	executor := newToolExecutorWithMockTools()
	ctx := context.Background()

	// Act
	result, err := executor.Execute(ctx, "mock_tool", `{}`)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "must return mock result", result, "mock_result")
}

func Test_ToolExecutor_Execute_With_MockTool_Error_Should_ReturnError(t *testing.T) {
	// Arrange
	executor := newToolExecutorWithMockTools()
	ctx := context.Background()

	// Act
	_, err := executor.Execute(ctx, "mock_tool", `{"fail": true}`)

	// Assert
	assert.That(t, "must return error", err != nil, true)
}

func Test_ToolExecutor_Execute_With_UnknownTool_Should_ReturnError(t *testing.T) {
	// Arrange
	executor := newToolExecutorWithMockTools()
	ctx := context.Background()

	// Act
	_, err := executor.Execute(ctx, "unknown_tool", "{}")

	// Assert
	assert.That(t, "must return error for unknown tool", err != nil, true)
}

func Test_ToolExecutor_GetAvailableTools_Should_ContainMockTool(t *testing.T) {
	// Arrange
	executor := newToolExecutorWithMockTools()

	// Act
	tools := executor.GetAvailableTools()

	// Assert
	assert.That(t, "must contain mock_tool", slices.Contains(tools, "mock_tool"), true)
}

func Test_ToolExecutor_GetAvailableTools_Should_ContainAnotherTool(t *testing.T) {
	// Arrange
	executor := newToolExecutorWithMockTools()

	// Act
	tools := executor.GetAvailableTools()

	// Assert
	assert.That(t, "must contain another_tool", slices.Contains(tools, "another_tool"), true)
}

func Test_ToolExecutor_GetAvailableTools_Should_ReturnTwoTools(t *testing.T) {
	// Arrange
	executor := newToolExecutorWithMockTools()

	// Act
	tools := executor.GetAvailableTools()

	// Assert
	assert.That(t, "must have 2 tools", len(tools), 2)
}

func Test_ToolExecutor_GetToolDefinitions_Should_ReturnTwoDefinitions(t *testing.T) {
	// Arrange
	executor := newToolExecutorWithMockTools()

	// Act
	definitions := executor.GetToolDefinitions()

	// Assert
	assert.That(t, "must have 2 tool definitions", len(definitions), 2)
}

func Test_ToolExecutor_HasTool_With_MockTool_Should_ReturnTrue(t *testing.T) {
	// Arrange
	executor := newToolExecutorWithMockTools()

	// Act
	hasTool := executor.HasTool("mock_tool")

	// Assert
	assert.That(t, "must have mock_tool", hasTool, true)
}

func Test_ToolExecutor_HasTool_With_AnotherTool_Should_ReturnTrue(t *testing.T) {
	// Arrange
	executor := newToolExecutorWithMockTools()

	// Act
	hasTool := executor.HasTool("another_tool")

	// Assert
	assert.That(t, "must have another_tool", hasTool, true)
}

func Test_ToolExecutor_HasTool_With_NonexistentTool_Should_ReturnFalse(t *testing.T) {
	// Arrange
	executor := newToolExecutorWithMockTools()

	// Act
	hasTool := executor.HasTool("nonexistent")

	// Assert
	assert.That(t, "must not have nonexistent tool", hasTool, false)
}
