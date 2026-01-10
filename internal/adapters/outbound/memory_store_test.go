package outbound_test

import (
	"context"
	"errors"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/internal/domain/agent"
)

func Test_MemoryStore_Write_Should_StoreNote(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference).
		WithRawContent("User prefers German")

	// Act
	err := store.Write(context.Background(), note)

	// Assert
	assert.That(t, "error must be nil", err, nil)

	// Verify note was stored
	retrieved, _ := store.Get(context.Background(), "note-123")
	assert.That(t, "note must be retrievable", retrieved != nil, true)
	assert.That(t, "raw content must match", retrieved.RawContent, "User prefers German")
}

func Test_MemoryStore_Write_Should_UpdateExistingNote(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note1 := agent.NewMemoryNote("note-123", agent.SourceTypePreference).
		WithRawContent("Original content")
	_ = store.Write(context.Background(), note1)

	note2 := agent.NewMemoryNote("note-123", agent.SourceTypePreference).
		WithRawContent("Updated content")

	// Act
	err := store.Write(context.Background(), note2)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	retrieved, _ := store.Get(context.Background(), "note-123")
	assert.That(t, "content must be updated", retrieved.RawContent, "Updated content")
}

func Test_MemoryStore_Get_Should_ReturnNote(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference).
		WithRawContent("Test content").
		WithSummary("Test summary")
	_ = store.Write(context.Background(), note)

	// Act
	retrieved, err := store.Get(context.Background(), "note-123")

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "note must not be nil", retrieved != nil, true)
	assert.That(t, "raw content must match", retrieved.RawContent, "Test content")
	assert.That(t, "summary must match", retrieved.Summary, "Test summary")
}

func Test_MemoryStore_Get_WithNonexistent_Should_ReturnError(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()

	// Act
	retrieved, err := store.Get(context.Background(), "nonexistent")

	// Assert
	assert.That(t, "error must be ErrMemoryNoteNotFound", errors.Is(err, outbound.ErrMemoryNoteNotFound), true)
	assert.That(t, "note must be nil", retrieved == nil, true)
}

func Test_MemoryStore_Delete_Should_RemoveNote(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference)
	_ = store.Write(context.Background(), note)

	// Act
	err := store.Delete(context.Background(), "note-123")

	// Assert
	assert.That(t, "error must be nil", err, nil)
	_, getErr := store.Get(context.Background(), "note-123")
	assert.That(t, "get must return ErrMemoryNoteNotFound after delete", errors.Is(getErr, outbound.ErrMemoryNoteNotFound), true)
}

func Test_MemoryStore_Delete_WithNonexistent_Should_NotError(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()

	// Act
	err := store.Delete(context.Background(), "nonexistent")

	// Assert
	assert.That(t, "error must be nil for nonexistent", err, nil)
}

func Test_MemoryStore_Search_Should_FindByTextContent(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note1 := agent.NewMemoryNote("note-1", agent.SourceTypePreference).
		WithRawContent("User prefers German language")
	note2 := agent.NewMemoryNote("note-2", agent.SourceTypeToolResult).
		WithRawContent("API returned success")
	_ = store.Write(context.Background(), note1)
	_ = store.Write(context.Background(), note2)

	// Act
	results, err := store.Search(context.Background(), "German", 10, nil)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should find 1 result", len(results), 1)
	assert.That(t, "found note id must match", results[0].ID, agent.NoteID("note-1"))
}

func Test_MemoryStore_Search_Should_BeCaseInsensitive(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note := agent.NewMemoryNote("note-1", agent.SourceTypePreference).
		WithRawContent("User prefers GERMAN language")
	_ = store.Write(context.Background(), note)

	// Act
	results, err := store.Search(context.Background(), "german", 10, nil)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should find 1 result with case-insensitive search", len(results), 1)
}

func Test_MemoryStore_Search_Should_FilterByUserID(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note1 := agent.NewMemoryNote("note-1", agent.SourceTypePreference).
		WithUserID("user-1").WithRawContent("preference for user 1")
	note2 := agent.NewMemoryNote("note-2", agent.SourceTypePreference).
		WithUserID("user-2").WithRawContent("preference for user 2")
	_ = store.Write(context.Background(), note1)
	_ = store.Write(context.Background(), note2)

	// Act
	opts := &agent.MemorySearchOptions{UserID: "user-1"}
	results, err := store.Search(context.Background(), "preference", 10, opts)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should find 1 result for user-1", len(results), 1)
	assert.That(t, "result must be for user-1", results[0].UserID, "user-1")
}

func Test_MemoryStore_Search_Should_FilterByTags(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note1 := agent.NewMemoryNote("note-1", agent.SourceTypePreference).
		WithRawContent("content").WithTags("language", "formatting")
	note2 := agent.NewMemoryNote("note-2", agent.SourceTypeToolResult).
		WithRawContent("content").WithTags("api", "tool")
	_ = store.Write(context.Background(), note1)
	_ = store.Write(context.Background(), note2)

	// Act
	opts := &agent.MemorySearchOptions{Tags: []string{"language"}}
	results, err := store.Search(context.Background(), "content", 10, opts)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should find 1 result with tag", len(results), 1)
	assert.That(t, "result must have language tag", results[0].HasTag("language"), true)
}

func Test_MemoryStore_Search_Should_MatchKeywords(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note := agent.NewMemoryNote("note-1", agent.SourceTypePreference).
		WithRawContent("some content").
		WithKeywords("german", "language", "preference")
	_ = store.Write(context.Background(), note)

	// Act
	results, err := store.Search(context.Background(), "language", 10, nil)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should find by keyword match", len(results), 1)
}

func Test_MemoryStore_Search_Should_SortByImportance(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note1 := agent.NewMemoryNote("note-1", agent.SourceTypePreference).
		WithRawContent("preference").WithImportance(2)
	note2 := agent.NewMemoryNote("note-2", agent.SourceTypePreference).
		WithRawContent("preference").WithImportance(5)
	note3 := agent.NewMemoryNote("note-3", agent.SourceTypePreference).
		WithRawContent("preference").WithImportance(3)
	_ = store.Write(context.Background(), note1)
	_ = store.Write(context.Background(), note2)
	_ = store.Write(context.Background(), note3)

	// Act
	results, err := store.Search(context.Background(), "preference", 10, nil)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should find 3 results", len(results), 3)
	assert.That(t, "first result must have highest importance", results[0].Importance, 5)
	assert.That(t, "second result must have second highest importance", results[1].Importance, 3)
	assert.That(t, "third result must have lowest importance", results[2].Importance, 2)
}

func Test_MemoryStore_Search_Should_RespectLimit(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	for i := range 5 {
		note := agent.NewMemoryNote(agent.NoteID(string(rune('a'+i))), agent.SourceTypePreference).
			WithRawContent("preference")
		_ = store.Write(context.Background(), note)
	}

	// Act
	results, err := store.Search(context.Background(), "preference", 2, nil)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should respect limit", len(results), 2)
}

func Test_MemoryStore_Search_WithNoMatches_Should_ReturnEmpty(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note := agent.NewMemoryNote("note-1", agent.SourceTypePreference).
		WithRawContent("something")
	_ = store.Write(context.Background(), note)

	// Act
	results, err := store.Search(context.Background(), "nonexistent", 10, nil)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should return empty slice", len(results), 0)
}

func Test_MemoryStore_Search_Should_FilterBySourceTypes(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note1 := agent.NewMemoryNote("note-1", agent.SourceTypeDecision).
		WithRawContent("architectural decision")
	note2 := agent.NewMemoryNote("note-2", agent.SourceTypeFact).
		WithRawContent("fact about system")
	note3 := agent.NewMemoryNote("note-3", agent.SourceTypeRequirement).
		WithRawContent("requirement spec")
	_ = store.Write(context.Background(), note1)
	_ = store.Write(context.Background(), note2)
	_ = store.Write(context.Background(), note3)

	// Act
	opts := &agent.MemorySearchOptions{SourceTypes: []agent.SourceType{agent.SourceTypeDecision, agent.SourceTypeRequirement}}
	results, err := store.Search(context.Background(), "", 10, opts)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should find 2 results", len(results), 2)
}

func Test_MemoryStore_Search_Should_FilterBySingleSourceType(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note1 := agent.NewMemoryNote("note-1", agent.SourceTypeIssue).
		WithRawContent("bug report")
	note2 := agent.NewMemoryNote("note-2", agent.SourceTypeFact).
		WithRawContent("fact")
	_ = store.Write(context.Background(), note1)
	_ = store.Write(context.Background(), note2)

	// Act
	opts := &agent.MemorySearchOptions{SourceTypes: []agent.SourceType{agent.SourceTypeIssue}}
	results, err := store.Search(context.Background(), "", 10, opts)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should find 1 result", len(results), 1)
	assert.That(t, "result should be issue type", results[0].SourceType, agent.SourceTypeIssue)
}

func Test_MemoryStore_Search_Should_FilterByMinImportance(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note1 := agent.NewMemoryNote("note-1", agent.SourceTypeFact).
		WithRawContent("low importance").WithImportance(1)
	note2 := agent.NewMemoryNote("note-2", agent.SourceTypeFact).
		WithRawContent("medium importance").WithImportance(3)
	note3 := agent.NewMemoryNote("note-3", agent.SourceTypeFact).
		WithRawContent("high importance").WithImportance(5)
	_ = store.Write(context.Background(), note1)
	_ = store.Write(context.Background(), note2)
	_ = store.Write(context.Background(), note3)

	// Act
	opts := &agent.MemorySearchOptions{MinImportance: 3}
	results, err := store.Search(context.Background(), "importance", 10, opts)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should find 2 results with importance >= 3", len(results), 2)
}

func Test_MemoryStore_Search_Should_CombineSourceTypeAndMinImportance(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryMemoryStore()
	note1 := agent.NewMemoryNote("note-1", agent.SourceTypeDecision).
		WithRawContent("decision").WithImportance(2)
	note2 := agent.NewMemoryNote("note-2", agent.SourceTypeDecision).
		WithRawContent("decision").WithImportance(4)
	note3 := agent.NewMemoryNote("note-3", agent.SourceTypeFact).
		WithRawContent("fact").WithImportance(5)
	_ = store.Write(context.Background(), note1)
	_ = store.Write(context.Background(), note2)
	_ = store.Write(context.Background(), note3)

	// Act
	opts := &agent.MemorySearchOptions{
		SourceTypes:   []agent.SourceType{agent.SourceTypeDecision},
		MinImportance: 3,
	}
	results, err := store.Search(context.Background(), "", 10, opts)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should find 1 result (decision with importance >= 3)", len(results), 1)
	assert.That(t, "result should be decision type", results[0].SourceType, agent.SourceTypeDecision)
	assert.That(t, "result importance should be >= 3", results[0].Importance >= 3, true)
}
