package immutable

// Role represents the role of a message in a conversation.
// It follows the OpenAI chat completion API role convention.
type Role string

// Standard conversation roles for LLM chat completions.
const (
	RoleSystem    Role = "system"    // System instructions/prompt
	RoleUser      Role = "user"      // Human input
	RoleAssistant Role = "assistant" // LLM response
	RoleTool      Role = "tool"      // Tool execution result
)
