package openai

// ChatCompletionUsage represents token usage statistics.
type ChatCompletionUsage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
