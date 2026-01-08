package agent_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/pkg/agent"
)

func Test_ToolDefinition_NewToolDefinition_With_ValidParams_Should_ReturnDefinition(t *testing.T) {
	// Arrange
	name := "search"
	description := "Search the web"

	// Act
	td := agent.NewToolDefinition(name, description)

	// Assert
	assert.That(t, "tool definition name must match", td.Name, name)
	assert.That(t, "tool definition description must match", td.Description, description)
	assert.That(t, "tool definition parameters must be empty", len(td.Parameters), 0)
}

func Test_ToolDefinition_WithParameter_With_Params_Should_HaveParameters(t *testing.T) {
	// Arrange
	td := agent.NewToolDefinition("search", "Search the web")

	// Act
	td = td.WithParameter("query", "The search query")

	// Assert
	assert.That(t, "tool definition must have one parameter", len(td.Parameters), 1)
	assert.That(t, "parameter name must match", td.Parameters[0].Name, "query")
	assert.That(t, "parameter description must match", td.Parameters[0].Description, "The search query")
}
