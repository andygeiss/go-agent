package openai

// ChatCompletionRequest represents a request to the chat completions endpoint.
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
}

// NewChatCompletionRequest creates a new chat completion request.
func NewChatCompletionRequest(model string, messages []Message) ChatCompletionRequest {
	return ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	}
}

// WithTools adds tools to the request.
func (r ChatCompletionRequest) WithTools(tools []Tool) ChatCompletionRequest {
	r.Tools = tools
	return r
}
