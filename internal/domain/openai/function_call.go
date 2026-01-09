package openai

// FunctionCall represents a function call made by the model.
type FunctionCall struct {
	Arguments string `json:"arguments"`
	Name      string `json:"name"`
}
