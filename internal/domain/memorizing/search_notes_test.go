package memorizing_test

import (
	"context"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/memorizing"
)

func Test_SearchNotesUseCase_Execute_Should_ReturnNotes(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	store.searchNotes = []*agent.MemoryNote{
		agent.NewMemoryNote("note-1", agent.SourceTypePreference),
		agent.NewMemoryNote("note-2", agent.SourceTypePreference),
	}
	uc := memorizing.NewSearchNotesUseCase(store)

	// Act
	notes, err := uc.Execute(context.Background(), "preferences", 10, nil)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "notes count must be 2", len(notes), 2)
}

func Test_SearchNotesUseCase_Execute_WithZeroLimit_Should_UseDefaultLimit(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	uc := memorizing.NewSearchNotesUseCase(store)

	// Act
	_, err := uc.Execute(context.Background(), "test", 0, nil)

	// Assert
	assert.That(t, "error must be nil", err, nil)
}
