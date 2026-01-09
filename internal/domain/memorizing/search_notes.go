package memorizing

import (
	"context"

	"github.com/andygeiss/go-agent/internal/domain/agent"
)

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
