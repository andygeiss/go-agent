package agent_test

import (
	"errors"
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

// DecodeArgs tests

func Test_DecodeArgs_With_ValidJSON_Should_DecodeSuccessfully(t *testing.T) {
	// Arrange
	type args struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	var dst args

	// Act
	err := agent.DecodeArgs(`{"query": "test", "limit": 10}`, &dst)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "query must match", dst.Query, "test")
	assert.That(t, "limit must match", dst.Limit, 10)
}

func Test_DecodeArgs_With_InvalidJSON_Should_ReturnError(t *testing.T) {
	// Arrange
	type args struct {
		Query string `json:"query"`
	}
	var dst args

	// Act
	err := agent.DecodeArgs(`{invalid json}`, &dst)

	// Assert
	assert.That(t, "must return error", err != nil, true)
	assert.That(t, "error must be ErrInvalidArguments", errors.Is(err, agent.ErrInvalidArguments), true)
}

func Test_DecodeArgs_With_EmptyJSON_Should_DecodeSuccessfully(t *testing.T) {
	// Arrange
	type args struct {
		Query string `json:"query"`
	}
	var dst args

	// Act
	err := agent.DecodeArgs(`{}`, &dst)

	// Assert
	assert.That(t, "must not return error", err, nil)
	assert.That(t, "query must be empty string", dst.Query, "")
}

// ValidateArgs tests

func Test_ValidateArgs_With_AllRequiredPresent_Should_ReturnNil(t *testing.T) {
	// Arrange
	def := agent.NewToolDefinition("search", "Search").
		WithParameterDef(agent.NewParameterDefinition("query", agent.ParamTypeString).WithRequired())

	// Act
	err := agent.ValidateArgs(def, `{"query": "test"}`)

	// Assert
	assert.That(t, "must not return error", err, nil)
}

func Test_ValidateArgs_With_MissingRequired_Should_ReturnValidationError(t *testing.T) {
	// Arrange
	def := agent.NewToolDefinition("search", "Search").
		WithParameterDef(agent.NewParameterDefinition("query", agent.ParamTypeString).WithRequired())

	// Act
	err := agent.ValidateArgs(def, `{}`)

	// Assert
	assert.That(t, "must return error", err != nil, true)
	var valErr *agent.ValidationError
	assert.That(t, "error must be ValidationError", errors.As(err, &valErr), true)
	assert.That(t, "tool name must match", valErr.ToolName, "search")
	assert.That(t, "must have query error", valErr.Errors["query"] != "", true)
}

func Test_ValidateArgs_With_ValidEnum_Should_ReturnNil(t *testing.T) {
	// Arrange
	def := agent.NewToolDefinition("sort", "Sort items").
		WithParameterDef(agent.NewParameterDefinition("order", agent.ParamTypeString).
			WithEnum("asc", "desc"))

	// Act
	err := agent.ValidateArgs(def, `{"order": "asc"}`)

	// Assert
	assert.That(t, "must not return error", err, nil)
}

func Test_ValidateArgs_With_InvalidEnum_Should_ReturnValidationError(t *testing.T) {
	// Arrange
	def := agent.NewToolDefinition("sort", "Sort items").
		WithParameterDef(agent.NewParameterDefinition("order", agent.ParamTypeString).
			WithEnum("asc", "desc"))

	// Act
	err := agent.ValidateArgs(def, `{"order": "invalid"}`)

	// Assert
	assert.That(t, "must return error", err != nil, true)
	var valErr *agent.ValidationError
	assert.That(t, "error must be ValidationError", errors.As(err, &valErr), true)
}

func Test_ValidateArgs_With_WrongType_String_Should_ReturnValidationError(t *testing.T) {
	// Arrange
	def := agent.NewToolDefinition("test", "Test").
		WithParameterDef(agent.NewParameterDefinition("name", agent.ParamTypeString))

	// Act
	err := agent.ValidateArgs(def, `{"name": 123}`)

	// Assert
	assert.That(t, "must return error", err != nil, true)
	var valErr *agent.ValidationError
	assert.That(t, "error must be ValidationError", errors.As(err, &valErr), true)
}

func Test_ValidateArgs_With_WrongType_Number_Should_ReturnValidationError(t *testing.T) {
	// Arrange
	def := agent.NewToolDefinition("test", "Test").
		WithParameterDef(agent.NewParameterDefinition("count", agent.ParamTypeNumber))

	// Act
	err := agent.ValidateArgs(def, `{"count": "not a number"}`)

	// Assert
	assert.That(t, "must return error", err != nil, true)
	var valErr *agent.ValidationError
	assert.That(t, "error must be ValidationError", errors.As(err, &valErr), true)
}

func Test_ValidateArgs_With_WrongType_Integer_Should_ReturnValidationError(t *testing.T) {
	// Arrange
	def := agent.NewToolDefinition("test", "Test").
		WithParameterDef(agent.NewParameterDefinition("count", agent.ParamTypeInteger))

	// Act
	err := agent.ValidateArgs(def, `{"count": 3.14}`)

	// Assert
	assert.That(t, "must return error", err != nil, true)
	var valErr *agent.ValidationError
	assert.That(t, "error must be ValidationError", errors.As(err, &valErr), true)
}

func Test_ValidateArgs_With_ValidInteger_Should_ReturnNil(t *testing.T) {
	// Arrange
	def := agent.NewToolDefinition("test", "Test").
		WithParameterDef(agent.NewParameterDefinition("count", agent.ParamTypeInteger))

	// Act
	err := agent.ValidateArgs(def, `{"count": 10}`)

	// Assert
	assert.That(t, "must not return error", err, nil)
}

func Test_ValidateArgs_With_WrongType_Boolean_Should_ReturnValidationError(t *testing.T) {
	// Arrange
	def := agent.NewToolDefinition("test", "Test").
		WithParameterDef(agent.NewParameterDefinition("enabled", agent.ParamTypeBoolean))

	// Act
	err := agent.ValidateArgs(def, `{"enabled": "yes"}`)

	// Assert
	assert.That(t, "must return error", err != nil, true)
	var valErr *agent.ValidationError
	assert.That(t, "error must be ValidationError", errors.As(err, &valErr), true)
}

func Test_ValidateArgs_With_WrongType_Array_Should_ReturnValidationError(t *testing.T) {
	// Arrange
	def := agent.NewToolDefinition("test", "Test").
		WithParameterDef(agent.NewParameterDefinition("items", agent.ParamTypeArray))

	// Act
	err := agent.ValidateArgs(def, `{"items": "not an array"}`)

	// Assert
	assert.That(t, "must return error", err != nil, true)
	var valErr *agent.ValidationError
	assert.That(t, "error must be ValidationError", errors.As(err, &valErr), true)
}

func Test_ValidateArgs_With_WrongType_Object_Should_ReturnValidationError(t *testing.T) {
	// Arrange
	def := agent.NewToolDefinition("test", "Test").
		WithParameterDef(agent.NewParameterDefinition("config", agent.ParamTypeObject))

	// Act
	err := agent.ValidateArgs(def, `{"config": "not an object"}`)

	// Assert
	assert.That(t, "must return error", err != nil, true)
	var valErr *agent.ValidationError
	assert.That(t, "error must be ValidationError", errors.As(err, &valErr), true)
}

func Test_ValidateArgs_With_InvalidJSON_Should_ReturnErrInvalidArguments(t *testing.T) {
	// Arrange
	def := agent.NewToolDefinition("test", "Test")

	// Act
	err := agent.ValidateArgs(def, `{invalid}`)

	// Assert
	assert.That(t, "must return error", err != nil, true)
	assert.That(t, "error must be ErrInvalidArguments", errors.Is(err, agent.ErrInvalidArguments), true)
}

func Test_ValidateArgs_With_OptionalParamMissing_Should_ReturnNil(t *testing.T) {
	// Arrange
	def := agent.NewToolDefinition("search", "Search").
		WithParameterDef(agent.NewParameterDefinition("query", agent.ParamTypeString).WithRequired()).
		WithParameterDef(agent.NewParameterDefinition("limit", agent.ParamTypeInteger))

	// Act
	err := agent.ValidateArgs(def, `{"query": "test"}`)

	// Assert
	assert.That(t, "must not return error", err, nil)
}

// ValidationError tests

func Test_ValidationError_Error_With_NoErrors_Should_ReturnBasicMessage(t *testing.T) {
	// Arrange
	valErr := agent.NewValidationError("test_tool")

	// Act
	msg := valErr.Error()

	// Assert
	assert.That(t, "message must contain tool name", msg, "validation failed for tool test_tool")
}

func Test_ValidationError_Error_With_Errors_Should_IncludeFieldDetails(t *testing.T) {
	// Arrange
	valErr := agent.NewValidationError("test_tool")
	valErr.AddError("field1", "is required")

	// Act
	msg := valErr.Error()

	// Assert
	assert.That(t, "message must contain tool name", msg != "", true)
	assert.That(t, "must have errors", valErr.HasErrors(), true)
}

func Test_ValidationError_HasErrors_With_NoErrors_Should_ReturnFalse(t *testing.T) {
	// Arrange
	valErr := agent.NewValidationError("test_tool")

	// Act & Assert
	assert.That(t, "must not have errors", valErr.HasErrors(), false)
}

func Test_ValidationError_HasErrors_With_Errors_Should_ReturnTrue(t *testing.T) {
	// Arrange
	valErr := agent.NewValidationError("test_tool")
	valErr.AddError("field", "error message")

	// Act & Assert
	assert.That(t, "must have errors", valErr.HasErrors(), true)
}
