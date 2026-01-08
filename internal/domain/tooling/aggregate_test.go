package tooling_test

import (
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

func Test_NewGetCurrentTimeTool_Should_ReturnToolWithDefinition(t *testing.T) {
	// Arrange & Act
	tool := tooling.NewGetCurrentTimeTool()

	// Assert
	assert.That(t, "must have name", tool.Definition.Name, "get_current_time")
	assert.That(t, "must have description", tool.Definition.Description != "", true)
	assert.That(t, "must have func", tool.Func != nil, true)
}
