package memorizing_test

import (
	"context"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/memorizing"
)

func Test_WriteNoteUseCase_Execute_Should_StoreNote(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	uc := memorizing.NewWriteNoteUseCase(store)
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference).
		WithRawContent("User prefers German")

	// Act
	err := uc.Execute(context.Background(), note)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "note must be stored", store.notes["note-123"], note)
}

func Test_WriteNoteUseCase_Execute_WithNilNote_Should_ReturnError(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	uc := memorizing.NewWriteNoteUseCase(store)

	// Act
	err := uc.Execute(context.Background(), nil)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_WriteNoteUseCase_Execute_WithEmptyID_Should_ReturnError(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	uc := memorizing.NewWriteNoteUseCase(store)
	note := agent.NewMemoryNote("", agent.SourceTypePreference)

	// Act
	err := uc.Execute(context.Background(), note)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}
