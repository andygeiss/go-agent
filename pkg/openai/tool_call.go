package openai

// ToolCall represents a tool call in a message.
type ToolCall struct {
	Function FunctionCall `json:"function"`
	ID       string       `json:"id"`
	Type     string       `json:"type"`
}

// NewToolCall creates a new tool call.
func NewToolCall(id, name, arguments string) ToolCall {
	return ToolCall{
		ID:   id,
		Type: "function",
		Function: FunctionCall{
			Name:      name,
			Arguments: arguments,
		},
	}
}
