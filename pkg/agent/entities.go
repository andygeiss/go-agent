package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/andygeiss/cloud-native-utils/slices"
)

// LLMResponse represents the response from an LLM.
// It contains the assistant message and any tool calls requested.
type LLMResponse struct {
	FinishReason string     // Why the LLM stopped (e.g., "stop", "tool_calls")
	Message      Message    // The response message from the LLM
	ToolCalls    []ToolCall // Tool calls requested by the LLM
}

// NewLLMResponse creates a new LLMResponse with the given message and finish reason.
func NewLLMResponse(message Message, finishReason string) LLMResponse {
	return LLMResponse{
		FinishReason: finishReason,
		Message:      message,
		ToolCalls:    make([]ToolCall, 0),
	}
}

// HasToolCalls returns true if the response contains tool calls.
func (r LLMResponse) HasToolCalls() bool {
	return len(r.ToolCalls) > 0
}

// WithToolCalls sets the tool calls on the response.
func (r LLMResponse) WithToolCalls(toolCalls []ToolCall) LLMResponse {
	r.ToolCalls = toolCalls
	return r
}

// Message represents a single message in a conversation.
// It follows the OpenAI chat completion message format.
type Message struct {
	Content    string     `json:"content"`
	Role       Role       `json:"role"`
	ToolCallID ToolCallID `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// NewMessage creates a new Message with the given role and content.
func NewMessage(role Role, content string) Message {
	return Message{
		Content:   content,
		Role:      role,
		ToolCalls: make([]ToolCall, 0),
	}
}

// WithToolCallID sets the tool call ID for tool response messages.
func (m Message) WithToolCallID(id ToolCallID) Message {
	m.ToolCallID = id
	return m
}

// WithToolCalls attaches tool calls to the message.
func (m Message) WithToolCalls(toolCalls []ToolCall) Message {
	m.ToolCalls = toolCalls
	return m
}

// ParameterType represents the JSON schema type of a tool parameter.
type ParameterType string

// Supported parameter types for tool definitions.
const (
	ParamTypeArray   ParameterType = "array"
	ParamTypeBoolean ParameterType = "boolean"
	ParamTypeInteger ParameterType = "integer"
	ParamTypeNumber  ParameterType = "number"
	ParamTypeObject  ParameterType = "object"
	ParamTypeString  ParameterType = "string"
)

// ParameterDefinition describes a single parameter for a tool.
type ParameterDefinition struct {
	Default     string
	Description string
	Name        string
	Type        ParameterType
	Enum        []string
	Required    bool
}

// NewParameterDefinition creates a new parameter definition with the given name and type.
func NewParameterDefinition(name string, paramType ParameterType) ParameterDefinition {
	return ParameterDefinition{
		Name: name,
		Type: paramType,
	}
}

// WithDefault sets the default value for the parameter.
func (p ParameterDefinition) WithDefault(value string) ParameterDefinition {
	p.Default = value
	return p
}

// WithDescription sets the description for the parameter.
func (p ParameterDefinition) WithDescription(desc string) ParameterDefinition {
	p.Description = desc
	return p
}

// WithEnum sets the allowed values for the parameter.
func (p ParameterDefinition) WithEnum(values ...string) ParameterDefinition {
	p.Enum = values
	return p
}

// WithRequired marks the parameter as required.
func (p ParameterDefinition) WithRequired() ParameterDefinition {
	p.Required = true
	return p
}

// Task represents a unit of work for the agent to execute.
// It has a defined lifecycle: Pending → Running → Completed/Failed.
type Task struct {
	CompletedAt time.Time
	CreatedAt   time.Time
	StartedAt   time.Time
	Error       string
	Input       string
	Name        string
	Output      string
	ID          TaskID
	Status      TaskStatus
	Iterations  int
}

// NewTask creates a new Task with the given ID, name, and input.
func NewTask(id TaskID, name string, input string) *Task {
	return &Task{
		CreatedAt: time.Now(),
		ID:        id,
		Input:     input,
		Name:      name,
		Status:    TaskStatusPending,
	}
}

// Complete marks the task as successfully completed with the given output.
func (t *Task) Complete(output string) {
	t.CompletedAt = time.Now()
	t.Output = output
	t.Status = TaskStatusCompleted
}

// Duration returns the execution duration of the task.
// Returns 0 if the task hasn't started or hasn't completed.
func (t *Task) Duration() time.Duration {
	if t.StartedAt.IsZero() {
		return 0
	}
	if t.CompletedAt.IsZero() {
		return time.Since(t.StartedAt)
	}
	return t.CompletedAt.Sub(t.StartedAt)
}

// Fail marks the task as failed with the given error message.
func (t *Task) Fail(errMsg string) {
	t.CompletedAt = time.Now()
	t.Error = errMsg
	t.Status = TaskStatusFailed
}

// IncrementIterations increments the iteration counter.
func (t *Task) IncrementIterations() {
	t.Iterations++
}

// IsTerminal returns true if the task is in a terminal state (completed or failed).
func (t *Task) IsTerminal() bool {
	return t.Status == TaskStatusCompleted || t.Status == TaskStatusFailed
}

// Start marks the task as running.
func (t *Task) Start() {
	t.StartedAt = time.Now()
	t.Status = TaskStatusRunning
}

// WaitTime returns how long the task waited before starting.
// Returns 0 if the task hasn't started.
func (t *Task) WaitTime() time.Duration {
	if t.StartedAt.IsZero() {
		return time.Since(t.CreatedAt)
	}
	return t.StartedAt.Sub(t.CreatedAt)
}

// ToolCall represents a tool invocation requested by the LLM.
// It tracks the tool name, arguments, and execution result.
type ToolCall struct {
	Arguments string         `json:"arguments"`        // JSON-encoded arguments
	Error     string         `json:"error,omitempty"`  // Error message if failed
	ID        ToolCallID     `json:"id"`               // Unique identifier for this call
	Name      string         `json:"name"`             // Name of the tool to execute
	Result    string         `json:"result,omitempty"` // Execution result
	Status    ToolCallStatus `json:"status,omitempty"` // Current execution state
}

// NewToolCall creates a new ToolCall with the given ID, name, and arguments.
func NewToolCall(id ToolCallID, name string, arguments string) ToolCall {
	return ToolCall{
		Arguments: arguments,
		ID:        id,
		Name:      name,
		Status:    ToolCallStatusPending,
	}
}

// Complete marks the tool call as successfully completed with the given result.
func (tc *ToolCall) Complete(result string) {
	tc.Result = result
	tc.Status = ToolCallStatusCompleted
}

// Execute marks the tool call as currently executing.
func (tc *ToolCall) Execute() {
	tc.Status = ToolCallStatusExecuting
}

// Fail marks the tool call as failed with the given error message.
func (tc *ToolCall) Fail(errMsg string) {
	tc.Error = errMsg
	tc.Status = ToolCallStatusFailed
}

// ToMessage converts the tool call result to a tool response message.
func (tc *ToolCall) ToMessage() Message {
	content := tc.Result
	if tc.Status == ToolCallStatusFailed {
		content = "Error: " + tc.Error
	}
	return NewMessage(RoleTool, content).WithToolCallID(tc.ID)
}

// ToolDefinition describes a tool that can be used by the LLM.
// It follows the OpenAI function calling schema.
type ToolDefinition struct {
	Description string                // Human-readable description
	Name        string                // Unique name of the tool
	Parameters  []ParameterDefinition // Ordered parameter definitions
}

// NewToolDefinition creates a new ToolDefinition with the given name and description.
func NewToolDefinition(name string, description string) ToolDefinition {
	return ToolDefinition{
		Name:        name,
		Description: description,
		Parameters:  make([]ParameterDefinition, 0),
	}
}

// WithParameter adds a simple string parameter to the tool definition.
// For more control, use WithParameterDef instead.
func (td ToolDefinition) WithParameter(name string, description string) ToolDefinition {
	td.Parameters = append(td.Parameters, ParameterDefinition{
		Name:        name,
		Description: description,
		Type:        ParamTypeString,
		Required:    false,
	})
	return td
}

// WithParameterDef adds a parameter definition to the tool.
func (td ToolDefinition) WithParameterDef(param ParameterDefinition) ToolDefinition {
	td.Parameters = append(td.Parameters, param)
	return td
}

// GetRequiredParameters returns the names of all required parameters.
func (td ToolDefinition) GetRequiredParameters() []string {
	requiredParams := slices.Filter(td.Parameters, func(p ParameterDefinition) bool {
		return p.Required
	})
	return slices.Map(requiredParams, func(p ParameterDefinition) string {
		return p.Name
	})
}

// HasParameter checks if a parameter with the given name exists.
func (td ToolDefinition) HasParameter(name string) bool {
	for _, p := range td.Parameters {
		if p.Name == name {
			return true
		}
	}
	return false
}

// GetParameter returns the parameter definition for the given name.
// Returns an empty definition if not found.
func (td ToolDefinition) GetParameter(name string) ParameterDefinition {
	for _, p := range td.Parameters {
		if p.Name == name {
			return p
		}
	}
	return ParameterDefinition{}
}

// ValidationError represents argument validation failures with field-level details.
type ValidationError struct {
	Errors   map[string]string // Map of field name to error description
	ToolName string            // Name of the tool being validated
}

// NewValidationError creates a new ValidationError for the given tool.
func NewValidationError(toolName string) *ValidationError {
	return &ValidationError{
		ToolName: toolName,
		Errors:   make(map[string]string),
	}
}

// AddError adds a field-level validation error.
func (v *ValidationError) AddError(field, message string) {
	v.Errors[field] = message
}

// HasErrors returns true if there are validation errors.
func (v *ValidationError) HasErrors() bool {
	return len(v.Errors) > 0
}

// Error implements the error interface.
func (v *ValidationError) Error() string {
	if len(v.Errors) == 0 {
		return "validation failed for tool " + v.ToolName
	}

	parts := make([]string, 0, len(v.Errors))
	for field, msg := range v.Errors {
		parts = append(parts, fmt.Sprintf("%s: %s", field, msg))
	}
	return fmt.Sprintf("validation failed for tool %s: %s", v.ToolName, strings.Join(parts, "; "))
}

// DecodeArgs decodes JSON arguments into the destination struct.
// This centralizes argument parsing to avoid repetitive json.Unmarshal patterns.
// Returns ErrInvalidArguments if the JSON is malformed.
func DecodeArgs(args string, dst any) error {
	if err := json.Unmarshal([]byte(args), dst); err != nil {
		return fmt.Errorf("%w: failed to decode arguments: %s", ErrInvalidArguments, err.Error())
	}
	return nil
}

// ValidateArgs validates raw JSON arguments against a tool definition.
// It checks:
//   - Required parameters are present
//   - Enum values are valid (if specified)
//   - Type compatibility (basic JSON type checking)
//
// Returns nil if validation passes, or a *ValidationError with details.
func ValidateArgs(def ToolDefinition, rawArgs string) error {
	// Parse the raw JSON into a map for inspection
	var argsMap map[string]any
	if err := json.Unmarshal([]byte(rawArgs), &argsMap); err != nil {
		return fmt.Errorf("%w: failed to parse arguments: %s", ErrInvalidArguments, err.Error())
	}

	validationErr := NewValidationError(def.Name)
	validateParameters(def.Parameters, argsMap, validationErr)

	if validationErr.HasErrors() {
		return validationErr
	}
	return nil
}

// validateParameters checks all parameters against the provided argument map.
func validateParameters(params []ParameterDefinition, argsMap map[string]any, valErr *ValidationError) {
	for _, param := range params {
		value, exists := argsMap[param.Name]

		// Check required parameters
		if param.Required && !exists {
			valErr.AddError(param.Name, "required parameter missing")
			continue
		}

		// Skip further validation if parameter not provided
		if !exists {
			continue
		}

		// Validate enum values
		if len(param.Enum) > 0 {
			validateEnumValue(param, value, valErr)
			continue
		}

		// Validate type compatibility
		if errMsg := validateType(param.Type, value); errMsg != "" {
			valErr.AddError(param.Name, errMsg)
		}
	}
}

// validateEnumValue checks if the value is a valid enum option.
func validateEnumValue(param ParameterDefinition, value any, valErr *ValidationError) {
	strValue, ok := value.(string)
	if !ok {
		valErr.AddError(param.Name, "expected string for enum parameter")
		return
	}
	if !slices.Contains(param.Enum, strValue) {
		valErr.AddError(param.Name, fmt.Sprintf("value '%s' not in allowed values: %v", strValue, param.Enum))
	}
}

// validateType checks if the provided value matches the expected parameter type.
// Returns an error message if validation fails, empty string if valid.
func validateType(paramType ParameterType, value any) string {
	switch paramType {
	case ParamTypeArray:
		return validateArray(value)
	case ParamTypeBoolean:
		return validateBoolean(value)
	case ParamTypeInteger:
		return validateInteger(value)
	case ParamTypeNumber:
		return validateNumber(value)
	case ParamTypeObject:
		return validateObject(value)
	case ParamTypeString:
		return validateString(value)
	}
	return ""
}

func validateArray(value any) string {
	if _, ok := value.([]any); !ok {
		return "expected array"
	}
	return ""
}

func validateBoolean(value any) string {
	if _, ok := value.(bool); !ok {
		return "expected boolean"
	}
	return ""
}

func validateInteger(value any) string {
	num, ok := value.(float64)
	if !ok || num != float64(int64(num)) {
		return "expected integer"
	}
	return ""
}

func validateNumber(value any) string {
	if _, ok := value.(float64); !ok {
		return "expected number"
	}
	return ""
}

func validateObject(value any) string {
	if _, ok := value.(map[string]any); !ok {
		return "expected object"
	}
	return ""
}

func validateString(value any) string {
	if _, ok := value.(string); !ok {
		return "expected string"
	}
	return ""
}
