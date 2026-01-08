package openai

// PropertyDefinition defines a single property in a JSON schema.
type PropertyDefinition struct {
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Enum        []string `json:"enum,omitempty"`
}
