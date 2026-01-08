package agent_test

import (
	"errors"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/pkg/agent"
)

// Message tests

func Test_Message_NewMessage_With_ValidParams_Should_ReturnMessage(t *testing.T) {
	// Arrange
	role := agent.RoleUser
	content := "Hello, world!"

	// Act
	msg := agent.NewMessage(role, content)

	// Assert
	assert.That(t, "message role must match", msg.Role, role)
	assert.That(t, "message content must match", msg.Content, content)
}

func Test_Message_WithToolCalls_Should_AttachToolCalls(t *testing.T) {
	// Arrange
	msg := agent.NewMessage(agent.RoleAssistant, "")
	tc := agent.NewToolCall("tc-1", "search", `{"query":"test"}`)

	// Act
	msg = msg.WithToolCalls([]agent.ToolCall{tc})

	// Assert
	assert.That(t, "message must have one tool call", len(msg.ToolCalls), 1)
	assert.That(t, "tool call name must match", msg.ToolCalls[0].Name, "search")
}

func Test_Message_WithToolCallID_Should_SetToolCallID(t *testing.T) {
	// Arrange
	msg := agent.NewMessage(agent.RoleTool, "result")

	// Act
	msg = msg.WithToolCallID("tc-1")

	// Assert
	assert.That(t, "tool call ID must match", msg.ToolCallID, agent.ToolCallID("tc-1"))
}

// Task tests

func Test_Task_NewTask_With_ValidParams_Should_ReturnPendingTask(t *testing.T) {
	// Arrange
	id := agent.TaskID("task-1")
	name := "Test Task"
	input := "test input"

	// Act
	task := agent.NewTask(id, name, input)

	// Assert
	assert.That(t, "task ID must match", task.ID, id)
	assert.That(t, "task name must match", task.Name, name)
	assert.That(t, "task input must match", task.Input, input)
	assert.That(t, "task status must be pending", task.Status, agent.TaskStatusPending)
}

func Test_Task_Complete_With_Output_Should_BeCompleted(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Test", "input")
	task.Start()

	// Act
	task.Complete("task output")

	// Assert
	assert.That(t, "task status must be completed", task.Status, agent.TaskStatusCompleted)
	assert.That(t, "task output must match", task.Output, "task output")
}

func Test_Task_Fail_With_Error_Should_BeFailed(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Test", "input")
	task.Start()

	// Act
	task.Fail("task failed")

	// Assert
	assert.That(t, "task status must be failed", task.Status, agent.TaskStatusFailed)
	assert.That(t, "task error must match", task.Error, "task failed")
}

func Test_Task_IsTerminal_With_PendingTask_Should_ReturnFalse(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Test", "input")

	// Act
	isTerminal := task.IsTerminal()

	// Assert
	assert.That(t, "task must not be terminal", isTerminal, false)
}

func Test_Task_IsTerminal_With_CompletedTask_Should_ReturnTrue(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Test", "input")
	task.Complete("done")

	// Act
	isTerminal := task.IsTerminal()

	// Assert
	assert.That(t, "task must be terminal", isTerminal, true)
}

func Test_Task_Start_With_PendingTask_Should_BeRunning(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Test", "input")

	// Act
	task.Start()

	// Assert
	assert.That(t, "task status must be running", task.Status, agent.TaskStatusRunning)
}

// ToolCall tests

func Test_ToolCall_NewToolCall_With_ValidParams_Should_ReturnPendingToolCall(t *testing.T) {
	// Arrange
	id := agent.ToolCallID("tc-1")
	name := "search"
	args := `{"query": "test"}`

	// Act
	tc := agent.NewToolCall(id, name, args)

	// Assert
	assert.That(t, "tool call ID must match", tc.ID, id)
	assert.That(t, "tool call name must match", tc.Name, name)
	assert.That(t, "tool call arguments must match", tc.Arguments, args)
	assert.That(t, "tool call status must be pending", tc.Status, agent.ToolCallStatusPending)
}

func Test_ToolCall_Complete_With_Result_Should_BeCompleted(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)
	tc.Execute()

	// Act
	tc.Complete("search result")

	// Assert
	assert.That(t, "tool call status must be completed", tc.Status, agent.ToolCallStatusCompleted)
	assert.That(t, "tool call result must match", tc.Result, "search result")
}

func Test_ToolCall_Execute_With_PendingToolCall_Should_BeExecuting(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)

	// Act
	tc.Execute()

	// Assert
	assert.That(t, "tool call status must be executing", tc.Status, agent.ToolCallStatusExecuting)
}

func Test_ToolCall_Fail_With_Error_Should_BeFailed(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)
	tc.Execute()

	// Act
	tc.Fail("tool error")

	// Assert
	assert.That(t, "tool call status must be failed", tc.Status, agent.ToolCallStatusFailed)
	assert.That(t, "tool call error must match", tc.Error, "tool error")
}

func Test_ToolCall_ToMessage_With_CompletedToolCall_Should_ReturnToolMessage(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)
	tc.Complete("result")

	// Act
	msg := tc.ToMessage()

	// Assert
	assert.That(t, "message role must be tool", msg.Role, agent.RoleTool)
	assert.That(t, "message content must be result", msg.Content, "result")
	assert.That(t, "message tool call ID must match", msg.ToolCallID, agent.ToolCallID("tc-1"))
}

func Test_ToolCall_ToMessage_With_FailedToolCall_Should_ReturnErrorMessage(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)
	tc.Fail("tool failed")

	// Act
	msg := tc.ToMessage()

	// Assert
	assert.That(t, "message role must be tool", msg.Role, agent.RoleTool)
	assert.That(t, "message content must contain error", msg.Content, "Error: tool failed")
}

// ToolDefinition tests

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
