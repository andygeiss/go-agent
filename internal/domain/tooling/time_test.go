package tooling_test

import (
	"context"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/tooling"
)

func Test_NewGetCurrentTimeTool_Should_ReturnToolDefinition(t *testing.T) {
	// Act
	tool := tooling.NewGetCurrentTimeTool()

	// Assert
	assert.That(t, "tool name must match", tool.Definition.Name, "get_current_time")
}

func Test_GetCurrentTime_With_ValidFormat_Should_ReturnFormattedTime(t *testing.T) {
	// Arrange
	ctx := context.Background()
	format := "2006-01-02"

	// Act
	result, err := tooling.GetCurrentTime(ctx, format)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "result must not be empty", len(result) > 0, true)
}

func Test_GetCurrentTime_With_EmptyFormat_Should_UseDefaultFormat(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Act
	result, err := tooling.GetCurrentTime(ctx, "")

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "result must not be empty", len(result) > 0, true)
}
