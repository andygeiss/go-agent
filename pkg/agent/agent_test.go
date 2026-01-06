package agent_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/pkg/agent"
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
	ag.WithMaxIterations(2)
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

func Test_Agent_WithMaxIterations_Should_SetLimit(t *testing.T) {
	// Arrange
	ag := agent.NewAgent("agent-1", "prompt")

	// Act
	ag.WithMaxIterations(5)

	// Assert
	assert.That(t, "agent max iterations must be 5", ag.MaxIterations, 5)
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
