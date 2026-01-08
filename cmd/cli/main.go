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
	"github.com/andygeiss/go-agent/pkg/agent"
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

	runInteractiveChat(taskService, &agentInstance, *verbose)
}

// handleCommand processes special commands. Returns true if a command was handled,
// and a second bool indicating if the loop should break.
func handleCommand(input string, ag *agent.Agent) (bool, bool) {
	switch input {
	case "quit", "exit":
		printFinalStats(ag)
		fmt.Println("Goodbye! 👋")
		return true, true
	case "clear":
		ag.ClearMessages()
		fmt.Println("🗑️  Conversation cleared.")
		fmt.Println()
		return true, false
	case "stats":
		printAgentStats(ag)
		return true, false
	default:
		return false, false
	}
}

func runInteractiveChat(taskService *agent.TaskService, ag *agent.Agent, verbose bool) {
	scanner := bufio.NewScanner(os.Stdin)
	taskCounter := 0

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if handled, shouldBreak := handleCommand(input, ag); handled {
			if shouldBreak {
				break
			}
			continue
		}

		// Create and run a task
		taskCounter++
		taskID := agent.TaskID(fmt.Sprintf("task-%d", taskCounter))
		task := agent.NewTask(taskID, "chat", input)

		ctx := context.Background()
		result, err := taskService.RunTask(ctx, ag, task)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			continue
		}

		printResult(result, verbose)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}

func printResult(result agent.Result, verbose bool) {
	if result.Success {
		fmt.Printf("🤖 Assistant: %s\n", result.Output)
		if verbose {
			fmt.Printf("   ⏱️  %s | 🔄 %d iterations | 🔧 %d tool calls\n",
				result.Duration.Round(1000000), // Round to milliseconds
				result.IterationCount,
				result.ToolCallCount)
		}
		fmt.Println()
	} else {
		fmt.Printf("⚠️  Task failed: %s\n\n", result.Error)
	}
}

func printAgentStats(ag *agent.Agent) {
	fmt.Println()
	fmt.Println("📊 Agent Statistics")
	fmt.Println("-------------------")
	fmt.Printf("Agent ID:        %s\n", ag.ID)
	fmt.Printf("Messages:        %d\n", ag.MessageCount())
	fmt.Printf("Tasks:           %d (✓ %d completed, ✗ %d failed)\n",
		ag.TaskCount(), ag.CompletedTaskCount(), ag.FailedTaskCount())
	fmt.Printf("Max iterations:  %d\n", ag.MaxIterations)
	fmt.Printf("Max messages:    %d\n", ag.MaxMessages)
	if model := ag.GetMetadata("model"); model != "" {
		fmt.Printf("Model:           %s\n", model)
	}
	fmt.Println()
}

func printFinalStats(ag *agent.Agent) {
	if ag.TaskCount() > 0 {
		fmt.Println()
		fmt.Printf("📈 Session summary: %d tasks (✓ %d, ✗ %d), %d messages\n",
			ag.TaskCount(), ag.CompletedTaskCount(), ag.FailedTaskCount(), ag.MessageCount())
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
