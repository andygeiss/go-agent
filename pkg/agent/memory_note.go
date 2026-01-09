package agent

import (
	"time"

	"github.com/andygeiss/cloud-native-utils/slices"
)

// NoteID is the unique identifier for a memory note.
// It is a value object that ensures type safety for note references.
type NoteID string

// SourceType categorizes what created a memory note.
type SourceType string

// Standard source types for memory notes.
const (
	SourceTypeFact        SourceType = "fact"
	SourceTypePlanStep    SourceType = "plan_step"
	SourceTypePreference  SourceType = "preference"
	SourceTypeSummary     SourceType = "summary"
	SourceTypeToolResult  SourceType = "tool_result"
	SourceTypeUserMessage SourceType = "user_message"
)

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
