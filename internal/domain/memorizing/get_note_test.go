package memorizing_test

import (
	"context"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/memorizing"
)

func Test_GetNoteUseCase_Execute_Should_ReturnNote(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	originalNote := agent.NewMemoryNote("note-123", agent.SourceTypePreference)
	store.notes["note-123"] = originalNote
	uc := memorizing.NewGetNoteUseCase(store)

	// Act
	note, err := uc.Execute(context.Background(), "note-123")

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "note must match", note, originalNote)
}

func Test_GetNoteUseCase_Execute_WithEmptyID_Should_ReturnError(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	uc := memorizing.NewGetNoteUseCase(store)

	// Act
	_, err := uc.Execute(context.Background(), "")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}
