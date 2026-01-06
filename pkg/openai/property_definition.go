package openai

// PropertyDefinition defines a single property in a JSON schema.
type PropertyDefinition struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}
