package entities_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent/entities"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
)

func Test_Result_NewResult_With_Success_Should_ReturnSuccessResult(t *testing.T) {
	// Arrange
	taskID := immutable.TaskID("task-1")
	output := "Task completed successfully"

	// Act
	result := entities.NewResult(taskID, true, output)

	// Assert
	assert.That(t, "result task ID must match", result.TaskID, taskID)
	assert.That(t, "result success must be true", result.Success, true)
	assert.That(t, "result output must match", result.Output, output)
}

func Test_Result_WithError_With_ErrorMessage_Should_HaveError(t *testing.T) {
	// Arrange
	result := entities.NewResult("task-1", false, "")

	// Act
	result = result.WithError("something failed")

	// Assert
	assert.That(t, "result error must match", result.Error, "something failed")
}
