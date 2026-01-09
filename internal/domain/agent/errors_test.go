package agent_test

import (
	"errors"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent"
)

func Test_ErrMaxIterationsReached_Error_Should_ReturnMessage(t *testing.T) {
	// Act
	msg := agent.ErrMaxIterationsReached.Error()

	// Assert
	assert.That(t, "error message must match", msg, "max iterations reached")
}

func Test_LLMError_Error_With_Cause_Should_IncludeCause(t *testing.T) {
	// Arrange
	cause := errors.New("connection timeout")
	err := agent.NewLLMError("failed to call LLM", cause)

	// Act
	msg := err.Error()

	// Assert
	assert.That(t, "error message must include cause", msg, "failed to call LLM: connection timeout")
}

func Test_LLMError_Error_Without_Cause_Should_ReturnMessage(t *testing.T) {
	// Arrange
	err := agent.NewLLMError("failed to call LLM", nil)

	// Act
	msg := err.Error()

	// Assert
	assert.That(t, "error message must match", msg, "failed to call LLM")
}

func Test_LLMError_Unwrap_Should_ReturnCause(t *testing.T) {
	// Arrange
	cause := errors.New("connection timeout")
	err := agent.NewLLMError("failed to call LLM", cause)

	// Act
	unwrapped := err.Unwrap()

	// Assert
	assert.That(t, "unwrapped error must match cause", unwrapped, cause)
}

func Test_ToolError_Error_Should_IncludeToolName(t *testing.T) {
	// Arrange
	cause := errors.New("invalid input")
	err := agent.NewToolError("search", "execution failed", cause)

	// Act
	msg := err.Error()

	// Assert
	assert.That(t, "error message must include tool name", msg, "tool search: execution failed: invalid input")
}

func Test_TaskError_Error_Should_IncludeTaskID(t *testing.T) {
	// Arrange
	cause := errors.New("timeout")
	err := agent.NewTaskError("task-123", "execution failed", cause)

	// Act
	msg := err.Error()

	// Assert
	assert.That(t, "error message must include task ID", msg, "task task-123: execution failed: timeout")
}

func Test_ErrorsIs_With_LLMError_And_Cause_Should_Match(t *testing.T) {
	// Arrange
	cause := agent.ErrNoResponse
	err := agent.NewLLMError("LLM error", cause)

	// Act
	matches := errors.Is(err, agent.ErrNoResponse)

	// Assert
	assert.That(t, "errors.Is must match cause", matches, true)
}
