package outbound_test

import (
	"context"
	"slices"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
)

func Test_ToolExecutor_GetAvailableTools_Should_ReturnTwoTools(t *testing.T) {
	// Arrange
	executor := outbound.NewToolExecutor()

	// Act
	tools := executor.GetAvailableTools()

	// Assert
	assert.That(t, "must have 2 tools", len(tools), 2)
}

func Test_ToolExecutor_GetAvailableTools_Should_ContainGetCurrentTime(t *testing.T) {
	// Arrange
	executor := outbound.NewToolExecutor()

	// Act
	tools := executor.GetAvailableTools()

	// Assert
	assert.That(t, "must contain get_current_time tool", slices.Contains(tools, "get_current_time"), true)
}

func Test_ToolExecutor_GetAvailableTools_Should_ContainCalculate(t *testing.T) {
	// Arrange
	executor := outbound.NewToolExecutor()

	// Act
	tools := executor.GetAvailableTools()

	// Assert
	assert.That(t, "must contain calculate tool", slices.Contains(tools, "calculate"), true)
}

func Test_ToolExecutor_HasTool_With_GetCurrentTime_Should_ReturnTrue(t *testing.T) {
	// Arrange
	executor := outbound.NewToolExecutor()

	// Act
	hasTool := executor.HasTool("get_current_time")

	// Assert
	assert.That(t, "must have get_current_time tool", hasTool, true)
}

func Test_ToolExecutor_HasTool_With_Calculate_Should_ReturnTrue(t *testing.T) {
	// Arrange
	executor := outbound.NewToolExecutor()

	// Act
	hasTool := executor.HasTool("calculate")

	// Assert
	assert.That(t, "must have calculate tool", hasTool, true)
}

func Test_ToolExecutor_HasTool_With_NonexistentTool_Should_ReturnFalse(t *testing.T) {
	// Arrange
	executor := outbound.NewToolExecutor()

	// Act
	hasTool := executor.HasTool("nonexistent")

	// Assert
	assert.That(t, "must not have nonexistent tool", hasTool, false)
}

func Test_ToolExecutor_GetToolDefinitions_Should_ReturnTwoDefinitions(t *testing.T) {
	// Arrange
	executor := outbound.NewToolExecutor()

	// Act
	definitions := executor.GetToolDefinitions()

	// Assert
	assert.That(t, "must have 2 tool definitions", len(definitions), 2)
}

func Test_ToolExecutor_Execute_With_GetCurrentTime_Should_ReturnNonEmptyResult(t *testing.T) {
	// Arrange
	executor := outbound.NewToolExecutor()
	ctx := context.Background()

	// Act
	result, err := executor.Execute(ctx, "get_current_time", "{}")

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "must return non-empty result", result != "", true)
}

func Test_ToolExecutor_Execute_With_Calculate_Should_ReturnNonEmptyResult(t *testing.T) {
	// Arrange
	executor := outbound.NewToolExecutor()
	ctx := context.Background()

	// Act
	result, err := executor.Execute(ctx, "calculate", `{"expression": "2 + 2"}`)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "must return non-empty result", result != "", true)
}

func Test_ToolExecutor_Execute_With_UnknownTool_Should_ReturnError(t *testing.T) {
	// Arrange
	executor := outbound.NewToolExecutor()
	ctx := context.Background()

	// Act
	_, err := executor.Execute(ctx, "unknown_tool", "{}")

	// Assert
	assert.That(t, "must return error for unknown tool", err != nil, true)
}
