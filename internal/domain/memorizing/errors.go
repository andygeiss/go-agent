package memorizing

import "errors"

// Sentinel errors for memory service validation.
var (
	ErrNoteNil     = errors.New("note cannot be nil")
	ErrNoteIDEmpty = errors.New("note ID cannot be empty")
)
