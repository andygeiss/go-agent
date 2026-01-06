package entities_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent/entities"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
)

func Test_Task_NewTask_With_ValidParams_Should_ReturnPendingTask(t *testing.T) {
	// Arrange
	id := immutable.TaskID("task-1")
	name := "Test Task"
	input := "Do something"

	// Act
	task := entities.NewTask(id, name, input)

	// Assert
	assert.That(t, "task ID must match", task.ID, id)
	assert.That(t, "task name must match", task.Name, name)
	assert.That(t, "task input must match", task.Input, input)
	assert.That(t, "task status must be pending", task.Status, immutable.TaskStatusPending)
}

func Test_Task_Complete_With_Output_Should_BeCompleted(t *testing.T) {
	// Arrange
	task := entities.NewTask("task-1", "Task", "input")
	task.Start()

	// Act
	task.Complete("output result")

	// Assert
	assert.That(t, "task status must be completed", task.Status, immutable.TaskStatusCompleted)
	assert.That(t, "task output must match", task.Output, "output result")
}

func Test_Task_Fail_With_Error_Should_BeFailed(t *testing.T) {
	// Arrange
	task := entities.NewTask("task-1", "Task", "input")
	task.Start()

	// Act
	task.Fail("something went wrong")

	// Assert
	assert.That(t, "task status must be failed", task.Status, immutable.TaskStatusFailed)
	assert.That(t, "task error must match", task.Error, "something went wrong")
}

func Test_Task_IsTerminal_With_CompletedTask_Should_ReturnTrue(t *testing.T) {
	// Arrange
	task := entities.NewTask("task-1", "Task", "input")
	task.Complete("done")

	// Act
	isTerminal := task.IsTerminal()

	// Assert
	assert.That(t, "completed task must be terminal", isTerminal, true)
}

func Test_Task_IsTerminal_With_FailedTask_Should_ReturnTrue(t *testing.T) {
	// Arrange
	task := entities.NewTask("task-1", "Task", "input")
	task.Fail("error")

	// Act
	isTerminal := task.IsTerminal()

	// Assert
	assert.That(t, "failed task must be terminal", isTerminal, true)
}

func Test_Task_IsTerminal_With_PendingTask_Should_ReturnFalse(t *testing.T) {
	// Arrange
	task := entities.NewTask("task-1", "Task", "input")

	// Act
	isTerminal := task.IsTerminal()

	// Assert
	assert.That(t, "pending task must not be terminal", isTerminal, false)
}

func Test_Task_Start_With_PendingTask_Should_BeInProgress(t *testing.T) {
	// Arrange
	task := entities.NewTask("task-1", "Task", "input")

	// Act
	task.Start()

	// Assert
	assert.That(t, "task status must be in progress", task.Status, immutable.TaskStatusInProgress)
}
