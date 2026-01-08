package agent

// Message represents a single message in a conversation.
// It follows the OpenAI chat completion message format.
type Message struct {
	Content    string     `json:"content"`
	Role       Role       `json:"role"`
	ToolCallID ToolCallID `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// NewMessage creates a new Message with the given role and content.
func NewMessage(role Role, content string) Message {
	return Message{
		Content:   content,
		Role:      role,
		ToolCalls: make([]ToolCall, 0),
	}
}

// WithToolCallID sets the tool call ID for tool response messages.
func (m Message) WithToolCallID(id ToolCallID) Message {
	m.ToolCallID = id
	return m
}

// WithToolCalls attaches tool calls to the message.
func (m Message) WithToolCalls(toolCalls []ToolCall) Message {
	m.ToolCalls = toolCalls
	return m
}
