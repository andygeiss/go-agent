package agent_test

import (
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent"
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

func Test_MemoryNote_WithEmbedding_Should_SetEmbedding(t *testing.T) {
	// Arrange
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference)
	embedding := agent.Embedding{0.1, 0.2, 0.3, 0.4}

	// Act
	note.WithEmbedding(embedding)

	// Assert
	assert.That(t, "embedding length must be 4", len(note.Embedding), 4)
	assert.That(t, "first element must match", note.Embedding[0], float32(0.1))
	assert.That(t, "last element must match", note.Embedding[3], float32(0.4))
}

func Test_MemoryNote_WithEmbedding_Should_UpdateTimestamp(t *testing.T) {
	// Arrange
	note := agent.NewMemoryNote("note-123", agent.SourceTypePreference)
	initialTime := note.UpdatedAt

	// Act
	note.WithEmbedding(agent.Embedding{1.0, 2.0})

	// Assert
	assert.That(t, "updated_at must be >= initial time", note.UpdatedAt.After(initialTime) || note.UpdatedAt.Equal(initialTime), true)
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
		WithContextDescription("Apply to all responses").
		WithTags("language", "preference").
		WithKeywords("german", "locale")

	// Act
	text := note.SearchableText()

	// Assert
	assert.That(t, "searchable text must contain raw content", text, "User prefers German Language preference Apply to all responses language preference german locale")
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

// Tests for new source type validation

func Test_ValidSourceTypes_Should_ReturnAllTypes(t *testing.T) {
	// Act
	types := agent.ValidSourceTypes()

	// Assert
	assert.That(t, "should have 12 source types", len(types), 12)
}

func Test_IsValidSourceType_Should_ReturnTrueForValidTypes(t *testing.T) {
	// Arrange
	validTypes := []agent.SourceType{
		agent.SourceTypeDecision,
		agent.SourceTypeExperiment,
		agent.SourceTypeExternalSource,
		agent.SourceTypeFact,
		agent.SourceTypeIssue,
		agent.SourceTypePlanStep,
		agent.SourceTypePreference,
		agent.SourceTypeRequirement,
		agent.SourceTypeRetrospective,
		agent.SourceTypeSummary,
		agent.SourceTypeToolResult,
		agent.SourceTypeUserMessage,
	}

	// Act & Assert
	for _, st := range validTypes {
		assert.That(t, "source type should be valid: "+string(st), agent.IsValidSourceType(st), true)
	}
}

func Test_IsValidSourceType_Should_ReturnFalseForInvalidType(t *testing.T) {
	// Act & Assert
	assert.That(t, "invalid source type should return false", agent.IsValidSourceType("invalid_type"), false)
}

func Test_ParseSourceType_Should_ReturnCorrectType(t *testing.T) {
	// Act & Assert
	assert.That(t, "should parse decision", agent.ParseSourceType("decision"), agent.SourceTypeDecision)
	assert.That(t, "should parse fact", agent.ParseSourceType("fact"), agent.SourceTypeFact)
	assert.That(t, "should parse preference", agent.ParseSourceType("preference"), agent.SourceTypePreference)
}

func Test_ParseSourceType_WithInvalidString_Should_ReturnFact(t *testing.T) {
	// Act
	result := agent.ParseSourceType("invalid")

	// Assert
	assert.That(t, "should return fact for invalid string", result, agent.SourceTypeFact)
}

// Tests for schema-convention helper constructors

func Test_NewDecisionNote_Should_CreateWithCorrectDefaults(t *testing.T) {
	// Act
	note := agent.NewDecisionNote("note-1", "Use PostgreSQL for persistence", "database", "architecture")

	// Assert
	assert.That(t, "source type should be decision", note.SourceType, agent.SourceTypeDecision)
	assert.That(t, "importance should be 4", note.Importance, 4)
	assert.That(t, "should have decision tag", note.HasTag("decision"), true)
	assert.That(t, "should have custom tag", note.HasTag("database"), true)
	assert.That(t, "raw content should match", note.RawContent, "Use PostgreSQL for persistence")
}

func Test_NewExperimentNote_Should_CreateWithHypothesisAndResult(t *testing.T) {
	// Act
	note := agent.NewExperimentNote("note-1", "Caching improves latency", "Latency reduced by 50%", "performance")

	// Assert
	assert.That(t, "source type should be experiment", note.SourceType, agent.SourceTypeExperiment)
	assert.That(t, "importance should be 3", note.Importance, 3)
	assert.That(t, "should have experiment tag", note.HasTag("experiment"), true)
	assert.That(t, "content should contain hypothesis", note.RawContent != "", true)
}

func Test_NewExternalSourceNote_Should_CreateWithURLAndAnnotation(t *testing.T) {
	// Act
	note := agent.NewExternalSourceNote("note-1", "https://example.com/docs", "API documentation", "api")

	// Assert
	assert.That(t, "source type should be external_source", note.SourceType, agent.SourceTypeExternalSource)
	assert.That(t, "importance should be 2", note.Importance, 2)
	assert.That(t, "should have external tag", note.HasTag("external"), true)
	assert.That(t, "should have reference tag", note.HasTag("reference"), true)
}

func Test_NewFactNote_Should_CreateWithCorrectDefaults(t *testing.T) {
	// Act
	note := agent.NewFactNote("note-1", "The API rate limit is 100 requests per minute", "api", "limits")

	// Assert
	assert.That(t, "source type should be fact", note.SourceType, agent.SourceTypeFact)
	assert.That(t, "importance should be 3", note.Importance, 3)
	assert.That(t, "should have fact tag", note.HasTag("fact"), true)
}

func Test_NewIssueNote_Should_CreateWithCorrectDefaults(t *testing.T) {
	// Act
	note := agent.NewIssueNote("note-1", "Memory leak in connection pool", "bug", "critical")

	// Assert
	assert.That(t, "source type should be issue", note.SourceType, agent.SourceTypeIssue)
	assert.That(t, "importance should be 4", note.Importance, 4)
	assert.That(t, "should have issue tag", note.HasTag("issue"), true)
	assert.That(t, "should have problem tag", note.HasTag("problem"), true)
}

func Test_NewPlanStepNote_Should_CreateWithPlanContext(t *testing.T) {
	// Act
	note := agent.NewPlanStepNote("note-1", "Implement user authentication", "plan-123", 2, "security")

	// Assert
	assert.That(t, "source type should be plan_step", note.SourceType, agent.SourceTypePlanStep)
	assert.That(t, "importance should be 3", note.Importance, 3)
	assert.That(t, "should have plan tag", note.HasTag("plan"), true)
	assert.That(t, "should have step tag", note.HasTag("step"), true)
	assert.That(t, "task id should be plan id", note.TaskID, "plan-123")
}

func Test_NewPreferenceNote_Should_CreateWithCorrectDefaults(t *testing.T) {
	// Act
	note := agent.NewPreferenceNote("note-1", "User prefers dark mode", "ui", "settings")

	// Assert
	assert.That(t, "source type should be preference", note.SourceType, agent.SourceTypePreference)
	assert.That(t, "importance should be 4", note.Importance, 4)
	assert.That(t, "should have preference tag", note.HasTag("preference"), true)
}

func Test_NewRequirementNote_Should_CreateWithHighImportance(t *testing.T) {
	// Act
	note := agent.NewRequirementNote("note-1", "System must support 1000 concurrent users", "scalability")

	// Assert
	assert.That(t, "source type should be requirement", note.SourceType, agent.SourceTypeRequirement)
	assert.That(t, "importance should be 5", note.Importance, 5)
	assert.That(t, "should have requirement tag", note.HasTag("requirement"), true)
}

func Test_NewRetrospectiveNote_Should_CreateWithCorrectDefaults(t *testing.T) {
	// Act
	note := agent.NewRetrospectiveNote("note-1", "Early testing catches more bugs", "process")

	// Assert
	assert.That(t, "source type should be retrospective", note.SourceType, agent.SourceTypeRetrospective)
	assert.That(t, "importance should be 3", note.Importance, 3)
	assert.That(t, "should have retrospective tag", note.HasTag("retrospective"), true)
	assert.That(t, "should have lessons-learned tag", note.HasTag("lessons-learned"), true)
}

func Test_NewSummaryNote_Should_CreateWithSourceReferences(t *testing.T) {
	// Act
	note := agent.NewSummaryNote("note-1", "Summary of today's decisions", []string{"note-a", "note-b"}, "daily")

	// Assert
	assert.That(t, "source type should be summary", note.SourceType, agent.SourceTypeSummary)
	assert.That(t, "importance should be 3", note.Importance, 3)
	assert.That(t, "should have summary tag", note.HasTag("summary"), true)
	assert.That(t, "context should reference sources", note.ContextDescription != "", true)
}
