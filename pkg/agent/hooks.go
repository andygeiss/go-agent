package agent

import (
	"context"
)

// Hook represents a function that can be called at specific points during task execution.
// Hooks can inspect or modify the context, returning an error to abort execution.
type Hook func(ctx context.Context, agent *Agent, task *Task) error

// Hooks contains callback functions for various points in the task lifecycle.
type Hooks struct {
	// BeforeTask is called before a task starts executing.
	// Can be used for logging, validation, or setup.
	BeforeTask Hook

	// AfterTask is called after a task completes (success or failure).
	// Can be used for cleanup, metrics, or logging.
	AfterTask Hook

	// BeforeLLMCall is called before each LLM request.
	// Can be used for rate limiting, logging, or request modification.
	BeforeLLMCall Hook

	// AfterLLMCall is called after each LLM response is received.
	// Can be used for logging, caching, or response inspection.
	AfterLLMCall Hook

	// BeforeToolCall is called before each tool execution.
	// Can be used for authorization, logging, or argument validation.
	BeforeToolCall func(ctx context.Context, agent *Agent, toolCall *ToolCall) error

	// AfterToolCall is called after each tool execution completes.
	// Can be used for logging, result caching, or result modification.
	AfterToolCall func(ctx context.Context, agent *Agent, toolCall *ToolCall) error
}

// NewHooks creates an empty Hooks struct.
func NewHooks() Hooks {
	return Hooks{}
}

// WithBeforeTask sets the before task hook.
func (h Hooks) WithBeforeTask(hook Hook) Hooks {
	h.BeforeTask = hook
	return h
}

// WithAfterTask sets the after task hook.
func (h Hooks) WithAfterTask(hook Hook) Hooks {
	h.AfterTask = hook
	return h
}

// WithBeforeLLMCall sets the before LLM call hook.
func (h Hooks) WithBeforeLLMCall(hook Hook) Hooks {
	h.BeforeLLMCall = hook
	return h
}

// WithAfterLLMCall sets the after LLM call hook.
func (h Hooks) WithAfterLLMCall(hook Hook) Hooks {
	h.AfterLLMCall = hook
	return h
}

// WithBeforeToolCall sets the before tool call hook.
func (h Hooks) WithBeforeToolCall(hook func(ctx context.Context, agent *Agent, toolCall *ToolCall) error) Hooks {
	h.BeforeToolCall = hook
	return h
}

// WithAfterToolCall sets the after tool call hook.
func (h Hooks) WithAfterToolCall(hook func(ctx context.Context, agent *Agent, toolCall *ToolCall) error) Hooks {
	h.AfterToolCall = hook
	return h
}
