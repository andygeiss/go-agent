package openai

// ---------------------------------------------------------------------------
// ChatCompletionChoice
// ---------------------------------------------------------------------------

// ChatCompletionChoice represents a single choice in the response.
type ChatCompletionChoice struct {
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
	Index        int     `json:"index"`
}

// ---------------------------------------------------------------------------
// ChatCompletionRequest
// ---------------------------------------------------------------------------

// ChatCompletionRequest represents a request to the chat completions endpoint.
type ChatCompletionRequest struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`
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

// ---------------------------------------------------------------------------
// ChatCompletionResponse
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// ChatCompletionUsage
// ---------------------------------------------------------------------------

// ChatCompletionUsage represents token usage statistics.
type ChatCompletionUsage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ---------------------------------------------------------------------------
// Message
// ---------------------------------------------------------------------------

// Message represents a message in the chat completion request/response.
type Message struct {
	Content    string     `json:"content"`
	Role       string     `json:"role"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// NewMessage creates a new message with the given role and content.
func NewMessage(role, content string) Message {
	return Message{
		Role:    role,
		Content: content,
	}
}

// WithToolCallID sets the tool call ID for tool response messages.
func (m Message) WithToolCallID(id string) Message {
	m.ToolCallID = id
	return m
}

// WithToolCalls sets the tool calls for assistant messages.
func (m Message) WithToolCalls(toolCalls []ToolCall) Message {
	m.ToolCalls = toolCalls
	return m
}
