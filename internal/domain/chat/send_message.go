package chat

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/andygeiss/go-agent/pkg/agent"
)

// SendMessageInput contains the input for sending a message.
type SendMessageInput struct {
	Message string
}

// SendMessageOutput contains the output from sending a message.
type SendMessageOutput struct {
	Response       string
	Error          string
	Duration       string
	IterationCount int
	ToolCallCount  int
	Success        bool
}

// SendMessageUseCase handles sending a message to the agent and getting a response.
type SendMessageUseCase struct {
	taskRunner  TaskRunner
	agent       *agent.Agent
	taskCounter atomic.Int64
}

// NewSendMessageUseCase creates a new SendMessageUseCase.
func NewSendMessageUseCase(runner TaskRunner, ag *agent.Agent) *SendMessageUseCase {
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
