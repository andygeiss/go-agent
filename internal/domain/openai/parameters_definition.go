package openai

// ParametersDefinition defines the parameters schema for a function.
type ParametersDefinition struct {
	Properties           map[string]PropertyDefinition `json:"properties"`
	Type                 string                        `json:"type"`
	Required             []string                      `json:"required,omitempty"`
	AdditionalProperties bool                          `json:"additionalProperties"`
}
