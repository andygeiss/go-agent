package chatting

import (
	"context"
	"fmt"
	"sync/atomic"

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

// SendMessageUseCase handles sending a message to the agent and getting a response.
type SendMessageUseCase struct {
	agent       *agent.Agent
	taskRunner  agent.TaskRunner
	taskCounter atomic.Int64
}

// NewSendMessageUseCase creates a new SendMessageUseCase.
func NewSendMessageUseCase(runner agent.TaskRunner, ag *agent.Agent) *SendMessageUseCase {
	return &SendMessageUseCase{
		taskRunner: runner,
		agent:      ag,
	}
}

// Execute sends a message to the agent and returns the response.
func (uc *SendMessageUseCase) Execute(ctx context.Context, input SendMessageInput) (SendMessageOutput, error) {
	taskNum := uc.taskCounter.Add(1)
	taskID := agent.TaskID(fmt.Sprintf("task-%d", taskNum))
	task := agent.NewTask(taskID, "chat", input.Message)

	result, err := uc.taskRunner.RunTask(ctx, uc.agent, task)
	if err != nil {
		return SendMessageOutput{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return SendMessageOutput{
		Response:       result.Output,
		Success:        result.Success,
		Error:          result.Error,
		Duration:       result.Duration.Round(1000000).String(),
		IterationCount: result.IterationCount,
		ToolCallCount:  result.ToolCallCount,
	}, nil
}
