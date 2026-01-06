package openai

// FunctionDefinition defines a function that can be called by the model.
type FunctionDefinition struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Parameters  ParametersDefinition `json:"parameters"`
}
