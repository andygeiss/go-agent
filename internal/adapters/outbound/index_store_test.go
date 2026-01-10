package outbound_test

import (
	"context"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/internal/domain/indexing"
)

func TestIndexStore_GetLatestSnapshot_Empty(t *testing.T) {
	store := outbound.NewInMemoryIndexStore()
	ctx := context.Background()

	latest, err := store.GetLatestSnapshot(ctx)
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "latest ID must be empty", string(latest.ID), "")
}

func TestIndexStore_GetSnapshot_NotFound(t *testing.T) {
	store := outbound.NewInMemoryIndexStore()
	ctx := context.Background()

	_, err := store.GetSnapshot(ctx, "nonexistent")
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "error must be ErrSnapshotNotFound", err, outbound.ErrSnapshotNotFound)
}

func TestIndexStore_SaveAndGetSnapshot(t *testing.T) {
	store := outbound.NewInMemoryIndexStore()
	ctx := context.Background()

	files := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
		indexing.NewFileInfo("/path/to/file2.go", time.Now(), 200),
	}
	snapshot := indexing.NewSnapshot("snap-1", files)

	err := store.SaveSnapshot(ctx, snapshot)
	assert.That(t, "save error must be nil", err == nil, true)

	retrieved, err := store.GetSnapshot(ctx, "snap-1")
	assert.That(t, "get error must be nil", err == nil, true)
	assert.That(t, "snapshot ID must match", string(retrieved.ID), "snap-1")
	assert.That(t, "files count must be 2", len(retrieved.Files), 2)
}

func TestIndexStore_SaveSnapshot_GetLatestSnapshot(t *testing.T) {
	store := outbound.NewInMemoryIndexStore()
	ctx := context.Background()

	// First snapshot
	files1 := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
	}
	snapshot1 := indexing.NewSnapshot("snap-1", files1)
	_ = store.SaveSnapshot(ctx, snapshot1)

	// Second snapshot (should become latest)
	files2 := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file2.go", time.Now(), 200),
	}
	snapshot2 := indexing.NewSnapshot("snap-2", files2)
	_ = store.SaveSnapshot(ctx, snapshot2)

	latest, err := store.GetLatestSnapshot(ctx)
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "latest ID must be snap-2", string(latest.ID), "snap-2")
}

func TestIndexStore_UpdateSnapshot(t *testing.T) {
	store := outbound.NewInMemoryIndexStore()
	ctx := context.Background()

	// Create initial snapshot
	files1 := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
	}
	snapshot1 := indexing.NewSnapshot("snap-1", files1)
	_ = store.SaveSnapshot(ctx, snapshot1)

	// Update with same ID
	files2 := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
		indexing.NewFileInfo("/path/to/file2.go", time.Now(), 200),
	}
	snapshot2 := indexing.NewSnapshot("snap-1", files2)
	err := store.SaveSnapshot(ctx, snapshot2)
	assert.That(t, "update error must be nil", err == nil, true)

	// Verify update
	retrieved, err := store.GetSnapshot(ctx, "snap-1")
	assert.That(t, "get error must be nil", err == nil, true)
	assert.That(t, "files count must be 2", len(retrieved.Files), 2)
}
