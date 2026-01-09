package memorizing

import "errors"

// Sentinel errors for memory service validation (alphabetically sorted).
var (
	ErrNoteIDEmpty = errors.New("note ID cannot be empty")
	ErrNoteNil     = errors.New("note cannot be nil")
)
