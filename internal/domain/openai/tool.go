package openai

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
