package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/chatting"
	"github.com/andygeiss/go-agent/internal/domain/memorizing"
	"github.com/andygeiss/go-agent/internal/domain/tooling"
)

const defaultSystemPrompt = `You are a helpful AI assistant with access to tools and long-term memory.

Available tools:
- calculate: Perform arithmetic calculations (e.g., "2 + 2", "10 * 5")
- get_current_time: Get the current date and time
- memory_get: Retrieve a specific memory note by ID
- memory_search: Search your long-term memory for relevant notes
- memory_write: Save important information to long-term memory

When the user shares preferences, important facts, or asks you to remember something,
use memory_write to save it. When they refer to past conversations or preferences,
use memory_search to recall relevant information.

Be concise, helpful, and proactive about using your memory capabilities.`

func main() {
	// Parse command line flags (alphabetically sorted)
	baseURL := flag.String("url", "http://localhost:1234", "LM Studio API base URL")
	maxIterations := flag.Int("max-iterations", 10, "Maximum iterations per task")
	maxMessages := flag.Int("max-messages", 50, "Maximum messages to retain (0 = unlimited)")
	memoryFile := flag.String("memory-file", "", "JSON file for persistent memory (empty = in-memory)")
	model := flag.String("model", os.Getenv("LM_STUDIO_MODEL"), "Model name to use")
	parallelTools := flag.Bool("parallel-tools", false, "Enable parallel tool execution")
	verbose := flag.Bool("verbose", false, "Show detailed metrics after each response")
	flag.Parse()

	// Print banner
	printBanner(*baseURL, *model, *maxIterations, *maxMessages, *memoryFile, *parallelTools)

	// Setup infrastructure
	infrastructure := setupInfrastructure(*baseURL, *model, *memoryFile, *verbose, *parallelTools)

	// Create the agent with options
	agentInstance := agent.NewAgent(
		"demo-agent",
		defaultSystemPrompt,
		agent.WithMaxIterations(*maxIterations),
		agent.WithMaxMessages(*maxMessages),
		agent.WithMetadata(agent.Metadata{
			"created_by": "cli",
			"model":      *model,
			"session_id": fmt.Sprintf("session-%d", time.Now().Unix()),
		}),
	)

	// Create use cases from all domain contexts
	uc := createUseCases(infrastructure, &agentInstance)

	// Run the interactive chat loop
	runInteractiveChat(uc, *verbose)
}

// infrastructure holds all infrastructure components.
type infrastructure struct {
	dispatcher    messaging.Dispatcher
	llmClient     *outbound.OpenAIClient
	logger        *slog.Logger
	memoryStore   *outbound.MemoryStore
	memoryToolSvc *tooling.MemoryToolService
	publisher     *outbound.EventPublisher
	taskService   *agent.TaskService
	toolExecutor  *outbound.ToolExecutor
}

// useCases holds all domain use cases for the CLI.
type useCases struct {
	// chatting context
	clearConversation *chatting.ClearConversationUseCase
	getAgentStats     *chatting.GetAgentStatsUseCase
	sendMessage       *chatting.SendMessageUseCase

	// memorizing context
	deleteNote  *memorizing.DeleteNoteUseCase
	getNote     *memorizing.GetNoteUseCase
	searchNotes *memorizing.SearchNotesUseCase
	writeNote   *memorizing.WriteNoteUseCase
}

// createUseCases initializes all domain use cases.
func createUseCases(infra *infrastructure, ag *agent.Agent) *useCases {
	return &useCases{
		// chatting context
		clearConversation: chatting.NewClearConversationUseCase(ag),
		getAgentStats:     chatting.NewGetAgentStatsUseCase(ag),
		sendMessage:       chatting.NewSendMessageUseCase(infra.taskService, ag),

		// memorizing context
		deleteNote:  memorizing.NewDeleteNoteUseCase(infra.memoryStore),
		getNote:     memorizing.NewGetNoteUseCase(infra.memoryStore),
		searchNotes: memorizing.NewSearchNotesUseCase(infra.memoryStore),
		writeNote:   memorizing.NewWriteNoteUseCase(infra.memoryStore),
	}
}

// generateNoteID creates a unique note ID.
func generateNoteID() string {
	return fmt.Sprintf("note-%d", time.Now().UnixNano())
}

// handleCommand processes special commands. Returns (handled, shouldBreak).
func handleCommand(ctx context.Context, input string, uc *useCases) (bool, bool) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false, false
	}

	cmd := strings.ToLower(parts[0])
	switch cmd {
	case "clear":
		uc.clearConversation.Execute()
		fmt.Println("🗑️  Conversation cleared.")
		fmt.Println()
		return true, false

	case "exit", "quit":
		printFinalStats(uc.getAgentStats)
		fmt.Println("Goodbye! 👋")
		return true, true

	case "help":
		printHelp()
		return true, false

	case "memory":
		handleMemoryCommand(ctx, parts[1:], uc)
		return true, false

	case "stats":
		printAgentStats(uc.getAgentStats)
		return true, false

	default:
		return false, false
	}
}

// handleMemoryCommand handles memory subcommands.
func handleMemoryCommand(ctx context.Context, args []string, uc *useCases) {
	if len(args) == 0 {
		printMemoryUsage()
		return
	}

	subcmd := strings.ToLower(args[0])
	subArgs := args[1:]

	switch subcmd {
	case "delete":
		handleMemoryDelete(ctx, subArgs, uc)
	case "get":
		handleMemoryGet(ctx, subArgs, uc)
	case "search":
		handleMemorySearch(ctx, subArgs, uc)
	case "write":
		handleMemoryWrite(ctx, subArgs, uc)
	default:
		fmt.Printf("Unknown memory command: %s\n", subcmd)
	}
}

// handleMemoryDelete handles the memory delete subcommand.
func handleMemoryDelete(ctx context.Context, args []string, uc *useCases) {
	if len(args) < 1 {
		fmt.Println("Usage: memory delete <id>")
		return
	}
	noteID := agent.NoteID(args[0])
	if err := uc.deleteNote.Execute(ctx, noteID); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("🗑️  Note %s deleted.\n", noteID)
	}
}

// handleMemoryGet handles the memory get subcommand.
func handleMemoryGet(ctx context.Context, args []string, uc *useCases) {
	if len(args) < 1 {
		fmt.Println("Usage: memory get <id>")
		return
	}
	noteID := agent.NoteID(args[0])
	note, err := uc.getNote.Execute(ctx, noteID)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	if note == nil {
		fmt.Printf("Note %s not found.\n", noteID)
		return
	}
	printMemoryNote(note)
}

// handleMemorySearch handles the memory search subcommand.
func handleMemorySearch(ctx context.Context, args []string, uc *useCases) {
	query := strings.Join(args, " ")
	if query == "" {
		query = "*" // Search all
	}
	notes, err := uc.searchNotes.Execute(ctx, query, 10, nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	printMemorySearchResults(notes)
}

// handleMemoryWrite handles the memory write subcommand.
func handleMemoryWrite(ctx context.Context, args []string, uc *useCases) {
	if len(args) < 1 {
		fmt.Println("Usage: memory write <text>")
		return
	}
	content := strings.Join(args, " ")
	note := agent.NewMemoryNote(agent.NoteID(generateNoteID()), agent.SourceTypeUserMessage).
		WithRawContent(content).
		WithSummary(content).
		WithImportance(3)
	if err := uc.writeNote.Execute(ctx, note); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("💾 Note saved with ID: %s\n", note.ID)
	}
}

// printMemoryUsage prints memory command usage information.
func printMemoryUsage() {
	fmt.Println("Usage: memory <search|get|write|delete> [args...]")
	fmt.Println("  memory search <query>     - Search memory notes")
	fmt.Println("  memory get <id>           - Get a specific note")
	fmt.Println("  memory write <text>       - Write a new note")
	fmt.Println("  memory delete <id>        - Delete a note")
	fmt.Println()
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

// printBanner displays the startup banner.
func printBanner(baseURL, model string, maxIter, maxMsg int, memoryFile string, parallelTools bool) {
	fmt.Println("🤖 Go Agent Demo - Full Feature Showcase")
	fmt.Println("========================================")
	fmt.Printf("LLM Endpoint:    %s\n", baseURL)
	fmt.Printf("Model:           %s\n", model)
	fmt.Printf("Max iterations:  %d\n", maxIter)
	fmt.Printf("Max messages:    %d\n", maxMsg)
	if memoryFile != "" {
		fmt.Printf("Memory file:     %s\n", memoryFile)
	} else {
		fmt.Println("Memory:          in-memory (ephemeral)")
	}
	fmt.Printf("Parallel tools:  %v\n", parallelTools)
	fmt.Println()
	fmt.Println("Type 'help' for available commands.")
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

// printHelp displays available commands.
func printHelp() {
	fmt.Println()
	fmt.Println("📖 Available Commands")
	fmt.Println("---------------------")
	fmt.Println("  clear              Clear conversation history")
	fmt.Println("  help               Show this help message")
	fmt.Println("  memory <subcmd>    Memory operations (search, get, write, delete)")
	fmt.Println("  quit / exit        Exit the CLI")
	fmt.Println("  stats              Show agent statistics")
	fmt.Println()
	fmt.Println("💡 Tips:")
	fmt.Println("  - Ask the agent to calculate: 'What is 42 * 17?'")
	fmt.Println("  - Ask for the time: 'What time is it?'")
	fmt.Println("  - Save to memory: 'Remember that my favorite color is blue'")
	fmt.Println("  - Recall memory: 'What is my favorite color?'")
	fmt.Println()
}

// printMemoryNote displays a single memory note.
func printMemoryNote(note *agent.MemoryNote) {
	fmt.Println()
	fmt.Println("📝 Memory Note")
	fmt.Println("--------------")
	fmt.Printf("ID:          %s\n", note.ID)
	fmt.Printf("Type:        %s\n", note.SourceType)
	fmt.Printf("Summary:     %s\n", note.Summary)
	fmt.Printf("Content:     %s\n", note.RawContent)
	fmt.Printf("Importance:  %d/5\n", note.Importance)
	if len(note.Tags) > 0 {
		fmt.Printf("Tags:        %s\n", strings.Join(note.Tags, ", "))
	}
	if len(note.Keywords) > 0 {
		fmt.Printf("Keywords:    %s\n", strings.Join(note.Keywords, ", "))
	}
	fmt.Printf("Created:     %s\n", note.CreatedAt.Format(time.RFC3339))
	fmt.Println()
}

// printMemorySearchResults displays memory search results.
func printMemorySearchResults(notes []*agent.MemoryNote) {
	fmt.Println()
	fmt.Printf("🔍 Found %d memory note(s)\n", len(notes))
	fmt.Println("--------------------------")
	if len(notes) == 0 {
		fmt.Println("No notes found.")
	}
	for _, note := range notes {
		fmt.Printf("  [%s] (%s, importance: %d) %s\n",
			note.ID, note.SourceType, note.Importance, truncate(note.Summary, 60))
	}
	fmt.Println()
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
	ctx := context.Background()

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if handled, shouldBreak := handleCommand(ctx, input, uc); handled {
			if shouldBreak {
				break
			}
			continue
		}

		// Send message using use case
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

// setupInfrastructure creates and wires all infrastructure components.
func setupInfrastructure(baseURL, model, memoryFile string, verbose, parallelTools bool) *infrastructure {
	logger := createLogger(verbose)
	dispatcher := messaging.NewExternalDispatcher()
	publisher := outbound.NewEventPublisher(dispatcher)
	memoryStore := createMemoryStore(memoryFile)
	memoryToolSvc := tooling.NewMemoryToolService(memoryStore, generateNoteID)
	toolExecutor := createToolExecutor(verbose, logger, memoryToolSvc)
	llmClient := createLLMClient(baseURL, model, verbose, logger)
	hooks := createHooks(verbose)
	taskService := createTaskService(llmClient, toolExecutor, publisher, hooks, parallelTools)

	return &infrastructure{
		dispatcher:    dispatcher,
		llmClient:     llmClient,
		logger:        logger,
		memoryStore:   memoryStore,
		memoryToolSvc: memoryToolSvc,
		publisher:     publisher,
		taskService:   taskService,
		toolExecutor:  toolExecutor,
	}
}

// createHooks creates task lifecycle hooks (verbose mode enables all hooks).
func createHooks(verbose bool) agent.Hooks {
	hooks := agent.NewHooks()
	if !verbose {
		return hooks
	}
	return hooks.
		WithBeforeTask(func(_ context.Context, _ *agent.Agent, t *agent.Task) error {
			fmt.Printf("   📋 Task started: %s\n", t.Name)
			return nil
		}).
		WithBeforeLLMCall(func(_ context.Context, _ *agent.Agent, _ *agent.Task) error {
			fmt.Println("   🔄 Calling LLM...")
			return nil
		}).
		WithAfterLLMCall(func(_ context.Context, _ *agent.Agent, _ *agent.Task) error {
			fmt.Println("   ✅ LLM response received")
			return nil
		}).
		WithBeforeToolCall(func(_ context.Context, _ *agent.Agent, tc *agent.ToolCall) error {
			fmt.Printf("   🔧 Executing tool: %s\n", tc.Name)
			return nil
		}).
		WithAfterToolCall(func(_ context.Context, _ *agent.Agent, tc *agent.ToolCall) error {
			fmt.Printf("   ✅ Tool result: %s\n", truncate(tc.Result, 50))
			return nil
		}).
		WithAfterTask(func(_ context.Context, _ *agent.Agent, t *agent.Task) error {
			fmt.Printf("   📋 Task completed: %s\n", t.Status)
			return nil
		})
}

// createLLMClient creates the OpenAI client with optional logging.
func createLLMClient(baseURL, model string, verbose bool, logger *slog.Logger) *outbound.OpenAIClient {
	client := outbound.NewOpenAIClient(baseURL, model)
	if verbose && logger != nil {
		client = client.WithLogger(logger)
	}
	return client
}

// createLogger creates a logger for verbose mode.
func createLogger(verbose bool) *slog.Logger {
	if !verbose {
		return nil
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// createMemoryStore creates either a file-backed or in-memory store.
func createMemoryStore(memoryFile string) *outbound.MemoryStore {
	if memoryFile != "" {
		return outbound.NewJsonFileMemoryStore(memoryFile)
	}
	return outbound.NewInMemoryMemoryStore()
}

// createTaskService creates the task service with hooks and optional parallelism.
func createTaskService(
	llmClient agent.LLMClient,
	toolExecutor agent.ToolExecutor,
	publisher *outbound.EventPublisher,
	hooks agent.Hooks,
	parallelTools bool,
) *agent.TaskService {
	svc := agent.NewTaskService(llmClient, toolExecutor, publisher).WithHooks(hooks)
	if parallelTools {
		svc = svc.WithParallelToolExecution()
	}
	return svc
}

// createToolExecutor creates and configures the tool executor with all tools.
func createToolExecutor(verbose bool, logger *slog.Logger, memoryToolSvc *tooling.MemoryToolService) *outbound.ToolExecutor {
	executor := outbound.NewToolExecutor()
	if verbose && logger != nil {
		executor = executor.WithLogger(logger)
	}
	registerTools(executor, memoryToolSvc)
	return executor
}

// registerTools registers all available tools with the executor.
func registerTools(executor *outbound.ToolExecutor, memoryToolSvc *tooling.MemoryToolService) {
	// Register calculate tool
	calculateTool := tooling.NewCalculateTool()
	executor.RegisterTool(string(calculateTool.ID), calculateTool.Func)
	executor.RegisterToolDefinition(calculateTool.Definition)

	// Register get_current_time tool
	timeTool := tooling.NewGetCurrentTimeTool()
	executor.RegisterTool(string(timeTool.ID), timeTool.Func)
	executor.RegisterToolDefinition(timeTool.Definition)

	// Register memory_get tool
	memoryGetTool := tooling.NewMemoryGetTool(memoryToolSvc)
	executor.RegisterTool(string(memoryGetTool.ID), memoryGetTool.Func)
	executor.RegisterToolDefinition(memoryGetTool.Definition)

	// Register memory_search tool
	memorySearchTool := tooling.NewMemorySearchTool(memoryToolSvc)
	executor.RegisterTool(string(memorySearchTool.ID), memorySearchTool.Func)
	executor.RegisterToolDefinition(memorySearchTool.Definition)

	// Register memory_write tool
	memoryWriteTool := tooling.NewMemoryWriteTool(memoryToolSvc)
	executor.RegisterTool(string(memoryWriteTool.ID), memoryWriteTool.Func)
	executor.RegisterToolDefinition(memoryWriteTool.Definition)
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	// Remove newlines for cleaner display
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
