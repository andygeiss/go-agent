package immutable_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
)

func Test_ToolDefinition_NewToolDefinition_With_ValidParams_Should_ReturnDefinition(t *testing.T) {
	// Arrange
	name := "search"
	description := "Search the web"

	// Act
	td := immutable.NewToolDefinition(name, description)

	// Assert
	assert.That(t, "tool definition name must match", td.Name, name)
	assert.That(t, "tool definition description must match", td.Description, description)
	assert.That(t, "tool definition parameters must be empty", len(td.Parameters), 0)
}

func Test_ToolDefinition_WithParameter_With_Params_Should_HaveParameters(t *testing.T) {
	// Arrange
	td := immutable.NewToolDefinition("search", "Search the web")

	// Act
	td = td.WithParameter("query", "The search query")

	// Assert
	assert.That(t, "tool definition must have one parameter", len(td.Parameters), 1)
	assert.That(t, "parameter description must match", td.Parameters["query"], "The search query")
}
