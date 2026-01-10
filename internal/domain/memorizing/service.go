package memorizing

import (
	"context"

	"github.com/andygeiss/go-agent/internal/domain/agent"
)

// DeleteNoteUseCase handles removing memory notes.
type DeleteNoteUseCase struct {
	store agent.MemoryStore
}

// NewDeleteNoteUseCase creates a new DeleteNoteUseCase with the given store.
func NewDeleteNoteUseCase(store agent.MemoryStore) *DeleteNoteUseCase {
	return &DeleteNoteUseCase{store: store}
}

// Execute removes a note by ID.
func (uc *DeleteNoteUseCase) Execute(ctx context.Context, id agent.NoteID) error {
	if id == "" {
		return ErrNoteIDEmpty
	}
	return uc.store.Delete(ctx, id)
}

// GetNoteUseCase handles retrieving a specific memory note.
type GetNoteUseCase struct {
	store agent.MemoryStore
}

// NewGetNoteUseCase creates a new GetNoteUseCase with the given store.
func NewGetNoteUseCase(store agent.MemoryStore) *GetNoteUseCase {
	return &GetNoteUseCase{store: store}
}

// Execute retrieves a specific note by ID.
// Returns nil if the note is not found.
func (uc *GetNoteUseCase) Execute(ctx context.Context, id agent.NoteID) (*agent.MemoryNote, error) {
	if id == "" {
		return nil, ErrNoteIDEmpty
	}
	return uc.store.Get(ctx, id)
}

// SearchNotesUseCase handles searching for memory notes.
type SearchNotesUseCase struct {
	store agent.MemoryStore
}

// NewSearchNotesUseCase creates a new SearchNotesUseCase with the given store.
func NewSearchNotesUseCase(store agent.MemoryStore) *SearchNotesUseCase {
	return &SearchNotesUseCase{store: store}
}

// Execute retrieves notes matching the query with optional filters.
// Returns up to `limit` notes sorted by relevance.
func (uc *SearchNotesUseCase) Execute(ctx context.Context, query string, limit int, opts *agent.MemorySearchOptions) ([]*agent.MemoryNote, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}
	return uc.store.Search(ctx, query, limit, opts)
}

// Service provides memory management use cases.
// It coordinates between memory tools and the storage backend.
type Service struct {
	store agent.MemoryStore
}

// NewService creates a new memory service with the given store.
func NewService(store agent.MemoryStore) *Service {
	return &Service{store: store}
}

// DeleteNote removes a note by ID.
func (s *Service) DeleteNote(ctx context.Context, id agent.NoteID) error {
	if id == "" {
		return ErrNoteIDEmpty
	}
	return s.store.Delete(ctx, id)
}

// GetNote retrieves a specific note by ID.
// Returns nil if the note is not found.
func (s *Service) GetNote(ctx context.Context, id agent.NoteID) (*agent.MemoryNote, error) {
	if id == "" {
		return nil, ErrNoteIDEmpty
	}
	return s.store.Get(ctx, id)
}

// SearchNotes retrieves notes matching the query with optional filters.
// Returns up to `limit` notes sorted by relevance.
func (s *Service) SearchNotes(ctx context.Context, query string, limit int, opts *agent.MemorySearchOptions) ([]*agent.MemoryNote, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}
	return s.store.Search(ctx, query, limit, opts)
}

// WriteNote stores a new memory note.
// Returns an error if the note cannot be stored.
func (s *Service) WriteNote(ctx context.Context, note *agent.MemoryNote) error {
	if note == nil {
		return ErrNoteNil
	}
	if note.ID == "" {
		return ErrNoteIDEmpty
	}
	return s.store.Write(ctx, note)
}

// WriteTypedNote stores a new memory note with the specified source type.
// This is a convenience method that uses the appropriate helper constructor.
func (s *Service) WriteTypedNote(ctx context.Context, id agent.NoteID, sourceType agent.SourceType, content string, opts *TypedNoteOptions) error {
	if id == "" {
		return ErrNoteIDEmpty
	}

	note := agent.NewMemoryNote(id, sourceType).
		WithRawContent(content).
		WithSummary(content)

	applyTypedNoteOptions(note, opts)

	return s.store.Write(ctx, note)
}

// applyTypedNoteOptions applies optional parameters to a note.
func applyTypedNoteOptions(note *agent.MemoryNote, opts *TypedNoteOptions) {
	if opts == nil {
		return
	}
	if len(opts.Tags) > 0 {
		note.WithTags(opts.Tags...)
	}
	if opts.Importance > 0 {
		note.WithImportance(opts.Importance)
	}
	if opts.ContextDescription != "" {
		note.WithContextDescription(opts.ContextDescription)
	}
	if len(opts.Keywords) > 0 {
		note.WithKeywords(opts.Keywords...)
	}
	if opts.UserID != "" {
		note.WithUserID(opts.UserID)
	}
	if opts.SessionID != "" {
		note.WithSessionID(opts.SessionID)
	}
	if opts.TaskID != "" {
		note.WithTaskID(opts.TaskID)
	}
}

// SearchBySourceTypes retrieves notes of specific source types.
func (s *Service) SearchBySourceTypes(ctx context.Context, query string, sourceTypes []agent.SourceType, limit int) ([]*agent.MemoryNote, error) {
	opts := &agent.MemorySearchOptions{
		SourceTypes: sourceTypes,
	}
	return s.store.Search(ctx, query, limit, opts)
}

// SearchDecisions retrieves decision notes matching the query.
func (s *Service) SearchDecisions(ctx context.Context, query string, limit int) ([]*agent.MemoryNote, error) {
	return s.SearchBySourceTypes(ctx, query, []agent.SourceType{agent.SourceTypeDecision}, limit)
}

// SearchFacts retrieves fact notes matching the query.
func (s *Service) SearchFacts(ctx context.Context, query string, limit int) ([]*agent.MemoryNote, error) {
	return s.SearchBySourceTypes(ctx, query, []agent.SourceType{agent.SourceTypeFact}, limit)
}

// SearchPlanSteps retrieves plan step notes matching the query.
func (s *Service) SearchPlanSteps(ctx context.Context, query string, limit int) ([]*agent.MemoryNote, error) {
	return s.SearchBySourceTypes(ctx, query, []agent.SourceType{agent.SourceTypePlanStep}, limit)
}

// SearchPreferences retrieves preference notes matching the query.
func (s *Service) SearchPreferences(ctx context.Context, query string, limit int) ([]*agent.MemoryNote, error) {
	return s.SearchBySourceTypes(ctx, query, []agent.SourceType{agent.SourceTypePreference}, limit)
}

// SearchRequirements retrieves requirement notes matching the query.
func (s *Service) SearchRequirements(ctx context.Context, query string, limit int) ([]*agent.MemoryNote, error) {
	return s.SearchBySourceTypes(ctx, query, []agent.SourceType{agent.SourceTypeRequirement}, limit)
}

// SearchSummaries retrieves summary notes matching the query.
func (s *Service) SearchSummaries(ctx context.Context, query string, limit int) ([]*agent.MemoryNote, error) {
	return s.SearchBySourceTypes(ctx, query, []agent.SourceType{agent.SourceTypeSummary}, limit)
}

// TypedNoteOptions provides optional parameters for WriteTypedNote.
type TypedNoteOptions struct {
	ContextDescription string
	SessionID          string
	TaskID             string
	UserID             string
	Keywords           []string
	Tags               []string
	Importance         int
}

// WriteNoteUseCase handles storing memory notes.
type WriteNoteUseCase struct {
	store agent.MemoryStore
}

// NewWriteNoteUseCase creates a new WriteNoteUseCase with the given store.
func NewWriteNoteUseCase(store agent.MemoryStore) *WriteNoteUseCase {
	return &WriteNoteUseCase{store: store}
}

// Execute stores a new memory note.
// Returns an error if the note cannot be stored.
func (uc *WriteNoteUseCase) Execute(ctx context.Context, note *agent.MemoryNote) error {
	if note == nil {
		return ErrNoteNil
	}
	if note.ID == "" {
		return ErrNoteIDEmpty
	}
	return uc.store.Write(ctx, note)
}
