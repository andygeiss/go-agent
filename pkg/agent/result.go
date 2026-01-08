package agent

import "time"

// Result represents the outcome of a task execution.
// It indicates success/failure and contains the output or error.
type Result struct {
	Error          string        // Error message if failed
	Output         string        // The output if successful
	TaskID         TaskID        // ID of the task that produced this result
	Tokens         TokenUsage    // Token usage statistics
	Duration       time.Duration // How long the task took to execute
	IterationCount int           // Number of agent loop iterations
	ToolCallCount  int           // Number of tool calls made
	Success        bool          // Whether the task completed successfully
}

// TokenUsage tracks the number of tokens used in an LLM interaction.
type TokenUsage struct {
	CompletionTokens int // Tokens in the output/completion
	PromptTokens     int // Tokens in the input/prompt
	TotalTokens      int // Total tokens used
}

// NewResult creates a new Result for the given task.
func NewResult(taskID TaskID, success bool, output string) Result {
	return Result{
		Output:  output,
		Success: success,
		TaskID:  taskID,
	}
}

// WithDuration sets the execution duration on the result.
func (r Result) WithDuration(d time.Duration) Result {
	r.Duration = d
	return r
}

// WithError sets an error message on the result.
func (r Result) WithError(errMsg string) Result {
	r.Error = errMsg
	return r
}

// WithIterationCount sets the iteration count on the result.
func (r Result) WithIterationCount(count int) Result {
	r.IterationCount = count
	return r
}

// WithTokens sets the token usage on the result.
func (r Result) WithTokens(tokens TokenUsage) Result {
	r.Tokens = tokens
	return r
}

// WithToolCallCount sets the tool call count on the result.
func (r Result) WithToolCallCount(count int) Result {
	r.ToolCallCount = count
	return r
}
