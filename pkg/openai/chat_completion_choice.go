package openai

// ChatCompletionChoice represents a single choice in the response.
type ChatCompletionChoice struct {
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
	Index        int     `json:"index"`
}
