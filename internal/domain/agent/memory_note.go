package agent

import (
	"strings"
	"time"

	"github.com/andygeiss/cloud-native-utils/slices"
)

// NoteID is the unique identifier for a memory note.
// It is a value object that ensures type safety for note references.
type NoteID string

// SourceType categorizes what created a memory note.
type SourceType string

// Standard source types for memory notes (alphabetically sorted).
const (
	SourceTypeDecision       SourceType = "decision"
	SourceTypeExperiment     SourceType = "experiment"
	SourceTypeExternalSource SourceType = "external_source"
	SourceTypeFact           SourceType = "fact"
	SourceTypeIssue          SourceType = "issue"
	SourceTypePlanStep       SourceType = "plan_step"
	SourceTypePreference     SourceType = "preference"
	SourceTypeRequirement    SourceType = "requirement"
	SourceTypeRetrospective  SourceType = "retrospective"
	SourceTypeSummary        SourceType = "summary"
	SourceTypeToolResult     SourceType = "tool_result"
	SourceTypeUserMessage    SourceType = "user_message"
)

// ValidSourceTypes returns all valid source type values.
func ValidSourceTypes() []SourceType {
	return []SourceType{
		SourceTypeDecision,
		SourceTypeExperiment,
		SourceTypeExternalSource,
		SourceTypeFact,
		SourceTypeIssue,
		SourceTypePlanStep,
		SourceTypePreference,
		SourceTypeRequirement,
		SourceTypeRetrospective,
		SourceTypeSummary,
		SourceTypeToolResult,
		SourceTypeUserMessage,
	}
}

// IsValidSourceType checks if the given source type is valid.
func IsValidSourceType(st SourceType) bool {
	return slices.Contains(ValidSourceTypes(), st)
}

// ParseSourceType converts a string to a SourceType.
// Returns SourceTypeFact if the string is not recognized.
func ParseSourceType(s string) SourceType {
	st := SourceType(s)
	if IsValidSourceType(st) {
		return st
	}
	return SourceTypeFact
}

// MemoryNote represents an atomic unit of long-term memory.
// Notes are stored with semantic enrichment for retrieval.
type MemoryNote struct {
	// Identity & scope
	ID        NoteID    `json:"id"`
	UserID    string    `json:"user_id,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
	TaskID    string    `json:"task_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Core content
	SourceType SourceType `json:"source_type"`
	RawContent string     `json:"raw_content"`
	Summary    string     `json:"summary"`

	// Semantic enrichment
	ContextDescription string   `json:"context_description"`
	Keywords           []string `json:"keywords"`
	Tags               []string `json:"tags"`
	Importance         int      `json:"importance"` // 1-5 scale
}

// NewMemoryNote creates a new MemoryNote with the given ID and source type.
func NewMemoryNote(id NoteID, sourceType SourceType) *MemoryNote {
	now := time.Now()
	return &MemoryNote{
		ID:         id,
		SourceType: sourceType,
		CreatedAt:  now,
		UpdatedAt:  now,
		Keywords:   make([]string, 0),
		Tags:       make([]string, 0),
		Importance: 1,
	}
}

// Schema-convention helper constructors.
// These factory functions create MemoryNote instances with consistent defaults
// for each source type, encoding best practices in one place.

// NewDecisionNote creates a decision note with appropriate defaults.
// Decisions have high importance (4) and are tagged with "decision".
func NewDecisionNote(id NoteID, content string, tags ...string) *MemoryNote {
	allTags := append([]string{"decision"}, tags...)
	return NewMemoryNote(id, SourceTypeDecision).
		WithRawContent(content).
		WithSummary(content).
		WithTags(allTags...).
		WithImportance(4)
}

// NewExperimentNote creates an experiment note with hypothesis and result.
// Experiments have medium importance (3) and are tagged with "experiment".
func NewExperimentNote(id NoteID, hypothesis, result string, tags ...string) *MemoryNote {
	allTags := append([]string{"experiment"}, tags...)
	content := "Hypothesis: " + hypothesis + "\nResult: " + result
	return NewMemoryNote(id, SourceTypeExperiment).
		WithRawContent(content).
		WithSummary("Experiment: " + hypothesis).
		WithTags(allTags...).
		WithImportance(3)
}

// NewExternalSourceNote creates a note referencing an external URL or source.
// External sources have medium importance (2) and are tagged with "external".
func NewExternalSourceNote(id NoteID, url, annotation string, tags ...string) *MemoryNote {
	allTags := append([]string{"external", "reference"}, tags...)
	content := "URL: " + url + "\nAnnotation: " + annotation
	return NewMemoryNote(id, SourceTypeExternalSource).
		WithRawContent(content).
		WithSummary(annotation).
		WithTags(allTags...).
		WithImportance(2)
}

// NewFactNote creates a fact note with appropriate defaults.
// Facts have medium importance (3) and are tagged with "fact".
func NewFactNote(id NoteID, content string, tags ...string) *MemoryNote {
	allTags := append([]string{"fact"}, tags...)
	return NewMemoryNote(id, SourceTypeFact).
		WithRawContent(content).
		WithSummary(content).
		WithTags(allTags...).
		WithImportance(3)
}

// NewIssueNote creates an issue note for tracking problems or bugs.
// Issues have high importance (4) and are tagged with "issue".
func NewIssueNote(id NoteID, description string, tags ...string) *MemoryNote {
	allTags := append([]string{"issue", "problem"}, tags...)
	return NewMemoryNote(id, SourceTypeIssue).
		WithRawContent(description).
		WithSummary(description).
		WithTags(allTags...).
		WithImportance(4)
}

// NewPlanStepNote creates a plan step note with plan ID and step index.
// Plan steps have medium importance (3) and are tagged with "plan".
func NewPlanStepNote(id NoteID, content, planID string, stepIndex int, tags ...string) *MemoryNote {
	allTags := append([]string{"plan", "step"}, tags...)
	return NewMemoryNote(id, SourceTypePlanStep).
		WithRawContent(content).
		WithSummary(content).
		WithTaskID(planID).
		WithTags(allTags...).
		WithContextDescription("Plan: " + planID + ", Step: " + string(rune('0'+stepIndex))).
		WithImportance(3)
}

// NewPreferenceNote creates a preference note with appropriate defaults.
// Preferences have high importance (4) and are tagged with "preference".
func NewPreferenceNote(id NoteID, content string, tags ...string) *MemoryNote {
	allTags := append([]string{"preference"}, tags...)
	return NewMemoryNote(id, SourceTypePreference).
		WithRawContent(content).
		WithSummary(content).
		WithTags(allTags...).
		WithImportance(4)
}

// NewRequirementNote creates a requirement note with appropriate defaults.
// Requirements have high importance (5) and are tagged with "requirement".
func NewRequirementNote(id NoteID, content string, tags ...string) *MemoryNote {
	allTags := append([]string{"requirement"}, tags...)
	return NewMemoryNote(id, SourceTypeRequirement).
		WithRawContent(content).
		WithSummary(content).
		WithTags(allTags...).
		WithImportance(5)
}

// NewRetrospectiveNote creates a retrospective note for lessons learned.
// Retrospectives have medium-high importance (3) and are tagged with "retrospective".
func NewRetrospectiveNote(id NoteID, content string, tags ...string) *MemoryNote {
	allTags := append([]string{"retrospective", "lessons-learned"}, tags...)
	return NewMemoryNote(id, SourceTypeRetrospective).
		WithRawContent(content).
		WithSummary(content).
		WithTags(allTags...).
		WithImportance(3)
}

// NewSummaryNote creates a summary note that references source notes.
// Summaries have medium importance (3) and are tagged with "summary".
func NewSummaryNote(id NoteID, content string, sourceIDs []string, tags ...string) *MemoryNote {
	allTags := append([]string{"summary"}, tags...)
	var b strings.Builder
	b.WriteString("Summarizes notes: ")
	for i, srcID := range sourceIDs {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(srcID)
	}
	return NewMemoryNote(id, SourceTypeSummary).
		WithRawContent(content).
		WithSummary(content).
		WithContextDescription(b.String()).
		WithTags(allTags...).
		WithImportance(3)
}

// WithUserID sets the user ID for the note.
func (n *MemoryNote) WithUserID(userID string) *MemoryNote {
	n.UserID = userID
	return n
}

// WithSessionID sets the session ID for the note.
func (n *MemoryNote) WithSessionID(sessionID string) *MemoryNote {
	n.SessionID = sessionID
	return n
}

// WithTaskID sets the task ID for the note.
func (n *MemoryNote) WithTaskID(taskID string) *MemoryNote {
	n.TaskID = taskID
	return n
}

// WithRawContent sets the raw content for the note.
func (n *MemoryNote) WithRawContent(content string) *MemoryNote {
	n.RawContent = content
	n.UpdatedAt = time.Now()
	return n
}

// WithSummary sets the summary for the note.
func (n *MemoryNote) WithSummary(summary string) *MemoryNote {
	n.Summary = summary
	n.UpdatedAt = time.Now()
	return n
}

// WithContextDescription sets the context description for the note.
func (n *MemoryNote) WithContextDescription(desc string) *MemoryNote {
	n.ContextDescription = desc
	n.UpdatedAt = time.Now()
	return n
}

// WithKeywords sets the keywords for the note.
func (n *MemoryNote) WithKeywords(keywords ...string) *MemoryNote {
	n.Keywords = keywords
	n.UpdatedAt = time.Now()
	return n
}

// WithTags sets the tags for the note.
func (n *MemoryNote) WithTags(tags ...string) *MemoryNote {
	n.Tags = tags
	n.UpdatedAt = time.Now()
	return n
}

// WithImportance sets the importance score (1-5) for the note.
func (n *MemoryNote) WithImportance(importance int) *MemoryNote {
	if importance < 1 {
		importance = 1
	}
	if importance > 5 {
		importance = 5
	}
	n.Importance = importance
	n.UpdatedAt = time.Now()
	return n
}

// HasTag checks if the note has a specific tag.
func (n *MemoryNote) HasTag(tag string) bool {
	return slices.Contains(n.Tags, tag)
}

// HasKeyword checks if the note has a specific keyword.
func (n *MemoryNote) HasKeyword(keyword string) bool {
	return slices.Contains(n.Keywords, keyword)
}

// SearchableText returns the combined text used for semantic search.
// This is typically used to generate embeddings.
func (n *MemoryNote) SearchableText() string {
	text := n.RawContent
	if n.Summary != "" {
		text = text + " " + n.Summary
	}
	if n.ContextDescription != "" {
		text = text + " " + n.ContextDescription
	}
	return text
}
