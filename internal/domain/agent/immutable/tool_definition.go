package immutable

// ToolDefinition represents the definition of a tool that can be used by the LLM.
// This is used to inform the LLM about available tools and their capabilities.
type ToolDefinition struct {
	Parameters  map[string]string `json:"parameters"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
}

// NewToolDefinition creates a new tool definition.
func NewToolDefinition(name, description string) ToolDefinition {
	return ToolDefinition{
		Name:        name,
		Description: description,
		Parameters:  make(map[string]string),
	}
}

// WithParameter adds a parameter to the tool definition.
func (td ToolDefinition) WithParameter(name, description string) ToolDefinition {
	td.Parameters[name] = description
	return td
}
