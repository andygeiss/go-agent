package openai_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/openai"
)

// ---------------------------------------------------------------------------
// FunctionCall tests
// ---------------------------------------------------------------------------

func Test_FunctionCall_Should_HaveCorrectFields(t *testing.T) {
	// Arrange & Act
	fc := openai.FunctionCall{
		Name:      "get_weather",
		Arguments: `{"location":"London"}`,
	}

	// Assert
	assert.That(t, "name must match", fc.Name, "get_weather")
	assert.That(t, "arguments must match", fc.Arguments, `{"location":"London"}`)
}

// ---------------------------------------------------------------------------
// FunctionDefinition tests
// ---------------------------------------------------------------------------

func Test_FunctionDefinition_Should_HaveCorrectFields(t *testing.T) {
	// Arrange & Act
	fd := openai.FunctionDefinition{
		Name:        "calculate",
		Description: "Calculate a math expression",
		Parameters: openai.ParametersDefinition{
			Type:       "object",
			Properties: make(map[string]openai.PropertyDefinition),
		},
	}

	// Assert
	assert.That(t, "name must match", fd.Name, "calculate")
	assert.That(t, "description must match", fd.Description, "Calculate a math expression")
	assert.That(t, "parameters type must be object", fd.Parameters.Type, "object")
}

// ---------------------------------------------------------------------------
// ParametersDefinition tests
// ---------------------------------------------------------------------------

func Test_ParametersDefinition_Should_HaveCorrectFields(t *testing.T) {
	// Arrange & Act
	pd := openai.ParametersDefinition{
		Type: "object",
		Properties: map[string]openai.PropertyDefinition{
			"name": {Type: "string", Description: "User name"},
		},
		Required:             []string{"name"},
		AdditionalProperties: false,
	}

	// Assert
	assert.That(t, "type must be object", pd.Type, "object")
	assert.That(t, "must have 1 property", len(pd.Properties), 1)
	assert.That(t, "must have 1 required", len(pd.Required), 1)
	assert.That(t, "additional properties must be false", pd.AdditionalProperties, false)
}

// ---------------------------------------------------------------------------
// PropertyDefinition tests
// ---------------------------------------------------------------------------

func Test_PropertyDefinition_Should_HaveCorrectFields(t *testing.T) {
	// Arrange & Act
	prop := openai.PropertyDefinition{
		Type:        "string",
		Description: "The user's name",
	}

	// Assert
	assert.That(t, "type must be string", prop.Type, "string")
	assert.That(t, "description must match", prop.Description, "The user's name")
}

func Test_PropertyDefinition_WithEnum_Should_HaveEnumValues(t *testing.T) {
	// Arrange & Act
	prop := openai.PropertyDefinition{
		Type:        "string",
		Description: "Temperature units",
		Enum:        []string{"celsius", "fahrenheit"},
	}

	// Assert
	assert.That(t, "type must be string", prop.Type, "string")
	assert.That(t, "description must match", prop.Description, "Temperature units")
	assert.That(t, "must have 2 enum values", len(prop.Enum), 2)
	assert.That(t, "first enum must be celsius", prop.Enum[0], "celsius")
	assert.That(t, "second enum must be fahrenheit", prop.Enum[1], "fahrenheit")
}

// ---------------------------------------------------------------------------
// Tool tests
// ---------------------------------------------------------------------------

func Test_NewTool_Should_CreateFunctionTool(t *testing.T) {
	// Arrange & Act
	tool := openai.NewTool("get_weather", "Get the current weather")

	// Assert
	assert.That(t, "type must be function", tool.Type, "function")
	assert.That(t, "name must match", tool.Function.Name, "get_weather")
	assert.That(t, "description must match", tool.Function.Description, "Get the current weather")
	assert.That(t, "parameters type must be object", tool.Function.Parameters.Type, "object")
}

func Test_NewTool_Should_HaveEmptyProperties(t *testing.T) {
	// Arrange & Act
	tool := openai.NewTool("simple", "A simple tool")

	// Assert
	assert.That(t, "properties must be empty", len(tool.Function.Parameters.Properties), 0)
	assert.That(t, "required must be nil", tool.Function.Parameters.Required == nil, true)
}

func Test_Tool_WithParameter_Should_AddRequiredParameter(t *testing.T) {
	// Arrange
	tool := openai.NewTool("get_weather", "Get the current weather")

	// Act
	tool = tool.WithParameter("location", "string", "The city name", true)

	// Assert
	assert.That(t, "must have 1 property", len(tool.Function.Parameters.Properties), 1)
	assert.That(t, "property type must be string", tool.Function.Parameters.Properties["location"].Type, "string")
	assert.That(t, "must have 1 required field", len(tool.Function.Parameters.Required), 1)
	assert.That(t, "required field must match", tool.Function.Parameters.Required[0], "location")
}

func Test_Tool_WithParameter_Should_AddOptionalParameter(t *testing.T) {
	// Arrange
	tool := openai.NewTool("get_weather", "Get the current weather")

	// Act
	tool = tool.WithParameter("units", "string", "Temperature units (celsius/fahrenheit)", false)

	// Assert
	assert.That(t, "must have 1 property", len(tool.Function.Parameters.Properties), 1)
	assert.That(t, "property type must be string", tool.Function.Parameters.Properties["units"].Type, "string")
	assert.That(t, "must have 0 required fields", len(tool.Function.Parameters.Required), 0)
}

func Test_Tool_WithParameter_Should_AddMultipleParameters(t *testing.T) {
	// Arrange
	tool := openai.NewTool("get_weather", "Get the current weather")

	// Act
	tool = tool.
		WithParameter("location", "string", "The city name", true).
		WithParameter("units", "string", "Temperature units", false)

	// Assert
	assert.That(t, "must have 2 properties", len(tool.Function.Parameters.Properties), 2)
	assert.That(t, "must have 1 required field", len(tool.Function.Parameters.Required), 1)
}

func Test_Tool_WithParameter_Should_SetPropertyDescription(t *testing.T) {
	// Arrange
	tool := openai.NewTool("test", "Test tool")

	// Act
	tool = tool.WithParameter("param", "number", "A numeric parameter", false)

	// Assert
	assert.That(t, "description must match", tool.Function.Parameters.Properties["param"].Description, "A numeric parameter")
}

// ---------------------------------------------------------------------------
// ToolCall tests
// ---------------------------------------------------------------------------

func Test_NewToolCall_Should_CreateToolCall(t *testing.T) {
	// Arrange & Act
	tc := openai.NewToolCall("call_abc123", "get_weather", `{"location":"London"}`)

	// Assert
	assert.That(t, "ID must match", tc.ID, "call_abc123")
	assert.That(t, "type must be function", tc.Type, "function")
	assert.That(t, "function name must match", tc.Function.Name, "get_weather")
	assert.That(t, "function arguments must match", tc.Function.Arguments, `{"location":"London"}`)
}

func Test_NewToolCall_Should_HandleEmptyArguments(t *testing.T) {
	// Arrange & Act
	tc := openai.NewToolCall("call_123", "get_time", "{}")

	// Assert
	assert.That(t, "arguments must be empty object", tc.Function.Arguments, "{}")
}

func Test_ToolCall_Should_HaveCorrectFields(t *testing.T) {
	// Arrange & Act
	tc := openai.ToolCall{
		ID:   "call_xyz",
		Type: "function",
		Function: openai.FunctionCall{
			Name:      "calculate",
			Arguments: `{"expression":"1+1"}`,
		},
	}

	// Assert
	assert.That(t, "ID must match", tc.ID, "call_xyz")
	assert.That(t, "type must be function", tc.Type, "function")
	assert.That(t, "function name must match", tc.Function.Name, "calculate")
}
