package outbound

import (
	"context"
	"errors"
	"strings"

	"github.com/andygeiss/cloud-native-utils/resource"
	"github.com/andygeiss/cloud-native-utils/slices"
	"github.com/andygeiss/go-agent/internal/domain/agent"
)

// ErrMemoryNoteNotFound is returned when a note is not found.
var ErrMemoryNoteNotFound = errors.New("memory note not found")

// MemoryStore persists memory notes using a generic resource.Access backend.
// Supports any backend: InMemoryAccess, JsonFileAccess, YamlFileAccess, SqliteAccess.
// Search is performed via basic text matching; for production use with embeddings,
// consider extending with a vector database.
type MemoryStore struct {
	access resource.Access[string, agent.MemoryNote]
}

// NewMemoryStore creates a MemoryStore with the given storage backend.
// Example backends:
//   - resource.NewInMemoryAccess[string, agent.MemoryNote]() - for testing
//   - resource.NewJsonFileAccess[string, agent.MemoryNote]("memory.json") - for file persistence
func NewMemoryStore(access resource.Access[string, agent.MemoryNote]) *MemoryStore {
	return &MemoryStore{access: access}
}

// NewInMemoryMemoryStore creates a MemoryStore backed by in-memory storage.
// Useful for testing or when persistence is not required.
func NewInMemoryMemoryStore() *MemoryStore {
	return NewMemoryStore(resource.NewInMemoryAccess[string, agent.MemoryNote]())
}

// NewJsonFileMemoryStore creates a MemoryStore backed by a JSON file.
// The file is created if it does not exist.
func NewJsonFileMemoryStore(path string) *MemoryStore {
	return NewMemoryStore(resource.NewJsonFileAccess[string, agent.MemoryNote](path))
}

// Write stores a new memory note.
// Creates a new record if none exists, or updates the existing one.
func (s *MemoryStore) Write(ctx context.Context, note *agent.MemoryNote) error {
	key := string(note.ID)

	// Try to create new note first (handles non-existent files)
	err := s.access.Create(ctx, key, *note)
	if err != nil && err.Error() == resource.ErrorResourceAlreadyExists {
		// Update existing note
		return s.access.Update(ctx, key, *note)
	}
	return err
}

// Search retrieves notes matching the query and filters.
// This implementation performs basic text matching on SearchableText.
// For production use with embeddings, extend or wrap this with vector similarity search.
func (s *MemoryStore) Search(ctx context.Context, query string, limit int, opts *agent.MemorySearchOptions) ([]*agent.MemoryNote, error) {
	allNotes, err := s.access.ReadAll(ctx)
	if err != nil {
		return nil, err
	}

	queryLower := strings.ToLower(query)
	var results []*agent.MemoryNote

	for i := range allNotes {
		note := &allNotes[i]

		if !matchesFilters(note, opts) {
			continue
		}

		if matchesQuery(note, queryLower) {
			noteCopy := *note
			results = append(results, &noteCopy)
		}
	}

	sortByImportance(results)
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// matchesFilters checks if a note passes all filter criteria.
func matchesFilters(note *agent.MemoryNote, opts *agent.MemorySearchOptions) bool {
	if opts == nil {
		return true
	}
	return matchesImportance(note, opts) &&
		matchesScope(note, opts) &&
		matchesSourceTypes(note, opts) &&
		matchesTags(note, opts)
}

// matchesImportance checks if note meets minimum importance requirement.
func matchesImportance(note *agent.MemoryNote, opts *agent.MemorySearchOptions) bool {
	return opts.MinImportance <= 0 || note.Importance >= opts.MinImportance
}

// matchesScope checks if note matches user/session/task scope filters.
func matchesScope(note *agent.MemoryNote, opts *agent.MemorySearchOptions) bool {
	if opts.UserID != "" && note.UserID != opts.UserID {
		return false
	}
	if opts.SessionID != "" && note.SessionID != opts.SessionID {
		return false
	}
	if opts.TaskID != "" && note.TaskID != opts.TaskID {
		return false
	}
	return true
}

// matchesSourceTypes checks if note has one of the required source types.
func matchesSourceTypes(note *agent.MemoryNote, opts *agent.MemorySearchOptions) bool {
	return len(opts.SourceTypes) == 0 || hasAnySourceType(note, opts.SourceTypes)
}

// matchesTags checks if note has any of the required tags.
func matchesTags(note *agent.MemoryNote, opts *agent.MemorySearchOptions) bool {
	return len(opts.Tags) == 0 || hasAnyTag(note, opts.Tags)
}

// hasAnySourceType checks if the note's source type matches any of the specified types.
func hasAnySourceType(note *agent.MemoryNote, sourceTypes []agent.SourceType) bool {
	return slices.Contains(sourceTypes, note.SourceType)
}

// matchesQuery checks if the note content matches the search query.
func matchesQuery(note *agent.MemoryNote, queryLower string) bool {
	searchText := strings.ToLower(note.SearchableText())
	return strings.Contains(searchText, queryLower) || matchesKeywords(note, queryLower)
}

// Get retrieves a specific note by ID.
// Returns ErrMemoryNoteNotFound if the note is not found.
func (s *MemoryStore) Get(ctx context.Context, id agent.NoteID) (*agent.MemoryNote, error) {
	key := string(id)
	note, err := s.access.Read(ctx, key)
	if err != nil {
		if err.Error() == resource.ErrorResourceNotFound {
			return nil, ErrMemoryNoteNotFound
		}
		return nil, err
	}
	return note, nil
}

// Delete removes a note by ID.
// Returns nil if the note does not exist.
func (s *MemoryStore) Delete(ctx context.Context, id agent.NoteID) error {
	key := string(id)
	err := s.access.Delete(ctx, key)
	if err != nil && err.Error() == resource.ErrorResourceNotFound {
		return nil // Not an error if note doesn't exist
	}
	return err
}

// hasAnyTag checks if the note has any of the specified tags.
func hasAnyTag(note *agent.MemoryNote, tags []string) bool {
	return slices.ContainsAny(note.Tags, tags)
}

// matchesKeywords checks if the query matches any of the note's keywords.
func matchesKeywords(note *agent.MemoryNote, queryLower string) bool {
	queryWords := strings.Fields(queryLower)
	for _, keyword := range note.Keywords {
		keywordLower := strings.ToLower(keyword)
		for _, queryWord := range queryWords {
			if strings.Contains(keywordLower, queryWord) || strings.Contains(queryWord, keywordLower) {
				return true
			}
		}
	}
	return false
}

// sortByImportance sorts notes by importance in descending order.
// Uses a simple insertion sort for small lists.
func sortByImportance(notes []*agent.MemoryNote) {
	for i := 1; i < len(notes); i++ {
		key := notes[i]
		j := i - 1
		for j >= 0 && notes[j].Importance < key.Importance {
			notes[j+1] = notes[j]
			j--
		}
		notes[j+1] = key
	}
}
