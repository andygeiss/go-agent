package openai

// FunctionDefinition defines a function that can be called by the model.
type FunctionDefinition struct {
	Description string               `json:"description"`
	Name        string               `json:"name"`
	Parameters  ParametersDefinition `json:"parameters"`
}
