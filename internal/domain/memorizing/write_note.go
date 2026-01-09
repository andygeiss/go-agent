package memorizing

import (
	"context"

	"github.com/andygeiss/go-agent/internal/domain/agent"
)

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
