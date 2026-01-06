package entities

import "github.com/andygeiss/go-agent/internal/domain/agent/immutable"

// Message represents a message in a conversation with the LLM.
type Message struct {
	Content    string         `json:"content"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	Role       immutable.Role `json:"role"`
	ToolCalls  []ToolCall     `json:"tool_calls,omitempty"`
}

// NewMessage creates a new message with the given role and content.
func NewMessage(role immutable.Role, content string) Message {
	return Message{
		Role:    role,
		Content: content,
	}
}

// WithToolCalls sets the tool calls for the message.
func (m Message) WithToolCalls(toolCalls []ToolCall) Message {
	m.ToolCalls = toolCalls
	return m
}

// WithToolCallID sets the tool call ID for the message.
func (m Message) WithToolCallID(id string) Message {
	m.ToolCallID = id
	return m
}
