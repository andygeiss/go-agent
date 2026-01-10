package memorizing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/memorizing"
)

// ErrMockNotFound is returned when a note is not found in the mock store.
var ErrMockNotFound = errors.New("note not found")

// mockMemoryStore is a test double for the MemoryStore interface.
type mockMemoryStore struct {
	notes       map[agent.NoteID]*agent.MemoryNote
	writeErr    error
	searchErr   error
	getErr      error
	deleteErr   error
	searchNotes []*agent.MemoryNote
}

func newMockMemoryStore() *mockMemoryStore {
	return &mockMemoryStore{
		notes: make(map[agent.NoteID]*agent.MemoryNote),
	}
}

func (m *mockMemoryStore) Write(_ context.Context, note *agent.MemoryNote) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.notes[note.ID] = note
	return nil
}

func (m *mockMemoryStore) Search(_ context.Context, _ string, limit int, _ *agent.MemorySearchOptions) ([]*agent.MemoryNote, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	if m.searchNotes != nil {
		if limit < len(m.searchNotes) {
			return m.searchNotes[:limit], nil
		}
		return m.searchNotes, nil
	}
	return []*agent.MemoryNote{}, nil
}

func (m *mockMemoryStore) Get(_ context.Context, id agent.NoteID) (*agent.MemoryNote, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	note, ok := m.notes[id]
	if !ok {
		return nil, ErrMockNotFound
	}
	return note, nil
}

func (m *mockMemoryStore) Delete(_ context.Context, id agent.NoteID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.notes, id)
	return nil
}

func Test_Service_WriteNote_Should_StoreNote(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := memorizing.NewService(store)
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference).
		WithRawContent("User prefers German")

	// Act
	err := svc.WriteNote(context.Background(), note)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "note must be stored", store.notes["note-123"], note)
}

func Test_Service_WriteNote_WithNilNote_Should_ReturnError(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := memorizing.NewService(store)

	// Act
	err := svc.WriteNote(context.Background(), nil)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_Service_WriteNote_WithEmptyID_Should_ReturnError(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := memorizing.NewService(store)
	note := agent.NewMemoryNote("", agent.SourceTypePreference)

	// Act
	err := svc.WriteNote(context.Background(), note)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_Service_SearchNotes_Should_ReturnNotes(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	store.searchNotes = []*agent.MemoryNote{
		agent.NewMemoryNote("note-1", agent.SourceTypePreference),
		agent.NewMemoryNote("note-2", agent.SourceTypePreference),
	}
	svc := memorizing.NewService(store)

	// Act
	notes, err := svc.SearchNotes(context.Background(), "preferences", 10, nil)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "notes count must be 2", len(notes), 2)
}

func Test_Service_SearchNotes_WithZeroLimit_Should_UseDefaultLimit(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := memorizing.NewService(store)

	// Act
	_, err := svc.SearchNotes(context.Background(), "test", 0, nil)

	// Assert
	assert.That(t, "error must be nil", err, nil)
}

func Test_Service_GetNote_Should_ReturnNote(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	originalNote := agent.NewMemoryNote("note-123", agent.SourceTypePreference)
	store.notes["note-123"] = originalNote
	svc := memorizing.NewService(store)

	// Act
	note, err := svc.GetNote(context.Background(), "note-123")

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "note must match", note, originalNote)
}

func Test_Service_GetNote_WithEmptyID_Should_ReturnError(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := memorizing.NewService(store)

	// Act
	_, err := svc.GetNote(context.Background(), "")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_Service_DeleteNote_Should_RemoveNote(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	store.notes["note-123"] = agent.NewMemoryNote("note-123", agent.SourceTypePreference)
	svc := memorizing.NewService(store)

	// Act
	err := svc.DeleteNote(context.Background(), "note-123")

	// Assert
	assert.That(t, "error must be nil", err, nil)
	_, exists := store.notes["note-123"]
	assert.That(t, "note must be deleted from store", exists, false)
}

func Test_Service_DeleteNote_WithEmptyID_Should_ReturnError(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := memorizing.NewService(store)

	// Act
	err := svc.DeleteNote(context.Background(), "")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

// DeleteNoteUseCase tests

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

// GetNoteUseCase tests

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

// SearchNotesUseCase tests

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

// WriteNoteUseCase tests

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

func Test_WriteNoteUseCase_Execute_WithNilNote_Should_ReturnError(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	uc := memorizing.NewWriteNoteUseCase(store)

	// Act
	err := uc.Execute(context.Background(), nil)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

// Service schema-focused use case tests

func Test_Service_WriteTypedNote_Should_CreateAndStoreNote(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := memorizing.NewService(store)

	// Act
	err := svc.WriteTypedNote(context.Background(), "note-1", agent.SourceTypeDecision, "Use PostgreSQL", nil)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "note must be stored", store.notes["note-1"] != nil, true)
	assert.That(t, "source type must be decision", store.notes["note-1"].SourceType, agent.SourceTypeDecision)
}

func Test_Service_WriteTypedNote_WithOptions_Should_ApplyOptions(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := memorizing.NewService(store)
	opts := &memorizing.TypedNoteOptions{
		Tags:       []string{"architecture", "database"},
		Importance: 5,
		UserID:     "user-1",
	}

	// Act
	err := svc.WriteTypedNote(context.Background(), "note-1", agent.SourceTypeRequirement, "Must support 1000 users", opts)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	note := store.notes["note-1"]
	assert.That(t, "importance must be 5", note.Importance, 5)
	assert.That(t, "user id must match", note.UserID, "user-1")
	assert.That(t, "tags must be set", len(note.Tags), 2)
}

func Test_Service_WriteTypedNote_WithEmptyID_Should_ReturnError(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := memorizing.NewService(store)

	// Act
	err := svc.WriteTypedNote(context.Background(), "", agent.SourceTypeFact, "content", nil)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_Service_SearchBySourceTypes_Should_FilterByTypes(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	store.searchNotes = []*agent.MemoryNote{
		agent.NewMemoryNote("note-1", agent.SourceTypeDecision),
	}
	svc := memorizing.NewService(store)

	// Act
	results, err := svc.SearchBySourceTypes(context.Background(), "test", []agent.SourceType{agent.SourceTypeDecision}, 10)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should return results", len(results), 1)
}

func Test_Service_SearchDecisions_Should_FilterByDecisionType(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	store.searchNotes = []*agent.MemoryNote{
		agent.NewMemoryNote("note-1", agent.SourceTypeDecision),
	}
	svc := memorizing.NewService(store)

	// Act
	results, err := svc.SearchDecisions(context.Background(), "architecture", 10)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should return results", len(results), 1)
}

func Test_Service_SearchFacts_Should_FilterByFactType(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	store.searchNotes = []*agent.MemoryNote{
		agent.NewMemoryNote("note-1", agent.SourceTypeFact),
	}
	svc := memorizing.NewService(store)

	// Act
	results, err := svc.SearchFacts(context.Background(), "api", 10)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should return results", len(results), 1)
}

func Test_Service_SearchRequirements_Should_FilterByRequirementType(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	store.searchNotes = []*agent.MemoryNote{
		agent.NewMemoryNote("note-1", agent.SourceTypeRequirement),
	}
	svc := memorizing.NewService(store)

	// Act
	results, err := svc.SearchRequirements(context.Background(), "scalability", 10)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "should return results", len(results), 1)
}
