package tooling_test

import (
	"context"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/tooling"
)

func Test_NewCalculateTool_Should_ReturnToolWithDefinition(t *testing.T) {
	// Arrange & Act
	tool := tooling.NewCalculateTool()

	// Assert
	assert.That(t, "must have name", tool.Definition.Name, "calculate")
	assert.That(t, "must have description", tool.Definition.Description != "", true)
	assert.That(t, "must have func", tool.Func != nil, true)
}

func Test_Calculate_With_Addition_Should_ReturnCorrectResult(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Act
	result, err := tooling.Calculate(ctx, `{"expression": "2 + 2"}`)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "must return correct result", result, "4")
}

func Test_Calculate_With_DivisionByZero_Should_ReturnError(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Act
	_, err := tooling.Calculate(ctx, `{"expression": "10 / 0"}`)

	// Assert
	assert.That(t, "must return error", err != nil, true)
}

func Test_Calculate_With_InvalidJSON_Should_ReturnError(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Act
	_, err := tooling.Calculate(ctx, `invalid json`)

	// Assert
	assert.That(t, "must return error", err != nil, true)
}

func Test_Calculate_With_Multiplication_Should_ReturnCorrectResult(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Act
	result, err := tooling.Calculate(ctx, `{"expression": "3 * 4"}`)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "must return correct result", result, "12")
}

func Test_Calculate_With_Parentheses_Should_ReturnCorrectResult(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Act
	result, err := tooling.Calculate(ctx, `{"expression": "(2 + 3) * 4"}`)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "must return correct result", result, "20")
}
