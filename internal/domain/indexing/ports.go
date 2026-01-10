package indexing

import "context"

// FileWalker walks a directory tree and collects file information.
type FileWalker interface {
	// Walk traverses the given root directories and returns file information.
	// Paths matching ignore patterns are excluded.
	Walk(ctx context.Context, roots []string, ignore []string) ([]FileInfo, error)
}

// IndexStore persists and retrieves snapshots.
type IndexStore interface {
	// GetLatestSnapshot retrieves the most recent snapshot.
	// Returns an empty snapshot if none exists.
	GetLatestSnapshot(ctx context.Context) (Snapshot, error)
	// GetSnapshot retrieves a snapshot by ID.
	GetSnapshot(ctx context.Context, id SnapshotID) (Snapshot, error)
	// SaveSnapshot persists a snapshot.
	SaveSnapshot(ctx context.Context, snapshot Snapshot) error
}
