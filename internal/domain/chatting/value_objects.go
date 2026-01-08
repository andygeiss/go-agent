package chatting

// AgentStats contains statistics about the agent.
type AgentStats struct {
	AgentID        string
	Model          string
	CompletedTasks int
	FailedTasks    int
	MaxIterations  int
	MaxMessages    int
	MessageCount   int
	TaskCount      int
}

// SendMessageInput contains the input for sending a message.
type SendMessageInput struct {
	Message string
}

// SendMessageOutput contains the output from sending a message.
type SendMessageOutput struct {
	Duration       string
	Error          string
	Response       string
	IterationCount int
	ToolCallCount  int
	Success        bool
}
