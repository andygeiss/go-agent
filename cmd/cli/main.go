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
	"github.com/andygeiss/go-agent/internal/adapters/inbound"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/chatting"
	"github.com/andygeiss/go-agent/internal/domain/indexing"
	"github.com/andygeiss/go-agent/internal/domain/memorizing"
	"github.com/andygeiss/go-agent/internal/domain/tooling"
)

const defaultSystemPrompt = `You are a helpful AI assistant with access to tools and long-term memory.

Available tools:
- index.scan: Scan directories and create a snapshot of files
- index.changed_since: Get files changed since a timestamp
- index.diff_snapshot: Compare two snapshots for changes
- memory_get: Retrieve a specific memory note by ID
- memory_search: Search your long-term memory for relevant notes
- memory_write: Save important information to long-term memory

When the user shares preferences, important facts, or asks you to remember something,
use memory_write to save it. When they refer to past conversations or preferences,
use memory_search to recall relevant information.

When asked to analyze code changes or track file modifications, use the indexing tools
to scan directories, compare snapshots, and identify what has changed.

Be concise, helpful, and proactive about using your memory and indexing capabilities.`

func main() {
	// Parse command line flags (alphabetically sorted)
	chattingModel := flag.String("chatting-model", os.Getenv("OPENAI_CHAT_MODEL"), "Model name to use")
	chattingURL := flag.String("chatting-url", "http://localhost:1234", "OpenAI API base URL")
	embeddingModel := flag.String("embedding-model", os.Getenv("OPENAI_EMBED_MODEL"), "Embedding model name (empty = no embeddings)")
	embeddingURL := flag.String("embedding-url", getEnvOrDefault("OPENAI_EMBED_URL", "http://localhost:1234"), "Embedding API URL (defaults to -chatting-url if not set)")
	indexFile := flag.String("index-file", "", "JSON file for persistent indexing (empty = in-memory)")
	maxIterations := flag.Int("max-iterations", 10, "Maximum iterations per task")
	maxMessages := flag.Int("max-messages", 50, "Maximum messages to retain (0 = unlimited)")
	memoryFile := flag.String("memory-file", "", "JSON file for persistent memory (empty = in-memory)")
	parallelTools := flag.Bool("parallel-tools", false, "Enable parallel tool execution")
	verbose := flag.Bool("verbose", false, "Show detailed metrics after each response")
	flag.Parse()

	// Print banner
	embURL := *embeddingURL
	if embURL == "" {
		embURL = *chattingURL
	}
	printBanner(*chattingURL, *chattingModel, embURL, *embeddingModel, *maxIterations, *maxMessages, *memoryFile, *indexFile, *parallelTools)

	// Setup infrastructure
	infrastructure := setupInfrastructure(*chattingURL, *chattingModel, *memoryFile, *indexFile, *verbose, *parallelTools, embURL, *embeddingModel)

	// Create the agent with options
	agentInstance := agent.NewAgent(
		"demo-agent",
		defaultSystemPrompt,
		agent.WithMaxIterations(*maxIterations),
		agent.WithMaxMessages(*maxMessages),
		agent.WithMetadata(agent.Metadata{
			"created_by": "cli",
			"model":      *chattingModel,
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
	indexService  *indexing.Service
	indexToolSvc  *tooling.IndexToolService
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

	// indexing context
	indexService *indexing.Service

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

		// indexing context
		indexService: infra.indexService,

		// memorizing context
		deleteNote:  memorizing.NewDeleteNoteUseCase(infra.memoryStore),
		getNote:     memorizing.NewGetNoteUseCase(infra.memoryStore),
		searchNotes: memorizing.NewSearchNotesUseCase(infra.memoryStore),
		writeNote:   memorizing.NewWriteNoteUseCase(infra.memoryStore),
	}
}

// memoryFlags holds parsed flags for memory commands.
type memoryFlags struct {
	sourceType  agent.SourceType
	sourceTypes []agent.SourceType
	tags        []string
	remaining   []string
	importance  int
}

// parseMemoryFlags parses common memory command flags from args.
func parseMemoryFlags(args []string) memoryFlags {
	var flags memoryFlags
	flags.sourceType = agent.SourceTypeUserMessage
	flags.importance = 0 // Default 0 means "no filter" for search; write applies default 3 separately

	skip := false
	for i, arg := range args {
		if skip {
			skip = false
			continue
		}
		if handled, shouldSkip := parseMemoryFlag(&flags, args, i, arg); handled {
			skip = shouldSkip
		} else {
			flags.remaining = append(flags.remaining, arg)
		}
	}
	return flags
}

// parseMemoryFlag handles a single flag argument.
// Returns (handled, skipNext).
func parseMemoryFlag(flags *memoryFlags, args []string, i int, arg string) (bool, bool) {
	if i+1 >= len(args) {
		return false, false
	}
	val := args[i+1]
	switch arg {
	case "--source-type":
		flags.sourceType = agent.ParseSourceType(val)
		flags.sourceTypes = parseSourceTypeList(val)
		return true, true
	case "--min-importance", "--importance":
		_, _ = fmt.Sscanf(val, "%d", &flags.importance)
		return true, true
	case "--tags":
		flags.tags = parseTagList(val)
		return true, true
	default:
		return false, false
	}
}

// parseSourceTypeList parses a comma-separated list of source types.
func parseSourceTypeList(s string) []agent.SourceType {
	parts := strings.Split(s, ",")
	result := make([]agent.SourceType, 0, len(parts))
	for _, part := range parts {
		result = append(result, agent.ParseSourceType(strings.TrimSpace(part)))
	}
	return result
}

// parseTagList parses a comma-separated list of tags.
func parseTagList(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}

// generateNoteID creates a unique note ID.
func generateNoteID() string {
	return fmt.Sprintf("note-%d", time.Now().UnixNano())
}

// getEnvOrDefault returns the environment variable value or a default if not set.
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
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

	case "index":
		handleIndexCommand(ctx, parts[1:], uc)
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

// handleIndexCommand handles index subcommands.
func handleIndexCommand(ctx context.Context, args []string, uc *useCases) {
	if len(args) == 0 {
		printIndexUsage()
		return
	}

	subcmd := strings.ToLower(args[0])
	subArgs := args[1:]

	switch subcmd {
	case "changed":
		handleIndexChanged(ctx, subArgs, uc)
	case "diff":
		handleIndexDiff(ctx, subArgs, uc)
	case "scan":
		handleIndexScan(ctx, subArgs, uc)
	default:
		fmt.Printf("Unknown index command: %s\n", subcmd)
		printIndexUsage()
	}
}

// handleIndexChanged handles the index changed subcommand.
func handleIndexChanged(ctx context.Context, args []string, uc *useCases) {
	since := parseSinceTime(args)
	if since.IsZero() {
		return // Error already printed by parseSinceTime
	}

	files, err := uc.indexService.ChangedSince(ctx, since)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	printChangedFiles(files, since)
}

// handleIndexDiff handles the index diff subcommand.
func handleIndexDiff(ctx context.Context, args []string, uc *useCases) {
	if len(args) < 2 {
		fmt.Println("Usage: index diff <from_snapshot_id> <to_snapshot_id>")
		return
	}

	fromID := indexing.SnapshotID(args[0])
	toID := indexing.SnapshotID(args[1])

	diff, err := uc.indexService.DiffSnapshots(ctx, fromID, toID)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	printDiffResult(diff, fromID, toID)
}

// handleIndexScan handles the index scan subcommand.
func handleIndexScan(ctx context.Context, args []string, uc *useCases) {
	paths, ignore := parseIndexScanArgs(args)

	fmt.Printf("🔍 Scanning %d path(s)...\n", len(paths))
	snapshot, err := uc.indexService.Scan(ctx, paths, ignore)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("✅ Scan complete!")
	fmt.Println("------------------------------------------")
	fmt.Printf("Snapshot ID:   %s\n", snapshot.ID)
	fmt.Printf("Files indexed: %d\n", snapshot.FileCount())
	fmt.Printf("Created at:    %s\n", snapshot.CreatedAt.Format(time.RFC3339))
	fmt.Println()
}

// printIndexUsage prints index command usage information.
func printIndexUsage() {
	fmt.Println("Usage: index <scan|changed|diff> [args...]")
	fmt.Println("  index scan [paths...] [-- ignore...]  - Scan directories and create a snapshot")
	fmt.Println("  index changed [since]                 - Show files changed since timestamp/duration")
	fmt.Println("  index diff <from_id> <to_id>          - Compare two snapshots")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  index scan                            - Scan current directory")
	fmt.Println("  index scan ./src ./lib                - Scan specific directories")
	fmt.Println("  index scan . -- .git node_modules     - Scan with custom ignore patterns")
	fmt.Println("  index changed 1h                      - Files changed in last hour")
	fmt.Println("  index changed 2024-01-15T10:00:00Z    - Files changed since timestamp")
	fmt.Println("  index diff snap-123 snap-456          - Compare snapshots")
	fmt.Println()
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
	flags := parseMemoryFlags(args)

	query := strings.Join(flags.remaining, " ")
	if query == "" {
		query = "*" // Search all
	}

	opts := buildSearchOptions(flags)
	notes, err := uc.searchNotes.Execute(ctx, query, 10, opts)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	printMemorySearchResults(notes)
}

// handleMemoryWrite handles the memory write subcommand.
func handleMemoryWrite(ctx context.Context, args []string, uc *useCases) {
	if len(args) < 1 {
		fmt.Println("Usage: memory write [--source-type TYPE] [--importance N] [--tags t1,t2] <text>")
		return
	}

	flags := parseMemoryFlags(args)

	content := strings.Join(flags.remaining, " ")
	if content == "" {
		fmt.Println("❌ Error: content cannot be empty")
		return
	}

	// Apply default importance of 3 for writes when not specified
	importance := flags.importance
	if importance == 0 {
		importance = 3
	}

	note := agent.NewMemoryNote(agent.NoteID(generateNoteID()), flags.sourceType).
		WithRawContent(content).
		WithSummary(content).
		WithImportance(importance)

	if len(flags.tags) > 0 {
		note.WithTags(flags.tags...)
	}

	if err := uc.writeNote.Execute(ctx, note); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("💾 Note saved with ID: %s\n", note.ID)
	}
}

// buildSearchOptions creates MemorySearchOptions from parsed flags.
func buildSearchOptions(flags memoryFlags) *agent.MemorySearchOptions {
	if len(flags.sourceTypes) == 0 && len(flags.tags) == 0 && flags.importance <= 0 {
		return nil
	}
	return &agent.MemorySearchOptions{
		MinImportance: flags.importance,
		SourceTypes:   flags.sourceTypes,
		Tags:          flags.tags,
	}
}

// printMemoryUsage prints memory command usage information.
func printMemoryUsage() {
	fmt.Println("Usage: memory <search|get|write|delete> [args...]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  memory search [options] <query>  - Search memory notes")
	fmt.Println("  memory get <id>                  - Get a specific note")
	fmt.Println("  memory write [options] <text>    - Write a new note")
	fmt.Println("  memory delete <id>               - Delete a note")
	fmt.Println()
	fmt.Println("Search options:")
	fmt.Println("  --source-type TYPE     Filter by source type (comma-separated)")
	fmt.Println("  --min-importance N     Filter by minimum importance (1-5)")
	fmt.Println("  --tags t1,t2           Filter by tags (comma-separated)")
	fmt.Println()
	fmt.Println("Write options:")
	fmt.Println("  --source-type TYPE     Set source type (default: user_message)")
	fmt.Println("  --importance N         Set importance 1-5 (default: 3)")
	fmt.Println("  --tags t1,t2           Set tags (comma-separated)")
	fmt.Println()
	fmt.Println("Source types: decision, experiment, external_source, fact, issue,")
	fmt.Println("              plan_step, preference, requirement, retrospective,")
	fmt.Println("              summary, tool_result, user_message")
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

// parseIndexScanArgs parses arguments for the index scan command.
// Returns paths and ignore patterns.
func parseIndexScanArgs(args []string) ([]string, []string) {
	var paths, ignore []string

	// Parse arguments: paths before --, ignore patterns after --
	inIgnore := false
	for _, arg := range args {
		if arg == "--" {
			inIgnore = true
			continue
		}
		if inIgnore {
			ignore = append(ignore, arg)
		} else {
			paths = append(paths, arg)
		}
	}

	// Default to current directory if no paths specified
	if len(paths) == 0 {
		cwd, _ := os.Getwd()
		paths = []string{cwd}
	}

	// Default ignore patterns
	if len(ignore) == 0 {
		ignore = []string{".git", "node_modules", "vendor", "__pycache__", ".DS_Store"}
	}

	return paths, ignore
}

// parseSinceTime parses a time string for the index changed command.
// Returns zero time if parsing fails (after printing error).
func parseSinceTime(args []string) time.Time {
	if len(args) == 0 {
		// Default to 24 hours ago
		return time.Now().Add(-24 * time.Hour)
	}

	// Try RFC3339 format first
	if parsed, err := time.Parse(time.RFC3339, args[0]); err == nil {
		return parsed
	}

	// Try duration format (e.g., "1h", "24h")
	if duration, err := time.ParseDuration(args[0]); err == nil {
		return time.Now().Add(-duration)
	}

	fmt.Printf("❌ Invalid time format. Use RFC3339 (e.g., 2024-01-15T10:00:00Z) or duration (e.g., 1h, 24h)\n")
	return time.Time{}
}

// printBanner displays the startup banner.
func printBanner(chattingURL, chattingModel, embeddingURL, embeddingModel string, maxIter, maxMsg int, memoryFile, indexFile string, parallelTools bool) {
	appName := getEnvOrDefault("APP_NAME", "Go Agent")
	appDescription := getEnvOrDefault("APP_DESCRIPTION", "AI Agent CLI")
	fmt.Printf("🤖 %s - %s\n", appName, appDescription)
	fmt.Println(strings.Repeat("=", len(appName)+len(appDescription)+6))
	fmt.Printf("Chatting URL:    %s\n", chattingURL)
	fmt.Printf("Chatting Model:  %s\n", chattingModel)
	if embeddingModel != "" {
		fmt.Printf("Embedding URL:   %s\n", embeddingURL)
		fmt.Printf("Embedding Model: %s\n", embeddingModel)
	} else {
		fmt.Println("Embeddings:      disabled")
	}
	fmt.Printf("Max iterations:  %d\n", maxIter)
	fmt.Printf("Max messages:    %d\n", maxMsg)
	if memoryFile != "" {
		fmt.Printf("Memory file:     %s\n", memoryFile)
	} else {
		fmt.Println("Memory:          in-memory (ephemeral)")
	}
	if indexFile != "" {
		fmt.Printf("Index file:      %s\n", indexFile)
	} else {
		fmt.Println("Index:           in-memory (ephemeral)")
	}
	fmt.Printf("Parallel tools:  %v\n", parallelTools)
	fmt.Println()
	fmt.Println("Type 'help' for available commands.")
	fmt.Println()
}

// printChangedFiles displays changed files since a timestamp.
func printChangedFiles(files []indexing.FileInfo, since time.Time) {
	fmt.Println()
	fmt.Printf("📁 Files changed since %s\n", since.Format(time.RFC3339))
	fmt.Println("------------------------------------------")
	if len(files) == 0 {
		fmt.Println("No files changed.")
	} else {
		for _, f := range files {
			fmt.Printf("  %s (%d bytes, %s)\n", f.Path, f.Size, f.ModTime.Format(time.RFC3339))
		}
		fmt.Printf("\nTotal: %d file(s)\n", len(files))
	}
	fmt.Println()
}

// printDiffResult displays the diff between two snapshots.
func printDiffResult(diff indexing.DiffResult, fromID, toID indexing.SnapshotID) {
	fmt.Println()
	fmt.Printf("📊 Diff: %s → %s\n", fromID, toID)
	fmt.Println("------------------------------------------")

	if len(diff.Added) > 0 {
		fmt.Printf("\n✅ Added (%d):\n", len(diff.Added))
		for _, f := range diff.Added {
			fmt.Printf("  + %s\n", f.Path)
		}
	}

	if len(diff.Changed) > 0 {
		fmt.Printf("\n📝 Changed (%d):\n", len(diff.Changed))
		for _, f := range diff.Changed {
			fmt.Printf("  ~ %s\n", f.Path)
		}
	}

	if len(diff.Removed) > 0 {
		fmt.Printf("\n❌ Removed (%d):\n", len(diff.Removed))
		for _, f := range diff.Removed {
			fmt.Printf("  - %s\n", f.Path)
		}
	}

	if len(diff.Added) == 0 && len(diff.Changed) == 0 && len(diff.Removed) == 0 {
		fmt.Println("No differences found.")
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

// printHelp displays available commands.
func printHelp() {
	fmt.Println()
	fmt.Println("📖 Available Commands")
	fmt.Println("---------------------")
	fmt.Println("  clear              Clear conversation history")
	fmt.Println("  help               Show this help message")
	fmt.Println("  index <subcmd>     Index operations (scan, changed, diff)")
	fmt.Println("  memory <subcmd>    Memory operations (search, get, write, delete)")
	fmt.Println("  quit / exit        Exit the CLI")
	fmt.Println("  stats              Show agent statistics")
	fmt.Println()
	fmt.Println("💡 Tips:")
	fmt.Println("  - Ask the agent to calculate: 'What is 42 * 17?'")
	fmt.Println("  - Ask for the time: 'What time is it?'")
	fmt.Println("  - Save to memory: 'Remember that my favorite color is blue'")
	fmt.Println("  - Recall memory: 'What is my favorite color?'")
	fmt.Println("  - Scan files: 'index scan ./src' or ask 'Scan my project directory'")
	fmt.Println("  - Find changes: 'index changed 1h' or ask 'What files changed today?'")
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
	if len(note.Embedding) > 0 {
		fmt.Printf("Embedding:   [%d dimensions]\n", len(note.Embedding))
	} else {
		fmt.Printf("Embedding:   (none)\n")
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
func setupInfrastructure(baseURL, model, memoryFile, indexFile string, verbose, parallelTools bool, embeddingURL, embeddingModel string) *infrastructure {
	logger := createLogger(verbose)
	dispatcher := messaging.NewExternalDispatcher()
	publisher := outbound.NewEventPublisher(dispatcher)
	memoryStore := createMemoryStore(memoryFile)
	memoryToolSvc := tooling.NewMemoryToolService(memoryStore, generateNoteID)

	// Configure embedding client if model is specified
	if embeddingModel != "" {
		embeddingClient := outbound.NewOpenAIEmbeddingClient(embeddingURL).
			WithModel(embeddingModel)
		if logger != nil {
			embeddingClient.WithLogger(logger)
		}
		memoryToolSvc.WithEmbedder(embeddingClient)
	}

	// Create indexing infrastructure
	indexStore := createIndexStore(indexFile)
	fileWalker := inbound.NewFSWalker()
	indexService := indexing.NewService(fileWalker, indexStore, generateSnapshotID)
	indexToolSvc := tooling.NewIndexToolService(indexService)

	toolExecutor := createToolExecutor(verbose, logger, memoryToolSvc, indexToolSvc)
	llmClient := createLLMClient(baseURL, model, verbose, logger)
	hooks := createHooks(verbose)
	taskService := createTaskService(llmClient, toolExecutor, publisher, hooks, parallelTools)

	return &infrastructure{
		dispatcher:    dispatcher,
		indexService:  indexService,
		indexToolSvc:  indexToolSvc,
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

// createIndexStore creates either a file-backed or in-memory index store.
func createIndexStore(indexFile string) *outbound.IndexStore {
	if indexFile != "" {
		return outbound.NewIndexStore(indexFile)
	}
	return outbound.NewInMemoryIndexStore()
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
func createToolExecutor(verbose bool, logger *slog.Logger, memoryToolSvc *tooling.MemoryToolService, indexToolSvc *tooling.IndexToolService) *outbound.ToolExecutor {
	executor := outbound.NewToolExecutor()
	if verbose && logger != nil {
		executor = executor.WithLogger(logger)
	}
	registerTools(executor, memoryToolSvc, indexToolSvc)
	return executor
}

// generateSnapshotID creates a unique snapshot ID.
func generateSnapshotID() string {
	return fmt.Sprintf("snap-%d", time.Now().UnixNano())
}

// registerTools registers all available tools with the executor.
func registerTools(executor *outbound.ToolExecutor, memoryToolSvc *tooling.MemoryToolService, indexToolSvc *tooling.IndexToolService) {
	// Register index.changed_since tool
	indexChangedSinceTool := tooling.NewIndexChangedSinceTool(indexToolSvc)
	executor.RegisterTool(string(indexChangedSinceTool.ID), indexChangedSinceTool.Func)
	executor.RegisterToolDefinition(indexChangedSinceTool.Definition)

	// Register index.diff_snapshot tool
	indexDiffSnapshotTool := tooling.NewIndexDiffSnapshotTool(indexToolSvc)
	executor.RegisterTool(string(indexDiffSnapshotTool.ID), indexDiffSnapshotTool.Func)
	executor.RegisterToolDefinition(indexDiffSnapshotTool.Definition)

	// Register index.scan tool
	indexScanTool := tooling.NewIndexScanTool(indexToolSvc)
	executor.RegisterTool(string(indexScanTool.ID), indexScanTool.Func)
	executor.RegisterToolDefinition(indexScanTool.Definition)

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
