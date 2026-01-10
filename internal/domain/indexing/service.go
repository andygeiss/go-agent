package indexing

import (
	"context"
	"errors"
	"time"

	"github.com/andygeiss/cloud-native-utils/slices"
)

// Sentinel errors for the indexing service.
var (
	ErrSnapshotNotFound = errors.New("snapshot not found")
)

// Service provides file system indexing use cases.
type Service struct {
	idGen  func() string
	store  IndexStore
	walker FileWalker
}

// NewService creates a new indexing service.
func NewService(walker FileWalker, store IndexStore, idGenerator func() string) *Service {
	return &Service{
		idGen:  idGenerator,
		store:  store,
		walker: walker,
	}
}

// ChangedSince returns files from the latest snapshot that were modified after the given time.
func (s *Service) ChangedSince(ctx context.Context, since time.Time) ([]FileInfo, error) {
	snapshot, err := s.store.GetLatestSnapshot(ctx)
	if err != nil {
		return nil, err
	}

	changed := slices.Filter(snapshot.Files, func(f FileInfo) bool {
		return f.ModTime.After(since)
	})

	return changed, nil
}

// DiffSnapshots compares two snapshots and returns the differences.
// fromID is the older snapshot, toID is the newer snapshot.
func (s *Service) DiffSnapshots(ctx context.Context, fromID, toID SnapshotID) (DiffResult, error) {
	fromSnapshot, err := s.store.GetSnapshot(ctx, fromID)
	if err != nil {
		return DiffResult{}, err
	}

	toSnapshot, err := s.store.GetSnapshot(ctx, toID)
	if err != nil {
		return DiffResult{}, err
	}

	return diffSnapshots(fromSnapshot, toSnapshot), nil
}

// Scan walks the given directories, builds a snapshot, and persists it.
// Returns the created snapshot.
func (s *Service) Scan(ctx context.Context, roots []string, ignore []string) (Snapshot, error) {
	files, err := s.walker.Walk(ctx, roots, ignore)
	if err != nil {
		return Snapshot{}, err
	}

	snapshot := NewSnapshot(SnapshotID(s.idGen()), files)

	if err := s.store.SaveSnapshot(ctx, snapshot); err != nil {
		return Snapshot{}, err
	}

	return snapshot, nil
}

// diffSnapshots computes the diff between two snapshots.
func diffSnapshots(from, to Snapshot) DiffResult {
	// Build path maps for quick lookup
	fromMap := make(map[string]FileInfo, len(from.Files))
	for _, f := range from.Files {
		fromMap[f.Path] = f
	}

	toMap := make(map[string]FileInfo, len(to.Files))
	for _, f := range to.Files {
		toMap[f.Path] = f
	}

	var result DiffResult

	// Find added and changed files (in 'to' but different or missing in 'from')
	for _, toFile := range to.Files {
		fromFile, exists := fromMap[toFile.Path]
		if !exists {
			result.Added = append(result.Added, toFile)
		} else if isFileChanged(fromFile, toFile) {
			result.Changed = append(result.Changed, toFile)
		}
	}

	// Find removed files (in 'from' but not in 'to')
	for _, fromFile := range from.Files {
		if _, exists := toMap[fromFile.Path]; !exists {
			result.Removed = append(result.Removed, fromFile)
		}
	}

	return result
}

// isFileChanged determines if a file has changed between two versions.
// Compares hash first (if available), then falls back to size and modtime.
func isFileChanged(a, b FileInfo) bool {
	// If both have hashes, compare them
	if a.Hash != "" && b.Hash != "" {
		return a.Hash != b.Hash
	}
	// Fall back to size + modtime comparison
	return a.Size != b.Size || !a.ModTime.Equal(b.ModTime)
}
