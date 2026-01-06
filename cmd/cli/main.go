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
	"github.com/andygeiss/go-agent/internal/domain/agent/aggregates"
	"github.com/andygeiss/go-agent/internal/domain/agent/entities"
	"github.com/andygeiss/go-agent/internal/domain/agent/immutable"
	"github.com/andygeiss/go-agent/internal/domain/agent/services"
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
	flag.Parse()

	fmt.Println("🤖 Go Agent Demo - LM Studio Chat")
	fmt.Println("==================================")
	fmt.Printf("Connecting to LM Studio at: %s\n", *baseURL)
	fmt.Printf("Using model: %s\n", *model)
	fmt.Println()
	fmt.Println("Type your message and press Enter. Type 'quit' or 'exit' to stop.")
	fmt.Println()

	// Create the agent infrastructure
	dispatcher := messaging.NewExternalDispatcher()
	llmClient := outbound.NewOpenAIClient(*baseURL, *model)
	toolExecutor := outbound.NewToolExecutor()
	publisher := outbound.NewEventPublisher(dispatcher)
	taskService := services.NewTaskService(llmClient, toolExecutor, publisher)

	// Create the agent
	agentInstance := aggregates.NewAgent("demo-agent", defaultSystemPrompt)

	runInteractiveChat(taskService, &agentInstance)
}

func runInteractiveChat(taskService *services.TaskService, agent *aggregates.Agent) {
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

		if input == "quit" || input == "exit" {
			fmt.Println("Goodbye! 👋")
			break
		}

		if input == "clear" {
			agent.ClearMessages()
			fmt.Println("🗑️  Conversation cleared.")
			fmt.Println()
			continue
		}

		// Create and run a task
		taskCounter++
		taskID := immutable.TaskID(fmt.Sprintf("task-%d", taskCounter))
		task := entities.NewTask(taskID, "chat", input)

		ctx := context.Background()
		result, err := taskService.RunTask(ctx, agent, task)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			continue
		}

		if result.Success {
			fmt.Printf("🤖 Assistant: %s\n\n", result.Output)
		} else {
			fmt.Printf("⚠️  Task failed: %s\n\n", result.Error)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}
