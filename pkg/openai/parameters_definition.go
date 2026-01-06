package openai

// ParametersDefinition defines the parameters schema for a function.
type ParametersDefinition struct {
	Type                 string                        `json:"type"`
	Properties           map[string]PropertyDefinition `json:"properties"`
	Required             []string                      `json:"required,omitempty"`
	AdditionalProperties bool                          `json:"additionalProperties"`
}
