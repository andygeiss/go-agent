package openai

// ChatCompletionChoice represents a single choice in the response.
type ChatCompletionChoice struct {
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
	Index        int     `json:"index"`
}

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

// ChatCompletionUsage represents token usage statistics.
type ChatCompletionUsage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
