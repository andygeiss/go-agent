package agent

// Message represents a single message in a conversation.
// It follows the OpenAI chat completion message format.
type Message struct {
	ToolCallID ToolCallID
	Content    string
	Role       Role
	ToolCalls  []ToolCall
}

// NewMessage creates a new Message with the given role and content.
func NewMessage(role Role, content string) Message {
	return Message{
		Role:      role,
		Content:   content,
		ToolCalls: make([]ToolCall, 0),
	}
}

// WithToolCalls attaches tool calls to the message.
func (m Message) WithToolCalls(toolCalls []ToolCall) Message {
	m.ToolCalls = toolCalls
	return m
}

// WithToolCallID sets the tool call ID for tool response messages.
func (m Message) WithToolCallID(id ToolCallID) Message {
	m.ToolCallID = id
	return m
}
