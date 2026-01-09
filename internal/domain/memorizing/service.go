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
