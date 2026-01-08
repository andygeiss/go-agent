package agent

// ParameterType represents the JSON schema type of a tool parameter.
type ParameterType string

// Supported parameter types for tool definitions.
const (
	ParamTypeString  ParameterType = "string"
	ParamTypeNumber  ParameterType = "number"
	ParamTypeInteger ParameterType = "integer"
	ParamTypeBoolean ParameterType = "boolean"
	ParamTypeArray   ParameterType = "array"
	ParamTypeObject  ParameterType = "object"
)

// ParameterDefinition describes a single parameter for a tool.
type ParameterDefinition struct {
	Name        string
	Description string
	Type        ParameterType
	Default     string
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

// WithDescription sets the description for the parameter.
func (p ParameterDefinition) WithDescription(desc string) ParameterDefinition {
	p.Description = desc
	return p
}

// WithRequired marks the parameter as required.
func (p ParameterDefinition) WithRequired() ParameterDefinition {
	p.Required = true
	return p
}

// WithEnum sets the allowed values for the parameter.
func (p ParameterDefinition) WithEnum(values ...string) ParameterDefinition {
	p.Enum = values
	return p
}

// WithDefault sets the default value for the parameter.
func (p ParameterDefinition) WithDefault(value string) ParameterDefinition {
	p.Default = value
	return p
}

// ToolDefinition describes a tool that can be used by the LLM.
// It follows the OpenAI function calling schema.
type ToolDefinition struct {
	Name        string                // Unique name of the tool
	Description string                // Human-readable description
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
	required := make([]string, 0)
	for _, p := range td.Parameters {
		if p.Required {
			required = append(required, p.Name)
		}
	}
	return required
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
