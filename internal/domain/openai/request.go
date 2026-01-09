package openai

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
