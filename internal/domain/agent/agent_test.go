package agent_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent"
)

func Test_Agent_NewAgent_With_ValidParams_Should_ReturnAgent(t *testing.T) {
	// Arrange
	id := agent.AgentID("agent-1")
	prompt := "You are helpful"

	// Act
	ag := agent.NewAgent(id, prompt)

	// Assert
	assert.That(t, "agent ID must match", ag.ID, id)
	assert.That(t, "agent prompt must match", ag.SystemPrompt, prompt)
	assert.That(t, "agent iteration must be 0", ag.Iteration, 0)
	assert.That(t, "agent max iterations must be default", ag.MaxIterations, 10)
}

func Test_Agent_AddTask_With_Task_Should_AddToQueue(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")
	task := agent.NewTask("task-1", "Test Task", "test input")

	// Act
	ag.AddTask(task)

	// Assert
	assert.That(t, "agent must have pending task", ag.HasPendingTasks(), true)
}

func Test_Agent_GetCurrentTask_With_NoTasks_Should_ReturnNil(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")

	// Act
	task := ag.GetCurrentTask()

	// Assert
	assert.That(t, "current task must be nil", task == nil, true)
}

func Test_Agent_GetCurrentTask_With_Task_Should_ReturnTask(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")
	task := agent.NewTask("task-1", "Test Task", "test input")
	ag.AddTask(task)

	// Act
	currentTask := ag.GetCurrentTask()

	// Assert
	assert.That(t, "current task must not be nil", currentTask != nil, true)
	assert.That(t, "current task ID must match", currentTask.ID, task.ID)
}

func Test_Agent_AddMessage_With_Message_Should_AddToHistory(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")
	msg := agent.NewMessage(agent.RoleUser, "hello")

	// Act
	ag.AddMessage(msg)

	// Assert
	messages := ag.GetMessages()
	assert.That(t, "agent must have one message", len(messages), 1)
	assert.That(t, "message content must match", messages[0].Content, "hello")
}

func Test_Agent_CanContinue_With_LowIteration_Should_ReturnTrue(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")

	// Act
	canContinue := ag.CanContinue()

	// Assert
	assert.That(t, "agent must be able to continue", canContinue, true)
}

func Test_Agent_CanContinue_With_MaxIteration_Should_ReturnFalse(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")
	ag.SetMaxIterations(2)
	ag.IncrementIteration()
	ag.IncrementIteration()

	// Act
	canContinue := ag.CanContinue()

	// Assert
	assert.That(t, "agent must not be able to continue", canContinue, false)
}

func Test_Agent_ClearMessages_Should_ClearHistory(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")
	ag.AddMessage(agent.NewMessage(agent.RoleUser, "hello"))
	ag.AddMessage(agent.NewMessage(agent.RoleAssistant, "hi"))

	// Act
	ag.ClearMessages()

	// Assert
	assert.That(t, "agent must have no messages", len(ag.GetMessages()), 0)
}

func Test_Agent_GetMessages_With_Messages_Should_ReturnAll(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")
	ag.AddMessage(agent.NewMessage(agent.RoleUser, "hello"))
	ag.AddMessage(agent.NewMessage(agent.RoleAssistant, "hi"))

	// Act
	messages := ag.GetMessages()

	// Assert
	assert.That(t, "agent must have two messages", len(messages), 2)
}

func Test_Agent_HasPendingTasks_With_NoTasks_Should_ReturnFalse(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")

	// Act
	hasPending := ag.HasPendingTasks()

	// Assert
	assert.That(t, "agent must not have pending tasks", hasPending, false)
}

func Test_Agent_HasPendingTasks_With_Tasks_Should_ReturnTrue(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")
	ag.AddTask(agent.NewTask("task-1", "Test", "input"))

	// Act
	hasPending := ag.HasPendingTasks()

	// Assert
	assert.That(t, "agent must have pending tasks", hasPending, true)
}

func Test_Agent_ResetIteration_Should_ResetCounter(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")
	ag.IncrementIteration()
	ag.IncrementIteration()

	// Act
	ag.ResetIteration()

	// Assert
	assert.That(t, "agent iteration must be 0", ag.Iteration, 0)
}

func Test_Agent_SetMaxIterations_Should_SetLimit(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")

	// Act
	ag.SetMaxIterations(5)

	// Assert
	assert.That(t, "agent max iterations must be 5", ag.MaxIterations, 5)
}

func Test_Agent_WithMaxIterations_Option_Should_SetLimit(t *testing.T) {
	// Arrange & Act
	ag := agent.NewAgent("agent-1", "prompt", agent.WithMaxIterations(15))

	// Assert
	assert.That(t, "agent max iterations must be 15", ag.MaxIterations, 15)
}

func Test_Agent_WithMaxMessages_Option_Should_SetLimit(t *testing.T) {
	// Arrange & Act
	ag := agent.NewAgent("agent-1", "prompt", agent.WithMaxMessages(100))

	// Assert
	assert.That(t, "agent max messages must be 100", ag.MaxMessages, 100)
}

func Test_Agent_WithMetadata_Option_Should_SetMetadata(t *testing.T) {
	// Arrange & Act
	meta := agent.Metadata{"key": "value"}
	ag := agent.NewAgent("agent-1", "prompt", agent.WithMetadata(meta))

	// Assert
	assert.That(t, "agent metadata must be set", ag.GetMetadata("key"), "value")
}

func Test_Agent_AddMessage_With_MaxMessages_Should_TrimOldMessages(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt", agent.WithMaxMessages(2))
	ag.AddMessage(agent.NewMessage(agent.RoleUser, "first"))
	ag.AddMessage(agent.NewMessage(agent.RoleAssistant, "second"))

	// Act
	ag.AddMessage(agent.NewMessage(agent.RoleUser, "third"))

	// Assert
	messages := ag.GetMessages()
	assert.That(t, "agent must have 2 messages", len(messages), 2)
	assert.That(t, "first message must be second", messages[0].Content, "second")
	assert.That(t, "second message must be third", messages[1].Content, "third")
}

func Test_Agent_SetMetadata_Should_StoreValue(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")

	// Act
	ag.SetMetadata("session_id", "abc123")

	// Assert
	assert.That(t, "metadata must be stored", ag.GetMetadata("session_id"), "abc123")
}

func Test_Agent_GetMetadata_With_MissingKey_Should_ReturnEmpty(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")

	// Act
	value := ag.GetMetadata("nonexistent")

	// Assert
	assert.That(t, "missing key must return empty string", value, "")
}

func Test_Agent_MessageCount_Should_ReturnCount(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")
	ag.AddMessage(agent.NewMessage(agent.RoleUser, "hello"))
	ag.AddMessage(agent.NewMessage(agent.RoleAssistant, "hi"))

	// Act
	count := ag.MessageCount()

	// Assert
	assert.That(t, "message count must be 2", count, 2)
}

func Test_Agent_TaskCount_Should_ReturnCount(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")
	ag.AddTask(agent.NewTask("task-1", "Task 1", "input"))
	ag.AddTask(agent.NewTask("task-2", "Task 2", "input"))

	// Act
	count := ag.TaskCount()

	// Assert
	assert.That(t, "task count must be 2", count, 2)
}

func Test_Agent_CompletedTaskCount_Should_ReturnCompletedCount(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")
	task1 := agent.NewTask("task-1", "Task 1", "input")
	task2 := agent.NewTask("task-2", "Task 2", "input")
	task1.Complete("done")
	ag.AddTask(task1)
	ag.AddTask(task2)

	// Act
	count := ag.CompletedTaskCount()

	// Assert
	assert.That(t, "completed task count must be 1", count, 1)
}

func Test_Agent_FailedTaskCount_Should_ReturnFailedCount(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")
	task1 := agent.NewTask("task-1", "Task 1", "input")
	task2 := agent.NewTask("task-2", "Task 2", "input")
	task1.Fail("error")
	ag.AddTask(task1)
	ag.AddTask(task2)

	// Act
	count := ag.FailedTaskCount()

	// Assert
	assert.That(t, "failed task count must be 1", count, 1)
}

func Test_LLMResponse_NewLLMResponse_Should_CreateResponse(t *testing.T) {
	// Arrange
	msg := agent.NewMessage(agent.RoleAssistant, "Hello")
	finishReason := "stop"

	// Act
	response := agent.NewLLMResponse(msg, finishReason)

	// Assert
	assert.That(t, "response message must match", response.Message.Content, "Hello")
	assert.That(t, "response finish reason must match", response.FinishReason, "stop")
}

func Test_LLMResponse_HasToolCalls_With_NoToolCalls_Should_ReturnFalse(t *testing.T) {
	// Arrange
	response := agent.NewLLMResponse(
		agent.NewMessage(agent.RoleAssistant, "Hello"),
		"stop",
	)

	// Act
	hasToolCalls := response.HasToolCalls()

	// Assert
	assert.That(t, "response must not have tool calls", hasToolCalls, false)
}

func Test_LLMResponse_HasToolCalls_With_ToolCalls_Should_ReturnTrue(t *testing.T) {
	// Arrange
	response := agent.NewLLMResponse(
		agent.NewMessage(agent.RoleAssistant, ""),
		"tool_calls",
	).WithToolCalls([]agent.ToolCall{
		agent.NewToolCall("tc-1", "search", `{}`),
	})

	// Act
	hasToolCalls := response.HasToolCalls()

	// Assert
	assert.That(t, "response must have tool calls", hasToolCalls, true)
}
