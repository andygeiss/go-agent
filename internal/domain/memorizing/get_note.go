package memorizing

import (
	"context"

	"github.com/andygeiss/go-agent/internal/domain/agent"
)

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
