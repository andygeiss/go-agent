package tooling

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/andygeiss/go-agent/internal/domain/agent"
)

// memoryWriteArgs represents the arguments for the memory_write tool.
type memoryWriteArgs struct {
	ContextDescription string   `json:"context_description"`
	RawContent         string   `json:"raw_content"`
	SessionID          string   `json:"session_id,omitempty"`
	SourceType         string   `json:"source_type"`
	Summary            string   `json:"summary"`
	TaskID             string   `json:"task_id,omitempty"`
	UserID             string   `json:"user_id,omitempty"`
	Keywords           []string `json:"keywords"`
	Tags               []string `json:"tags"`
	Importance         int      `json:"importance"`
}

// memorySearchArgs represents the arguments for the memory_search tool.
type memorySearchArgs struct {
	Query         string   `json:"query"`
	SessionID     string   `json:"session_id,omitempty"`
	SourceTypes   []string `json:"source_types,omitempty"`
	TaskID        string   `json:"task_id,omitempty"`
	UserID        string   `json:"user_id,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Limit         int      `json:"limit,omitempty"`
	MinImportance int      `json:"min_importance,omitempty"`
}

// memoryGetArgs represents the arguments for the memory_get tool.
type memoryGetArgs struct {
	ID string `json:"id"`
}

// memorySearchResult represents a single search result for JSON output.
type memorySearchResult struct {
	ContextDescription string   `json:"context_description"`
	ID                 string   `json:"id"`
	SourceType         string   `json:"source_type"`
	Summary            string   `json:"summary"`
	Tags               []string `json:"tags"`
	Importance         int      `json:"importance"`
}

// MemoryToolService provides memory tool implementations.
// It requires a MemoryStore to be injected for actual storage.
type MemoryToolService struct {
	embedder agent.EmbeddingClient
	idGen    func() string
	session  string
	store    agent.MemoryStore
	userID   string
}

// NewMemoryToolService creates a new memory tool service.
func NewMemoryToolService(store agent.MemoryStore, idGenerator func() string) *MemoryToolService {
	return &MemoryToolService{
		idGen: idGenerator,
		store: store,
	}
}

// MemoryGet retrieves a specific note by ID.
func (s *MemoryToolService) MemoryGet(ctx context.Context, arguments string) (string, error) {
	var args memoryGetArgs
	if err := agent.DecodeArgs(arguments, &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	note, err := s.store.Get(ctx, agent.NoteID(args.ID))
	if err != nil {
		return "", fmt.Errorf("failed to get memory note: %w", err)
	}

	output, err := json.Marshal(map[string]any{
		"status": "success",
		"note": map[string]any{
			"id":                  string(note.ID),
			"source_type":         string(note.SourceType),
			"raw_content":         note.RawContent,
			"summary":             note.Summary,
			"context_description": note.ContextDescription,
			"keywords":            note.Keywords,
			"tags":                note.Tags,
			"importance":          note.Importance,
			"user_id":             note.UserID,
			"session_id":          note.SessionID,
			"task_id":             note.TaskID,
			"created_at":          note.CreatedAt,
			"updated_at":          note.UpdatedAt,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal note: %w", err)
	}

	return string(output), nil
}

// MemorySearch retrieves notes matching the query and filters.
func (s *MemoryToolService) MemorySearch(ctx context.Context, arguments string) (string, error) {
	var args memorySearchArgs
	if err := agent.DecodeArgs(arguments, &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	opts := buildMemorySearchOpts(args)
	limit := defaultLimit(args.Limit, 10)

	notes, err := s.store.Search(ctx, args.Query, limit, opts)
	if err != nil {
		return "", fmt.Errorf("failed to search memory: %w", err)
	}

	return marshalSearchResults(notes)
}

// buildMemorySearchOpts creates MemorySearchOptions from args if any filters are set.
func buildMemorySearchOpts(args memorySearchArgs) *agent.MemorySearchOptions {
	if args.UserID == "" && args.SessionID == "" && args.TaskID == "" && len(args.Tags) == 0 && len(args.SourceTypes) == 0 && args.MinImportance == 0 {
		return nil
	}
	return &agent.MemorySearchOptions{
		MinImportance: args.MinImportance,
		SessionID:     args.SessionID,
		SourceTypes:   mapSourceTypes(args.SourceTypes),
		TaskID:        args.TaskID,
		UserID:        args.UserID,
		Tags:          args.Tags,
	}
}

// defaultLimit returns the limit or a default if limit is <= 0.
func defaultLimit(limit, defaultVal int) int {
	if limit <= 0 {
		return defaultVal
	}
	return limit
}

// marshalSearchResults converts notes to JSON search results.
func marshalSearchResults(notes []*agent.MemoryNote) (string, error) {
	results := make([]memorySearchResult, len(notes))
	for i, note := range notes {
		results[i] = memorySearchResult{
			Tags:               note.Tags,
			ContextDescription: note.ContextDescription,
			ID:                 string(note.ID),
			SourceType:         string(note.SourceType),
			Summary:            note.Summary,
			Importance:         note.Importance,
		}
	}

	output, err := json.Marshal(map[string]any{
		"status":  "success",
		"count":   len(results),
		"results": results,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}

	return string(output), nil
}

// MemoryWrite stores a new memory note.
func (s *MemoryToolService) MemoryWrite(ctx context.Context, arguments string) (string, error) {
	var args memoryWriteArgs
	if err := agent.DecodeArgs(arguments, &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	note := s.buildNote(args)
	s.applyEmbedding(ctx, note)

	if err := s.store.Write(ctx, note); err != nil {
		return "", fmt.Errorf("failed to write memory note: %w", err)
	}

	return fmt.Sprintf(`{"status": "success", "note_id": "%s"}`, note.ID), nil
}

// WithEmbedder sets the embedding client for generating note embeddings.
// If set, embeddings will be generated automatically when writing notes.
func (s *MemoryToolService) WithEmbedder(embedder agent.EmbeddingClient) *MemoryToolService {
	s.embedder = embedder
	return s
}

// WithSessionID sets the default session ID for notes.
func (s *MemoryToolService) WithSessionID(sessionID string) *MemoryToolService {
	s.session = sessionID
	return s
}

// WithUserID sets the default user ID for notes.
func (s *MemoryToolService) WithUserID(userID string) *MemoryToolService {
	s.userID = userID
	return s
}

// applyEmbedding generates and attaches an embedding to the note if configured.
func (s *MemoryToolService) applyEmbedding(ctx context.Context, note *agent.MemoryNote) {
	if s.embedder == nil {
		return
	}
	embedding, err := s.embedder.Embed(ctx, note.SearchableText())
	if err == nil && len(embedding) > 0 {
		note.WithEmbedding(embedding)
	}
	// Silently skip embedding on error - note is still useful without it
}

// applyScopeIDs sets user, session, and task IDs on the note.
func (s *MemoryToolService) applyScopeIDs(note *agent.MemoryNote, args memoryWriteArgs) {
	if args.UserID != "" {
		note.WithUserID(args.UserID)
	} else if s.userID != "" {
		note.WithUserID(s.userID)
	}

	if args.SessionID != "" {
		note.WithSessionID(args.SessionID)
	} else if s.session != "" {
		note.WithSessionID(s.session)
	}

	if args.TaskID != "" {
		note.WithTaskID(args.TaskID)
	}
}

// buildNote creates a MemoryNote from write arguments.
func (s *MemoryToolService) buildNote(args memoryWriteArgs) *agent.MemoryNote {
	noteID := agent.NoteID(s.idGen())
	sourceType := mapSourceType(args.SourceType)

	note := agent.NewMemoryNote(noteID, sourceType).
		WithRawContent(args.RawContent).
		WithSummary(args.Summary).
		WithContextDescription(args.ContextDescription).
		WithKeywords(args.Keywords...).
		WithTags(args.Tags...).
		WithImportance(args.Importance)

	s.applyScopeIDs(note, args)
	return note
}

// mapSourceType maps a string to a SourceType.
func mapSourceType(s string) agent.SourceType {
	return agent.ParseSourceType(s)
}

// mapSourceTypes converts a slice of strings to SourceTypes.
func mapSourceTypes(strs []string) []agent.SourceType {
	if len(strs) == 0 {
		return nil
	}
	result := make([]agent.SourceType, len(strs))
	for i, s := range strs {
		result[i] = agent.ParseSourceType(s)
	}
	return result
}

// NewMemoryGetTool creates the memory_get tool definition.
func NewMemoryGetTool(svc *MemoryToolService) agent.Tool {
	return agent.Tool{
		ID: "memory_get",
		Definition: agent.NewToolDefinition("memory_get", "Retrieve full details of a specific memory note by ID. Use after memory_search returns relevant IDs.").
			WithParameterDef(agent.NewParameterDefinition("id", agent.ParamTypeString).
				WithDescription("The unique identifier of the note to retrieve").
				WithRequired()),
		Func: svc.MemoryGet,
	}
}

// NewMemorySearchTool creates the memory_search tool definition.
func NewMemorySearchTool(svc *MemoryToolService) agent.Tool {
	return agent.Tool{
		ID: "memory_search",
		Definition: agent.NewToolDefinition("memory_search", "Search long-term memory for relevant notes. Use when user refers to past interactions or you need prior preferences/results.").
			WithParameterDef(agent.NewParameterDefinition("query", agent.ParamTypeString).
				WithDescription("Natural language query describing what you want to remember").
				WithRequired()).
			WithParameterDef(agent.NewParameterDefinition("limit", agent.ParamTypeInteger).
				WithDescription("Maximum number of notes to return (default: 10)").
				WithDefault("10")).
			WithParameterDef(agent.NewParameterDefinition("source_types", agent.ParamTypeArray).
				WithDescription("Filter by source types: decision, experiment, external_source, fact, issue, plan_step, preference, requirement, retrospective, summary, tool_result, user_message")).
			WithParameterDef(agent.NewParameterDefinition("min_importance", agent.ParamTypeInteger).
				WithDescription("Filter by minimum importance (1-5)")).
			WithParameterDef(agent.NewParameterDefinition("user_id", agent.ParamTypeString).
				WithDescription("Filter by user ID")).
			WithParameterDef(agent.NewParameterDefinition("session_id", agent.ParamTypeString).
				WithDescription("Filter by session ID")).
			WithParameterDef(agent.NewParameterDefinition("tags", agent.ParamTypeArray).
				WithDescription("Filter by tags (any match)")),
		Func: svc.MemorySearch,
	}
}

// NewMemoryWriteTool creates the memory_write tool definition.
func NewMemoryWriteTool(svc *MemoryToolService) agent.Tool {
	return agent.Tool{
		ID: "memory_write",
		Definition: agent.NewToolDefinition("memory_write", "Store a new memory note for long-term recall. Use this to save preferences, important facts, decisions, requirements, or summaries.").
			WithParameterDef(agent.NewParameterDefinition("source_type", agent.ParamTypeString).
				WithDescription("What category this note belongs to: decision (architectural choices), experiment (hypotheses/results), external_source (URLs/references), fact (verified information), issue (problems/bugs), plan_step (task steps), preference (user preferences), requirement (must-haves), retrospective (lessons learned), summary (condensed info), tool_result (tool output), user_message (user input)").
				WithEnum("decision", "experiment", "external_source", "fact", "issue", "plan_step", "preference", "requirement", "retrospective", "summary", "tool_result", "user_message").
				WithRequired()).
			WithParameterDef(agent.NewParameterDefinition("raw_content", agent.ParamTypeString).
				WithDescription("The core text or compact representation of the source").
				WithRequired()).
			WithParameterDef(agent.NewParameterDefinition("summary", agent.ParamTypeString).
				WithDescription("1-3 sentence summary understandable out of context").
				WithRequired()).
			WithParameterDef(agent.NewParameterDefinition("context_description", agent.ParamTypeString).
				WithDescription("Explains why this note might matter in the future")).
			WithParameterDef(agent.NewParameterDefinition("keywords", agent.ParamTypeArray).
				WithDescription("3-8 keywords for searchability")).
			WithParameterDef(agent.NewParameterDefinition("tags", agent.ParamTypeArray).
				WithDescription("Tags like: preference, config, api_result, task_summary, bug, codebase_fact")).
			WithParameterDef(agent.NewParameterDefinition("importance", agent.ParamTypeInteger).
				WithDescription("1-5 importance score: 1=minor session info, 5=critical preference").
				WithDefault("2")),
		Func: svc.MemoryWrite,
	}
}
