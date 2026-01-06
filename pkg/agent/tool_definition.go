package agent

// ToolDefinition describes a tool that can be used by the LLM.
// It follows the OpenAI function calling schema.
type ToolDefinition struct {
	Parameters  map[string]string // Parameter name to description mapping
	Name        string            // Unique name of the tool
	Description string            // Human-readable description
}

// NewToolDefinition creates a new ToolDefinition with the given name and description.
func NewToolDefinition(name string, description string) ToolDefinition {
	return ToolDefinition{
		Name:        name,
		Description: description,
		Parameters:  make(map[string]string),
	}
}

// WithParameter adds a parameter to the tool definition.
func (td ToolDefinition) WithParameter(name string, description string) ToolDefinition {
	td.Parameters[name] = description
	return td
}
