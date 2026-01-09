package memorizing

import (
	"context"

	"github.com/andygeiss/go-agent/internal/domain/agent"
)

// Service provides memory management use cases.
// It coordinates between memory tools and the storage backend.
type Service struct {
	store agent.MemoryStore
}

// NewService creates a new memory service with the given store.
func NewService(store agent.MemoryStore) *Service {
	return &Service{store: store}
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

// SearchNotes retrieves notes matching the query with optional filters.
// Returns up to `limit` notes sorted by relevance.
func (s *Service) SearchNotes(ctx context.Context, query string, limit int, opts *agent.MemorySearchOptions) ([]*agent.MemoryNote, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}
	return s.store.Search(ctx, query, limit, opts)
}

// GetNote retrieves a specific note by ID.
// Returns nil if the note is not found.
func (s *Service) GetNote(ctx context.Context, id agent.NoteID) (*agent.MemoryNote, error) {
	if id == "" {
		return nil, ErrNoteIDEmpty
	}
	return s.store.Get(ctx, id)
}

// DeleteNote removes a note by ID.
func (s *Service) DeleteNote(ctx context.Context, id agent.NoteID) error {
	if id == "" {
		return ErrNoteIDEmpty
	}
	return s.store.Delete(ctx, id)
}
