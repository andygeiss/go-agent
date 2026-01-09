package memorizing_test

import (
	"context"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/memorizing"
)

func Test_DeleteNoteUseCase_Execute_Should_RemoveNote(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	store.notes["note-123"] = agent.NewMemoryNote("note-123", agent.SourceTypePreference)
	uc := memorizing.NewDeleteNoteUseCase(store)

	// Act
	err := uc.Execute(context.Background(), "note-123")

	// Assert
	assert.That(t, "error must be nil", err, nil)
	_, exists := store.notes["note-123"]
	assert.That(t, "note must be deleted from store", exists, false)
}

func Test_DeleteNoteUseCase_Execute_WithEmptyID_Should_ReturnError(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	uc := memorizing.NewDeleteNoteUseCase(store)

	// Act
	err := uc.Execute(context.Background(), "")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}
