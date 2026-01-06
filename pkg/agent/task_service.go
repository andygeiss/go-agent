package agent

import (
	"context"

	"github.com/andygeiss/go-agent/pkg/agent/events"
)

// TaskService orchestrates the agent loop for task execution.
// It coordinates between the LLM, tools, and event publishing.
type TaskService struct {
	llmClient      LLMClient
	toolExecutor   ToolExecutor
	eventPublisher EventPublisher
}

// NewTaskService creates a new TaskService with the given dependencies.
func NewTaskService(llm LLMClient, executor ToolExecutor, publisher EventPublisher) *TaskService {
	return &TaskService{
		llmClient:      llm,
		toolExecutor:   executor,
		eventPublisher: publisher,
	}
}

// RunTask executes a task using the agent loop pattern.
// It runs iterations until the task completes, fails, or max iterations is reached.
func (s *TaskService) RunTask(ctx context.Context, agent *Agent, task *Task) (Result, error) {
	task.Start()
	agent.ResetIteration()

	// Publish task started event
	_ = s.eventPublisher.Publish(ctx, events.NewEventTaskStarted(string(task.ID), task.Name))

	// Add the task input as a user message
	agent.AddMessage(NewMessage(RoleUser, task.Input))

	for agent.CanContinue() {
		agent.IncrementIteration()

		// Build messages with system prompt
		messages := s.buildMessages(agent)

		// Call the LLM
		response, err := s.llmClient.Run(ctx, messages, s.toolExecutor.GetToolDefinitions())
		if err != nil {
			return s.failTask(ctx, task, err.Error())
		}

		// Add assistant response to history
		agent.AddMessage(response.Message)

		// Check if LLM wants to use tools
		if response.HasToolCalls() {
			s.executeToolCalls(ctx, agent, response.ToolCalls)
			continue
		}

		// Task completed - LLM gave a final answer
		task.Complete(response.Message.Content)

		// Publish task completed event
		_ = s.eventPublisher.Publish(ctx, events.NewEventTaskCompleted(string(task.ID), task.Output))

		return NewResult(task.ID, true, task.Output), nil
	}

	// Max iterations reached
	return s.failTask(ctx, task, "max iterations reached")
}

// buildMessages constructs the message list with system prompt.
func (s *TaskService) buildMessages(agent *Agent) []Message {
	messages := make([]Message, 0, len(agent.Messages)+1)
	messages = append(messages, NewMessage(RoleSystem, agent.SystemPrompt))
	messages = append(messages, agent.Messages...)
	return messages
}

// executeToolCalls runs each tool call and adds results to conversation.
func (s *TaskService) executeToolCalls(ctx context.Context, agent *Agent, toolCalls []ToolCall) {
	for i := range toolCalls {
		tc := &toolCalls[i]
		tc.Execute()

		result, err := s.toolExecutor.Execute(ctx, tc.Name, tc.Arguments)
		if err != nil {
			tc.Fail(err.Error())
		} else {
			tc.Complete(result)
		}

		// Publish tool call executed event
		_ = s.eventPublisher.Publish(ctx, events.NewEventToolCallExecuted(
			string(tc.ID),
			tc.Name,
			tc.Result,
			tc.Error,
		))

		agent.AddMessage(tc.ToMessage())
	}
}

// failTask marks the task as failed and publishes the event.
func (s *TaskService) failTask(ctx context.Context, task *Task, errMsg string) (Result, error) {
	task.Fail(errMsg)

	// Publish task failed event
	_ = s.eventPublisher.Publish(ctx, events.NewEventTaskFailed(string(task.ID), errMsg))

	return NewResult(task.ID, false, "").WithError(errMsg), nil
}
