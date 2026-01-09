package agent_test

import (
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/pkg/agent"
)

func Test_NewMemoryNote_Should_CreateWithDefaults(t *testing.T) {
	// Arrange
	id := agent.NoteID("note-123")
	sourceType := agent.SourceTypePreference

	// Act
	note := agent.NewMemoryNote(id, sourceType)

	// Assert
	assert.That(t, "id must match", note.ID, id)
	assert.That(t, "source type must match", note.SourceType, sourceType)
	assert.That(t, "importance must be 1", note.Importance, 1)
	assert.That(t, "keywords must be empty slice", len(note.Keywords), 0)
	assert.That(t, "tags must be empty slice", len(note.Tags), 0)
}

func Test_MemoryNote_WithUserID_Should_SetUserID(t *testing.T) {
	// Arrange
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference)

	// Act
	note.WithUserID("user-456")

	// Assert
	assert.That(t, "user id must match", note.UserID, "user-456")
}

func Test_MemoryNote_WithSessionID_Should_SetSessionID(t *testing.T) {
	// Arrange
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference)

	// Act
	note.WithSessionID("session-789")

	// Assert
	assert.That(t, "session id must match", note.SessionID, "session-789")
}

func Test_MemoryNote_WithTaskID_Should_SetTaskID(t *testing.T) {
	// Arrange
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference)

	// Act
	note.WithTaskID("task-abc")

	// Assert
	assert.That(t, "task id must match", note.TaskID, "task-abc")
}

func Test_MemoryNote_WithRawContent_Should_SetContentAndUpdateTime(t *testing.T) {
	// Arrange
	note := agent.NewMemoryNote("note-123", agent.SourceTypeUserMessage)
	originalTime := note.UpdatedAt
	time.Sleep(1 * time.Millisecond) // Ensure time difference

	// Act
	note.WithRawContent("User prefers German responses")

	// Assert
	assert.That(t, "raw content must match", note.RawContent, "User prefers German responses")
	assert.That(t, "updated_at must be newer", note.UpdatedAt.After(originalTime), true)
}

func Test_MemoryNote_WithSummary_Should_SetSummary(t *testing.T) {
	// Arrange
	note := agent.NewMemoryNote("note-123", agent.SourceTypeSummary)

	// Act
	note.WithSummary("User preference for German language")

	// Assert
	assert.That(t, "summary must match", note.Summary, "User preference for German language")
}

func Test_MemoryNote_WithContextDescription_Should_SetDescription(t *testing.T) {
	// Arrange
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference)

	// Act
	note.WithContextDescription("Apply this preference to all future responses")

	// Assert
	assert.That(t, "context description must match", note.ContextDescription, "Apply this preference to all future responses")
}

func Test_MemoryNote_WithKeywords_Should_SetKeywords(t *testing.T) {
	// Arrange
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference)

	// Act
	note.WithKeywords("language", "german", "formatting")

	// Assert
	assert.That(t, "keywords length must be 3", len(note.Keywords), 3)
	assert.That(t, "first keyword must match", note.Keywords[0], "language")
	assert.That(t, "second keyword must match", note.Keywords[1], "german")
}

func Test_MemoryNote_WithTags_Should_SetTags(t *testing.T) {
	// Arrange
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference)

	// Act
	note.WithTags("preference", "formatting")

	// Assert
	assert.That(t, "tags length must be 2", len(note.Tags), 2)
	assert.That(t, "first tag must match", note.Tags[0], "preference")
}

func Test_MemoryNote_WithImportance_Should_ClampToValidRange(t *testing.T) {
	// Arrange
	note1 := agent.NewMemoryNote("note-1", agent.SourceTypePreference)
	note2 := agent.NewMemoryNote("note-2", agent.SourceTypePreference)
	note3 := agent.NewMemoryNote("note-3", agent.SourceTypePreference)

	// Act
	note1.WithImportance(0)  // Below minimum
	note2.WithImportance(3)  // Valid
	note3.WithImportance(10) // Above maximum

	// Assert
	assert.That(t, "importance below 1 must clamp to 1", note1.Importance, 1)
	assert.That(t, "valid importance must remain unchanged", note2.Importance, 3)
	assert.That(t, "importance above 5 must clamp to 5", note3.Importance, 5)
}

func Test_MemoryNote_HasTag_Should_ReturnTrueWhenTagExists(t *testing.T) {
	// Arrange
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference).
		WithTags("preference", "formatting")

	// Act & Assert
	assert.That(t, "should find existing tag", note.HasTag("preference"), true)
	assert.That(t, "should not find non-existing tag", note.HasTag("nonexistent"), false)
}

func Test_MemoryNote_HasKeyword_Should_ReturnTrueWhenKeywordExists(t *testing.T) {
	// Arrange
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference).
		WithKeywords("language", "german")

	// Act & Assert
	assert.That(t, "should find existing keyword", note.HasKeyword("language"), true)
	assert.That(t, "should not find non-existing keyword", note.HasKeyword("nonexistent"), false)
}

func Test_MemoryNote_SearchableText_Should_CombineAllTextFields(t *testing.T) {
	// Arrange
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference).
		WithRawContent("User prefers German").
		WithSummary("Language preference").
		WithContextDescription("Apply to all responses")

	// Act
	text := note.SearchableText()

	// Assert
	assert.That(t, "searchable text must contain raw content", text, "User prefers German Language preference Apply to all responses")
}

func Test_MemoryNote_SearchableText_WithPartialFields_Should_ReturnAvailableText(t *testing.T) {
	// Arrange
	note := agent.NewMemoryNote("note-123", agent.SourceTypeUserMessage).
		WithRawContent("Just raw content")

	// Act
	text := note.SearchableText()

	// Assert
	assert.That(t, "searchable text must be raw content only", text, "Just raw content")
}

func Test_MemoryNote_MethodChaining_Should_Work(t *testing.T) {
	// Arrange & Act
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference).
		WithUserID("user-1").
		WithSessionID("session-1").
		WithRawContent("Content").
		WithSummary("Summary").
		WithContextDescription("Context").
		WithKeywords("key1", "key2").
		WithTags("tag1", "tag2").
		WithImportance(4)

	// Assert
	assert.That(t, "user id must match", note.UserID, "user-1")
	assert.That(t, "session id must match", note.SessionID, "session-1")
	assert.That(t, "raw content must match", note.RawContent, "Content")
	assert.That(t, "summary must match", note.Summary, "Summary")
	assert.That(t, "context description must match", note.ContextDescription, "Context")
	assert.That(t, "keywords length must be 2", len(note.Keywords), 2)
	assert.That(t, "tags length must be 2", len(note.Tags), 2)
	assert.That(t, "importance must be 4", note.Importance, 4)
}
