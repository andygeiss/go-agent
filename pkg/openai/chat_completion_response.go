package openai

// ChatCompletionResponse represents a response from the chat completions endpoint.
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Model   string                 `json:"model"`
	Object  string                 `json:"object"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   ChatCompletionUsage    `json:"usage"`
	Created int64                  `json:"created"`
}

// GetFirstChoice returns the first choice from the response, or nil if empty.
func (r ChatCompletionResponse) GetFirstChoice() *ChatCompletionChoice {
	if len(r.Choices) == 0 {
		return nil
	}
	return &r.Choices[0]
}
