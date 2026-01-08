package agent

import (
	"context"
	"time"

	"github.com/andygeiss/go-agent/pkg/agent/events"
)

// TaskService orchestrates the agent loop for task execution.
// It coordinates between the LLM, tools, and event publishing.
type TaskService struct {
	llmClient      LLMClient
	toolExecutor   ToolExecutor
	eventPublisher EventPublisher
	hooks          Hooks
}

// NewTaskService creates a new TaskService with the given dependencies.
func NewTaskService(llm LLMClient, executor ToolExecutor, publisher EventPublisher) *TaskService {
	return &TaskService{
		llmClient:      llm,
		toolExecutor:   executor,
		eventPublisher: publisher,
		hooks:          NewHooks(),
	}
}

// WithHooks sets the hooks for the task service.
func (s *TaskService) WithHooks(hooks Hooks) *TaskService {
	s.hooks = hooks
	return s
}

// taskState holds mutable state during task execution.
type taskState struct {
	startTime     time.Time
	toolCallCount int
}

// RunTask executes a task using the agent loop pattern.
// It runs iterations until the task completes, fails, or max iterations is reached.
func (s *TaskService) RunTask(ctx context.Context, agent *Agent, task *Task) (Result, error) {
	state := &taskState{startTime: time.Now()}

	task.Start()
	agent.ResetIteration()

	if err := s.runBeforeTaskHook(ctx, agent, task); err != nil {
		return s.failTask(ctx, task, err.Error(), state)
	}

	_ = s.eventPublisher.Publish(ctx, events.NewEventTaskStarted(string(task.ID), task.Name))
	agent.AddMessage(NewMessage(RoleUser, task.Input))

	return s.runAgentLoop(ctx, agent, task, state)
}

// runBeforeTaskHook executes the before task hook if configured.
func (s *TaskService) runBeforeTaskHook(ctx context.Context, agent *Agent, task *Task) error {
	if s.hooks.BeforeTask != nil {
		return s.hooks.BeforeTask(ctx, agent, task)
	}
	return nil
}

// runAgentLoop executes the main agent loop until completion or failure.
func (s *TaskService) runAgentLoop(ctx context.Context, agent *Agent, task *Task, state *taskState) (Result, error) {
	for agent.CanContinue() {
		if ctx.Err() != nil {
			return s.failTask(ctx, task, ErrContextCanceled.Error(), state)
		}

		agent.IncrementIteration()
		task.IncrementIterations()

		response, err := s.executeIteration(ctx, agent, task)
		if err != nil {
			return s.failTask(ctx, task, err.Error(), state)
		}

		agent.AddMessage(response.Message)

		if response.HasToolCalls() {
			state.toolCallCount += s.executeToolCalls(ctx, agent, response.ToolCalls)
			continue
		}

		return s.completeTask(ctx, agent, task, response.Message.Content, state)
	}

	return s.failTask(ctx, task, ErrMaxIterationsReached.Error(), state)
}

// executeIteration runs a single iteration of the agent loop.
func (s *TaskService) executeIteration(ctx context.Context, agent *Agent, task *Task) (LLMResponse, error) {
	if s.hooks.BeforeLLMCall != nil {
		if err := s.hooks.BeforeLLMCall(ctx, agent, task); err != nil {
			return LLMResponse{}, err
		}
	}

	messages := s.buildMessages(agent)
	response, err := s.llmClient.Run(ctx, messages, s.toolExecutor.GetToolDefinitions())
	if err != nil {
		return LLMResponse{}, err
	}

	if s.hooks.AfterLLMCall != nil {
		if err := s.hooks.AfterLLMCall(ctx, agent, task); err != nil {
			return LLMResponse{}, err
		}
	}

	return response, nil
}

// completeTask marks the task as completed and publishes the event.
func (s *TaskService) completeTask(ctx context.Context, agent *Agent, task *Task, output string, state *taskState) (Result, error) {
	task.Complete(output)

	if s.hooks.AfterTask != nil {
		_ = s.hooks.AfterTask(ctx, agent, task)
	}

	_ = s.eventPublisher.Publish(ctx, events.NewEventTaskCompleted(string(task.ID), task.Output))

	return NewResult(task.ID, true, task.Output).
		WithDuration(time.Since(state.startTime)).
		WithIterationCount(agent.Iteration).
		WithToolCallCount(state.toolCallCount), nil
}

// buildMessages constructs the message list with system prompt.
func (s *TaskService) buildMessages(agent *Agent) []Message {
	messages := make([]Message, 0, len(agent.Messages)+1)
	messages = append(messages, NewMessage(RoleSystem, agent.SystemPrompt))
	messages = append(messages, agent.Messages...)
	return messages
}

// executeToolCalls runs each tool call and adds results to conversation.
// Returns the number of tool calls executed.
func (s *TaskService) executeToolCalls(ctx context.Context, agent *Agent, toolCalls []ToolCall) int {
	count := 0
	for i := range toolCalls {
		tc := &toolCalls[i]

		// Run before tool call hook
		if s.hooks.BeforeToolCall != nil {
			if err := s.hooks.BeforeToolCall(ctx, agent, tc); err != nil {
				tc.Fail(err.Error())
				agent.AddMessage(tc.ToMessage())
				count++
				continue
			}
		}

		tc.Execute()

		result, err := s.toolExecutor.Execute(ctx, tc.Name, tc.Arguments)
		if err != nil {
			tc.Fail(err.Error())
		} else {
			tc.Complete(result)
		}

		// Run after tool call hook
		if s.hooks.AfterToolCall != nil {
			_ = s.hooks.AfterToolCall(ctx, agent, tc)
		}

		// Publish tool call executed event
		_ = s.eventPublisher.Publish(ctx, events.NewEventToolCallExecuted(
			string(tc.ID),
			tc.Name,
			tc.Result,
			tc.Error,
		))

		agent.AddMessage(tc.ToMessage())
		count++
	}
	return count
}

// failTask marks the task as failed and publishes the event.
func (s *TaskService) failTask(
	ctx context.Context,
	task *Task,
	errMsg string,
	state *taskState,
) (Result, error) {
	task.Fail(errMsg)

	// Run after task hook even on failure
	if s.hooks.AfterTask != nil {
		_ = s.hooks.AfterTask(ctx, nil, task)
	}

	// Publish task failed event
	_ = s.eventPublisher.Publish(ctx, events.NewEventTaskFailed(string(task.ID), errMsg))

	return NewResult(task.ID, false, "").
		WithError(errMsg).
		WithDuration(time.Since(state.startTime)).
		WithIterationCount(task.Iterations).
		WithToolCallCount(state.toolCallCount), nil
}
