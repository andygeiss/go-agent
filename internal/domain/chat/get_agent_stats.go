package chat

import (
	"github.com/andygeiss/go-agent/pkg/agent"
)

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

// GetAgentStatsUseCase handles retrieving agent statistics.
type GetAgentStatsUseCase struct {
	agent *agent.Agent
}

// NewGetAgentStatsUseCase creates a new GetAgentStatsUseCase.
func NewGetAgentStatsUseCase(ag *agent.Agent) *GetAgentStatsUseCase {
	return &GetAgentStatsUseCase{
		agent: ag,
	}
}

// Execute retrieves the agent statistics.
func (uc *GetAgentStatsUseCase) Execute() AgentStats {
	return AgentStats{
		AgentID:        string(uc.agent.ID),
		MessageCount:   uc.agent.MessageCount(),
		TaskCount:      uc.agent.TaskCount(),
		CompletedTasks: uc.agent.CompletedTaskCount(),
		FailedTasks:    uc.agent.FailedTaskCount(),
		MaxIterations:  uc.agent.MaxIterations,
		MaxMessages:    uc.agent.MaxMessages,
		Model:          uc.agent.GetMetadata("model"),
	}
}
