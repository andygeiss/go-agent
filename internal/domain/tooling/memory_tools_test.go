package tooling_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/tooling"
	"github.com/andygeiss/go-agent/pkg/agent"
)

// ErrMockNotFound is returned when a note is not found in the mock store.
var ErrMockNotFound = errors.New("note not found")

// mockMemoryStore is a test double for the MemoryStore interface.
type mockMemoryStore struct {
	notes       map[agent.NoteID]*agent.MemoryNote
	writeErr    error
	searchErr   error
	getErr      error
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
		if limit > 0 && limit < len(m.searchNotes) {
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
	delete(m.notes, id)
	return nil
}

func testIDGenerator() func() string {
	counter := 0
	return func() string {
		counter++
		return "test-note-" + string(rune('0'+counter))
	}
}

func Test_MemoryToolService_MemoryWrite_Should_StoreNote(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := tooling.NewMemoryToolService(store, testIDGenerator())

	args := `{
		"source_type": "preference",
		"raw_content": "User prefers German responses",
		"summary": "Language preference: German",
		"context_description": "Apply to all future responses",
		"keywords": ["language", "german"],
		"tags": ["preference", "formatting"],
		"importance": 4
	}`

	// Act
	result, err := svc.MemoryWrite(context.Background(), args)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	assert.That(t, "result must contain success", result != "", true)

	// Verify note was stored
	assert.That(t, "store must have 1 note", len(store.notes), 1)
	for _, note := range store.notes {
		assert.That(t, "source type must be preference", note.SourceType, agent.SourceTypePreference)
		assert.That(t, "raw content must match", note.RawContent, "User prefers German responses")
		assert.That(t, "importance must be 4", note.Importance, 4)
	}
}

func Test_MemoryToolService_MemoryWrite_WithDefaultUserID_Should_UseDefault(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := tooling.NewMemoryToolService(store, testIDGenerator()).
		WithUserID("default-user")

	args := `{
		"source_type": "fact",
		"raw_content": "Some fact",
		"summary": "A fact"
	}`

	// Act
	_, err := svc.MemoryWrite(context.Background(), args)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	for _, note := range store.notes {
		assert.That(t, "user id must be default", note.UserID, "default-user")
	}
}

func Test_MemoryToolService_MemoryWrite_WithExplicitUserID_Should_Override(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := tooling.NewMemoryToolService(store, testIDGenerator()).
		WithUserID("default-user")

	args := `{
		"source_type": "fact",
		"raw_content": "Some fact",
		"summary": "A fact",
		"user_id": "explicit-user"
	}`

	// Act
	_, err := svc.MemoryWrite(context.Background(), args)

	// Assert
	assert.That(t, "error must be nil", err, nil)
	for _, note := range store.notes {
		assert.That(t, "user id must be explicit", note.UserID, "explicit-user")
	}
}

func Test_MemoryToolService_MemorySearch_Should_ReturnResults(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	store.searchNotes = []*agent.MemoryNote{
		agent.NewMemoryNote("note-1", agent.SourceTypePreference).
			WithSummary("Language preference").
			WithContextDescription("Apply to responses").
			WithTags("preference").
			WithImportance(4),
		agent.NewMemoryNote("note-2", agent.SourceTypeToolResult).
			WithSummary("API result").
			WithTags("api").
			WithImportance(2),
	}
	svc := tooling.NewMemoryToolService(store, testIDGenerator())

	args := `{"query": "preferences", "limit": 10}`

	// Act
	result, err := svc.MemorySearch(context.Background(), args)

	// Assert
	assert.That(t, "error must be nil", err, nil)

	var response map[string]any
	_ = json.Unmarshal([]byte(result), &response)
	assert.That(t, "status must be success", response["status"], "success")
	assert.That(t, "count must be 2", int(response["count"].(float64)), 2)
}

func Test_MemoryToolService_MemorySearch_WithFilters_Should_PassOptions(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	store.searchNotes = []*agent.MemoryNote{}
	svc := tooling.NewMemoryToolService(store, testIDGenerator())

	args := `{
		"query": "preferences",
		"user_id": "user-123",
		"tags": ["preference"]
	}`

	// Act
	_, err := svc.MemorySearch(context.Background(), args)

	// Assert
	assert.That(t, "error must be nil", err, nil)
}

func Test_MemoryToolService_MemoryGet_Should_ReturnNote(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	existingNote := agent.NewMemoryNote("note-123", agent.SourceTypePreference).
		WithRawContent("User prefers German").
		WithSummary("Language preference").
		WithKeywords("german", "language").
		WithTags("preference").
		WithImportance(4)
	store.notes["note-123"] = existingNote
	svc := tooling.NewMemoryToolService(store, testIDGenerator())

	args := `{"id": "note-123"}`

	// Act
	result, err := svc.MemoryGet(context.Background(), args)

	// Assert
	assert.That(t, "error must be nil", err, nil)

	var response map[string]any
	_ = json.Unmarshal([]byte(result), &response)
	assert.That(t, "status must be success", response["status"], "success")
	note := response["note"].(map[string]any)
	assert.That(t, "note id must match", note["id"], "note-123")
	assert.That(t, "raw content must match", note["raw_content"], "User prefers German")
}

func Test_MemoryToolService_MemoryGet_WithNonexistent_Should_ReturnNotFound(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := tooling.NewMemoryToolService(store, testIDGenerator())

	args := `{"id": "nonexistent"}`

	// Act
	result, err := svc.MemoryGet(context.Background(), args)

	// Assert
	// Note: The tool returns an error when note is not found
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "result must be empty on error", result, "")
}

func Test_NewMemoryWriteTool_Should_CreateValidTool(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := tooling.NewMemoryToolService(store, testIDGenerator())

	// Act
	tool := tooling.NewMemoryWriteTool(svc)

	// Assert
	assert.That(t, "tool id must be memory_write", tool.ID, agent.ToolID("memory_write"))
	assert.That(t, "tool name must match", tool.Definition.Name, "memory_write")
	assert.That(t, "tool must have parameters", len(tool.Definition.Parameters) > 0, true)
	assert.That(t, "tool must have source_type param", tool.Definition.HasParameter("source_type"), true)
	assert.That(t, "tool must have raw_content param", tool.Definition.HasParameter("raw_content"), true)
	assert.That(t, "tool must have summary param", tool.Definition.HasParameter("summary"), true)
}

func Test_NewMemorySearchTool_Should_CreateValidTool(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := tooling.NewMemoryToolService(store, testIDGenerator())

	// Act
	tool := tooling.NewMemorySearchTool(svc)

	// Assert
	assert.That(t, "tool id must be memory_search", tool.ID, agent.ToolID("memory_search"))
	assert.That(t, "tool name must match", tool.Definition.Name, "memory_search")
	assert.That(t, "tool must have query param", tool.Definition.HasParameter("query"), true)
	assert.That(t, "tool must have limit param", tool.Definition.HasParameter("limit"), true)
}

func Test_NewMemoryGetTool_Should_CreateValidTool(t *testing.T) {
	// Arrange
	store := newMockMemoryStore()
	svc := tooling.NewMemoryToolService(store, testIDGenerator())

	// Act
	tool := tooling.NewMemoryGetTool(svc)

	// Assert
	assert.That(t, "tool id must be memory_get", tool.ID, agent.ToolID("memory_get"))
	assert.That(t, "tool name must match", tool.Definition.Name, "memory_get")
	assert.That(t, "tool must have id param", tool.Definition.HasParameter("id"), true)
}
