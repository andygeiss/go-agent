package agent_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent"
)

// Task tests

func Test_Task_NewTask_With_ValidParams_Should_ReturnPendingTask(t *testing.T) {
	// Arrange
	id := agent.TaskID("task-1")
	name := "Test Task"
	input := "test input"

	// Act
	task := agent.NewTask(id, name, input)

	// Assert
	assert.That(t, "task ID must match", task.ID, id)
	assert.That(t, "task name must match", task.Name, name)
	assert.That(t, "task input must match", task.Input, input)
	assert.That(t, "task status must be pending", task.Status, agent.TaskStatusPending)
}

func Test_Task_Complete_With_Output_Should_BeCompleted(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Test", "input")
	task.Start()

	// Act
	task.Complete("task output")

	// Assert
	assert.That(t, "task status must be completed", task.Status, agent.TaskStatusCompleted)
	assert.That(t, "task output must match", task.Output, "task output")
}

func Test_Task_Fail_With_Error_Should_BeFailed(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Test", "input")
	task.Start()

	// Act
	task.Fail("task failed")

	// Assert
	assert.That(t, "task status must be failed", task.Status, agent.TaskStatusFailed)
	assert.That(t, "task error must match", task.Error, "task failed")
}

func Test_Task_IsTerminal_With_PendingTask_Should_ReturnFalse(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Test", "input")

	// Act
	isTerminal := task.IsTerminal()

	// Assert
	assert.That(t, "task must not be terminal", isTerminal, false)
}

func Test_Task_IsTerminal_With_CompletedTask_Should_ReturnTrue(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Test", "input")
	task.Complete("done")

	// Act
	isTerminal := task.IsTerminal()

	// Assert
	assert.That(t, "task must be terminal", isTerminal, true)
}

func Test_Task_Start_With_PendingTask_Should_BeRunning(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Test", "input")

	// Act
	task.Start()

	// Assert
	assert.That(t, "task status must be running", task.Status, agent.TaskStatusRunning)
}
