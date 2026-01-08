package chat

import (
	"github.com/andygeiss/go-agent/pkg/agent"
)

// ClearConversationUseCase handles clearing the conversation history.
type ClearConversationUseCase struct {
	agent *agent.Agent
}

// NewClearConversationUseCase creates a new ClearConversationUseCase.
func NewClearConversationUseCase(ag *agent.Agent) *ClearConversationUseCase {
	return &ClearConversationUseCase{
		agent: ag,
	}
}

// Execute clears the conversation history.
func (uc *ClearConversationUseCase) Execute() {
	uc.agent.ClearMessages()
}
