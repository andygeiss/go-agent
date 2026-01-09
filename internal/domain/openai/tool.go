package openai

// FunctionCall represents a function call made by the model.
type FunctionCall struct {
	Arguments string `json:"arguments"`
	Name      string `json:"name"`
}

// FunctionDefinition defines a function that can be called by the model.
type FunctionDefinition struct {
	Description string               `json:"description"`
	Name        string               `json:"name"`
	Parameters  ParametersDefinition `json:"parameters"`
}

// ParametersDefinition defines the parameters schema for a function.
type ParametersDefinition struct {
	Properties           map[string]PropertyDefinition `json:"properties"`
	Type                 string                        `json:"type"`
	Required             []string                      `json:"required,omitempty"`
	AdditionalProperties bool                          `json:"additionalProperties"`
}

// PropertyDefinition defines a single property in a JSON schema.
type PropertyDefinition struct {
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Enum        []string `json:"enum,omitempty"`
}

// Tool defines a tool available to the model.
type Tool struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// NewTool creates a new function tool with the given name and description.
func NewTool(name, description string) Tool {
	return Tool{
		Type: "function",
		Function: FunctionDefinition{
			Name:        name,
			Description: description,
			Parameters: ParametersDefinition{
				Type:       "object",
				Properties: make(map[string]PropertyDefinition),
			},
		},
	}
}

// WithParameter adds a parameter to the tool definition.
func (t Tool) WithParameter(name, paramType, description string, required bool) Tool {
	t.Function.Parameters.Properties[name] = PropertyDefinition{
		Type:        paramType,
		Description: description,
	}
	if required {
		t.Function.Parameters.Required = append(t.Function.Parameters.Required, name)
	}
	return t
}

// ToolCall represents a tool call in a message.
type ToolCall struct {
	Function FunctionCall `json:"function"`
	ID       string       `json:"id"`
	Type     string       `json:"type"`
}

// NewToolCall creates a new tool call.
func NewToolCall(id, name, arguments string) ToolCall {
	return ToolCall{
		ID:   id,
		Type: "function",
		Function: FunctionCall{
			Name:      name,
			Arguments: arguments,
		},
	}
}
