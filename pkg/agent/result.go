package agent

import "time"

// TokenUsage tracks the number of tokens used in an LLM interaction.
type TokenUsage struct {
	PromptTokens     int // Tokens in the input/prompt
	CompletionTokens int // Tokens in the output/completion
	TotalTokens      int // Total tokens used
}

// Result represents the outcome of a task execution.
// It indicates success/failure and contains the output or error.
type Result struct {
	TaskID         TaskID        // ID of the task that produced this result
	Output         string        // The output if successful
	Error          string        // Error message if failed
	Success        bool          // Whether the task completed successfully
	Duration       time.Duration // How long the task took to execute
	IterationCount int           // Number of agent loop iterations
	ToolCallCount  int           // Number of tool calls made
	Tokens         TokenUsage    // Token usage statistics
}

// NewResult creates a new Result for the given task.
func NewResult(taskID TaskID, success bool, output string) Result {
	return Result{
		TaskID:  taskID,
		Success: success,
		Output:  output,
	}
}

// WithError sets an error message on the result.
func (r Result) WithError(errMsg string) Result {
	r.Error = errMsg
	return r
}

// WithDuration sets the execution duration on the result.
func (r Result) WithDuration(d time.Duration) Result {
	r.Duration = d
	return r
}

// WithIterationCount sets the iteration count on the result.
func (r Result) WithIterationCount(count int) Result {
	r.IterationCount = count
	return r
}

// WithToolCallCount sets the tool call count on the result.
func (r Result) WithToolCallCount(count int) Result {
	r.ToolCallCount = count
	return r
}

// WithTokens sets the token usage on the result.
func (r Result) WithTokens(tokens TokenUsage) Result {
	r.Tokens = tokens
	return r
}
