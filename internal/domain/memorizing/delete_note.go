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
