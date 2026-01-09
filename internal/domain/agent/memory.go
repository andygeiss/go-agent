package agent

import "context"

// MemorySearchOptions configures the search behavior.
type MemorySearchOptions struct {
	UserID    string   // Filter by user ID
	SessionID string   // Filter by session ID
	TaskID    string   // Filter by task ID
	Tags      []string // Filter by tags (any match)
}

// MemoryStore is the interface for persisting and retrieving memory notes.
// Implementations can use in-memory, JSON file, or database storage with embeddings.
type MemoryStore interface {
	// Write stores a new memory note.
	Write(ctx context.Context, note *MemoryNote) error
	// Search retrieves notes matching the query and filters.
	Search(ctx context.Context, query string, limit int, opts *MemorySearchOptions) ([]*MemoryNote, error)
	// Get retrieves a specific note by ID.
	Get(ctx context.Context, id NoteID) (*MemoryNote, error)
	// Delete removes a note by ID.
	Delete(ctx context.Context, id NoteID) error
}
