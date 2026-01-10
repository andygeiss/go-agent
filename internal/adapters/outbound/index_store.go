// Package outbound provides outbound adapters for the application.
package outbound

import (
	"context"
	"errors"

	"github.com/andygeiss/cloud-native-utils/resource"
	"github.com/andygeiss/go-agent/internal/domain/indexing"
)

const latestSnapshotKey = "_latest_"

// ErrSnapshotNotFound is returned when a snapshot is not found.
var ErrSnapshotNotFound = errors.New("snapshot not found")

// IndexStore persists snapshots to a JSON file.
type IndexStore struct {
	access resource.Access[string, indexing.Snapshot]
}

// NewIndexStore creates a new IndexStore with the given file path.
func NewIndexStore(path string) *IndexStore {
	return &IndexStore{
		access: resource.NewJsonFileAccess[string, indexing.Snapshot](path),
	}
}

// NewInMemoryIndexStore creates a new in-memory store for testing.
func NewInMemoryIndexStore() *IndexStore {
	return &IndexStore{
		access: resource.NewInMemoryAccess[string, indexing.Snapshot](),
	}
}

// GetLatestSnapshot retrieves the most recent snapshot.
// Returns an empty snapshot if none exists.
func (s *IndexStore) GetLatestSnapshot(ctx context.Context) (indexing.Snapshot, error) {
	snapshot, err := s.access.Read(ctx, latestSnapshotKey)
	if err != nil {
		if err.Error() == resource.ErrorResourceNotFound {
			return indexing.Snapshot{}, nil
		}
		return indexing.Snapshot{}, err
	}
	if snapshot == nil {
		return indexing.Snapshot{}, nil
	}
	return *snapshot, nil
}

// GetSnapshot retrieves a snapshot by ID.
func (s *IndexStore) GetSnapshot(ctx context.Context, id indexing.SnapshotID) (indexing.Snapshot, error) {
	snapshot, err := s.access.Read(ctx, string(id))
	if err != nil {
		if err.Error() == resource.ErrorResourceNotFound {
			return indexing.Snapshot{}, ErrSnapshotNotFound
		}
		return indexing.Snapshot{}, err
	}
	if snapshot == nil {
		return indexing.Snapshot{}, ErrSnapshotNotFound
	}
	return *snapshot, nil
}

// SaveSnapshot persists a snapshot and updates the latest pointer.
func (s *IndexStore) SaveSnapshot(ctx context.Context, snapshot indexing.Snapshot) error {
	key := string(snapshot.ID)

	// Save the snapshot by ID
	err := s.access.Create(ctx, key, snapshot)
	if err != nil && err.Error() == resource.ErrorResourceAlreadyExists {
		err = s.access.Update(ctx, key, snapshot)
	}
	if err != nil {
		return err
	}

	// Update the latest snapshot pointer
	err = s.access.Create(ctx, latestSnapshotKey, snapshot)
	if err != nil && err.Error() == resource.ErrorResourceAlreadyExists {
		err = s.access.Update(ctx, latestSnapshotKey, snapshot)
	}

	return err
}
