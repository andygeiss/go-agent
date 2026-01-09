package agent

import (
	"context"
	"sync"
	"time"

	"github.com/andygeiss/cloud-native-utils/efficiency"
	"github.com/andygeiss/cloud-native-utils/service"
)

// TaskService orchestrates the agent loop for task execution.
// It coordinates between the LLM, tools, and event publishing.
type TaskService struct {
	eventPublisher EventPublisher
	llmClient      LLMClient
	toolExecutor   ToolExecutor
	hooks          Hooks
	parallelTools  bool
}

// NewTaskService creates a new TaskService with the given dependencies.
func NewTaskService(llm LLMClient, executor ToolExecutor, publisher EventPublisher) *TaskService {
	return &TaskService{
		eventPublisher: publisher,
		llmClient:      llm,
		toolExecutor:   executor,
		hooks:          NewHooks(),
	}
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

	_ = s.eventPublisher.Publish(ctx, NewEventTaskStarted(string(task.ID), task.Name))
	agent.AddMessage(NewMessage(RoleUser, task.Input))

	return s.runAgentLoop(ctx, agent, task, state)
}

// WithHooks sets the hooks for the task service.
func (s *TaskService) WithHooks(hooks Hooks) *TaskService {
	s.hooks = hooks
	return s
}

// WithParallelToolExecution enables parallel execution of tool calls.
// When enabled, multiple tool calls from a single LLM response are
// executed concurrently using a worker pool. This can significantly
// improve performance when tools are I/O bound (e.g., API calls).
// Default is sequential execution.
func (s *TaskService) WithParallelToolExecution() *TaskService {
	s.parallelTools = true
	return s
}

// taskState holds mutable state during task execution.
type taskState struct {
	startTime     time.Time
	toolCallCount int
}

// toolCallInput bundles the data needed for parallel tool execution.
type toolCallInput struct {
	tc    *ToolCall
	index int
}

// toolCallOutput holds the result of a parallel tool execution.
type toolCallOutput struct {
	tc    *ToolCall
	index int
}

// buildMessages constructs the message list with system prompt.
func (s *TaskService) buildMessages(agent *Agent) []Message {
	messages := make([]Message, 0, len(agent.Messages)+1)
	messages = append(messages, NewMessage(RoleSystem, agent.SystemPrompt))
	messages = append(messages, agent.Messages...)
	return messages
}

// collectAndPublishResults gathers parallel results and adds messages in original order.
func (s *TaskService) collectAndPublishResults(
	ctx context.Context,
	agent *Agent,
	toolCalls []ToolCall,
	outCh <-chan toolCallOutput,
	errCh <-chan error,
) int {
	// Collect results
	results := make([]toolCallOutput, 0, len(toolCalls))
	var mu sync.Mutex

	// Drain output channel
	for out := range outCh {
		mu.Lock()
		results = append(results, out)
		mu.Unlock()
	}

	// Check for errors (non-blocking)
	select {
	case err := <-errCh:
		_ = err // Individual tool failures are handled in the processor
	default:
	}

	// Sort results by original index and add messages in order
	sortedResults := make([]*ToolCall, len(toolCalls))
	for _, r := range results {
		sortedResults[r.index] = r.tc
	}

	count := 0
	for _, tc := range sortedResults {
		if tc == nil {
			continue
		}

		// Publish tool call executed event
		_ = s.eventPublisher.Publish(ctx, NewEventToolCallExecuted(
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

// completeTask marks the task as completed and publishes the event.
func (s *TaskService) completeTask(ctx context.Context, agent *Agent, task *Task, output string, state *taskState) (Result, error) {
	task.Complete(output)

	if s.hooks.AfterTask != nil {
		_ = s.hooks.AfterTask(ctx, agent, task)
	}

	_ = s.eventPublisher.Publish(ctx, NewEventTaskCompleted(string(task.ID), task.Output))

	return NewResult(task.ID, true, task.Output).
		WithDuration(time.Since(state.startTime)).
		WithIterationCount(agent.Iteration).
		WithToolCallCount(state.toolCallCount), nil
}

// createToolCallProcessor returns a function that processes a single tool call.
func (s *TaskService) createToolCallProcessor(ctx context.Context, agent *Agent) service.Function[toolCallInput, toolCallOutput] {
	return func(_ context.Context, input toolCallInput) (toolCallOutput, error) {
		tc := input.tc

		// Run before tool call hook
		if s.hooks.BeforeToolCall != nil {
			if err := s.hooks.BeforeToolCall(ctx, agent, tc); err != nil {
				tc.Fail(err.Error())
				return toolCallOutput{tc: tc, index: input.index}, nil
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

		return toolCallOutput{tc: tc, index: input.index}, nil
	}
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

// executeToolCalls runs each tool call and adds results to conversation.
// Returns the number of tool calls executed.
// If parallelTools is enabled, tool calls are executed concurrently.
func (s *TaskService) executeToolCalls(ctx context.Context, agent *Agent, toolCalls []ToolCall) int {
	if s.parallelTools && len(toolCalls) > 1 {
		return s.executeToolCallsParallel(ctx, agent, toolCalls)
	}
	return s.executeToolCallsSequential(ctx, agent, toolCalls)
}

// executeToolCallsParallel runs tool calls concurrently using the efficiency package.
// Results are collected and added to the agent in the original order.
func (s *TaskService) executeToolCallsParallel(ctx context.Context, agent *Agent, toolCalls []ToolCall) int {
	// Generate channel from tool calls
	inCh := efficiency.Generate(s.prepareToolCallInputs(toolCalls)...)

	// Process tool calls in parallel
	outCh, errCh := efficiency.Process(inCh, s.createToolCallProcessor(ctx, agent))

	// Collect and publish results
	return s.collectAndPublishResults(ctx, agent, toolCalls, outCh, errCh)
}

// executeToolCallsSequential runs tool calls one at a time (default behavior).
func (s *TaskService) executeToolCallsSequential(ctx context.Context, agent *Agent, toolCalls []ToolCall) int {
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
		_ = s.eventPublisher.Publish(ctx, NewEventToolCallExecuted(
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
	_ = s.eventPublisher.Publish(ctx, NewEventTaskFailed(string(task.ID), errMsg))

	return NewResult(task.ID, false, "").
		WithError(errMsg).
		WithDuration(time.Since(state.startTime)).
		WithIterationCount(task.Iterations).
		WithToolCallCount(state.toolCallCount), nil
}

// prepareToolCallInputs creates indexed inputs for parallel processing.
func (s *TaskService) prepareToolCallInputs(toolCalls []ToolCall) []toolCallInput {
	inputs := make([]toolCallInput, len(toolCalls))
	for i := range toolCalls {
		inputs[i] = toolCallInput{index: i, tc: &toolCalls[i]}
	}
	return inputs
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

// runBeforeTaskHook executes the before task hook if configured.
func (s *TaskService) runBeforeTaskHook(ctx context.Context, agent *Agent, task *Task) error {
	if s.hooks.BeforeTask != nil {
		return s.hooks.BeforeTask(ctx, agent, task)
	}
	return nil
}
