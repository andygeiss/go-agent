package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/chatting"
	"github.com/andygeiss/go-agent/internal/domain/tooling"
)

const defaultSystemPrompt = `You are a helpful AI assistant. You can use tools to help answer questions.
Available tools:
- get_current_time: Get the current date and time
- calculate: Perform a simple arithmetic calculation

When the user asks a question, think about whether you need to use a tool to answer it.
If you have enough information, respond directly. Be concise and helpful.`

func main() {
	// Parse command line flags
	baseURL := flag.String("url", "http://localhost:1234", "LM Studio API base URL")
	model := flag.String("model", os.Getenv("LM_STUDIO_MODEL"), "Model name to use")
	maxIterations := flag.Int("max-iterations", 10, "Maximum iterations per task")
	maxMessages := flag.Int("max-messages", 50, "Maximum messages to retain (0 = unlimited)")
	verbose := flag.Bool("verbose", false, "Show detailed metrics after each response")
	flag.Parse()

	fmt.Println("🤖 Go Agent Demo - LM Studio Chat")
	fmt.Println("==================================")
	fmt.Printf("Connecting to LM Studio at: %s\n", *baseURL)
	fmt.Printf("Using model: %s\n", *model)
	fmt.Printf("Max iterations: %d | Max messages: %d\n", *maxIterations, *maxMessages)
	fmt.Println()
	fmt.Println("Commands: 'quit'/'exit' to stop, 'clear' to reset, 'stats' for agent stats")
	fmt.Println()

	// Create the agent infrastructure
	dispatcher := messaging.NewExternalDispatcher()
	llmClient := outbound.NewOpenAIClient(*baseURL, *model)
	toolExecutor := outbound.NewToolExecutor()
	publisher := outbound.NewEventPublisher(dispatcher)

	// Register tools
	calculateTool := tooling.NewCalculateTool()
	toolExecutor.RegisterTool("calculate", calculateTool.Func)
	toolExecutor.RegisterToolDefinition(calculateTool.Definition)

	getCurrentTimeTool := tooling.NewGetCurrentTimeTool()
	toolExecutor.RegisterTool("get_current_time", getCurrentTimeTool.Func)
	toolExecutor.RegisterToolDefinition(getCurrentTimeTool.Definition)

	// Create hooks for logging (when verbose)
	hooks := agent.NewHooks()
	if *verbose {
		hooks = hooks.WithAfterToolCall(func(_ context.Context, _ *agent.Agent, tc *agent.ToolCall) error {
			fmt.Printf("  🔧 Tool: %s → %s\n", tc.Name, truncate(tc.Result, 50))
			return nil
		})
	}

	taskService := agent.NewTaskService(llmClient, toolExecutor, publisher).WithHooks(hooks)

	// Create the agent with options
	agentInstance := agent.NewAgent(
		"demo-agent",
		defaultSystemPrompt,
		agent.WithMaxIterations(*maxIterations),
		agent.WithMaxMessages(*maxMessages),
		agent.WithMetadata(agent.Metadata{
			"created_by": "cli",
			"model":      *model,
		}),
	)

	// Create use cases
	uc := &useCases{
		clearConversation: chatting.NewClearConversationUseCase(&agentInstance),
		getAgentStats:     chatting.NewGetAgentStatsUseCase(&agentInstance),
		sendMessage:       chatting.NewSendMessageUseCase(taskService, &agentInstance),
	}

	runInteractiveChat(uc, *verbose)
}

// useCases holds the domain use cases for the CLI.
type useCases struct {
	clearConversation *chatting.ClearConversationUseCase
	getAgentStats     *chatting.GetAgentStatsUseCase
	sendMessage       *chatting.SendMessageUseCase
}

// handleCommand processes special commands. Returns true if a command was handled,
// and a second bool indicating if the loop should break.
func handleCommand(input string, uc *useCases) (bool, bool) {
	switch input {
	case "clear":
		uc.clearConversation.Execute()
		fmt.Println("🗑️  Conversation cleared.")
		fmt.Println()
		return true, false
	case "exit", "quit":
		printFinalStats(uc.getAgentStats)
		fmt.Println("Goodbye! 👋")
		return true, true
	case "stats":
		printAgentStats(uc.getAgentStats)
		return true, false
	default:
		return false, false
	}
}

// printAgentStats displays the current agent statistics.
func printAgentStats(uc *chatting.GetAgentStatsUseCase) {
	stats := uc.Execute()
	fmt.Println()
	fmt.Println("📊 Agent Statistics")
	fmt.Println("-------------------")
	fmt.Printf("Agent ID:        %s\n", stats.AgentID)
	fmt.Printf("Messages:        %d\n", stats.MessageCount)
	fmt.Printf("Tasks:           %d (✓ %d completed, ✗ %d failed)\n",
		stats.TaskCount, stats.CompletedTasks, stats.FailedTasks)
	fmt.Printf("Max iterations:  %d\n", stats.MaxIterations)
	fmt.Printf("Max messages:    %d\n", stats.MaxMessages)
	if stats.Model != "" {
		fmt.Printf("Model:           %s\n", stats.Model)
	}
	fmt.Println()
}

// printFinalStats shows a summary of the session upon exit.
func printFinalStats(uc *chatting.GetAgentStatsUseCase) {
	stats := uc.Execute()
	if stats.TaskCount > 0 {
		fmt.Println()
		fmt.Printf("📈 Session summary: %d tasks (✓ %d, ✗ %d), %d messages\n",
			stats.TaskCount, stats.CompletedTasks, stats.FailedTasks, stats.MessageCount)
	}
}

// printResult displays the result of a sent message.
func printResult(output chatting.SendMessageOutput, verbose bool) {
	if output.Success {
		fmt.Printf("🤖 Assistant: %s\n", output.Response)
		if verbose {
			fmt.Printf("   ⏱️  %s | 🔄 %d iterations | 🔧 %d tool calls\n",
				output.Duration,
				output.IterationCount,
				output.ToolCallCount)
		}
		fmt.Println()
	} else {
		fmt.Printf("⚠️  Task failed: %s\n\n", output.Error)
	}
}

// runInteractiveChat starts the interactive chat loop.
func runInteractiveChat(uc *useCases, verbose bool) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if handled, shouldBreak := handleCommand(input, uc); handled {
			if shouldBreak {
				break
			}
			continue
		}

		// Send message using use case
		ctx := context.Background()
		output, err := uc.sendMessage.Execute(ctx, chatting.SendMessageInput{Message: input})
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			continue
		}

		printResult(output, verbose)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
