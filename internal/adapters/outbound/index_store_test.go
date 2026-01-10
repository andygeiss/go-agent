package outbound_test

import (
	"context"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/adapters/outbound"
	"github.com/andygeiss/go-agent/internal/domain/indexing"
)

func Test_IndexStore_GetLatestSnapshot_With_EmptyStore_Should_ReturnEmptySnapshot(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryIndexStore()
	ctx := context.Background()

	// Act
	latest, err := store.GetLatestSnapshot(ctx)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "latest ID must be empty", string(latest.ID), "")
}

func Test_IndexStore_GetSnapshot_With_NonexistentID_Should_ReturnError(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryIndexStore()
	ctx := context.Background()

	// Act
	_, err := store.GetSnapshot(ctx, "nonexistent")

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
	assert.That(t, "error must be ErrSnapshotNotFound", err, outbound.ErrSnapshotNotFound)
}

func Test_IndexStore_SaveSnapshot_Should_PersistAndRetrieve(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryIndexStore()
	ctx := context.Background()
	files := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
		indexing.NewFileInfo("/path/to/file2.go", time.Now(), 200),
	}
	snapshot := indexing.NewSnapshot("snap-1", files)

	// Act
	err := store.SaveSnapshot(ctx, snapshot)

	// Assert
	assert.That(t, "save error must be nil", err == nil, true)
	retrieved, err := store.GetSnapshot(ctx, "snap-1")
	assert.That(t, "get error must be nil", err == nil, true)
	assert.That(t, "snapshot ID must match", string(retrieved.ID), "snap-1")
	assert.That(t, "files count must be 2", len(retrieved.Files), 2)
}

func Test_IndexStore_SaveSnapshot_Should_UpdateLatestSnapshot(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryIndexStore()
	ctx := context.Background()

	files1 := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
	}
	snapshot1 := indexing.NewSnapshot("snap-1", files1)
	_ = store.SaveSnapshot(ctx, snapshot1)

	files2 := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file2.go", time.Now(), 200),
	}
	snapshot2 := indexing.NewSnapshot("snap-2", files2)

	// Act
	_ = store.SaveSnapshot(ctx, snapshot2)

	// Assert
	latest, err := store.GetLatestSnapshot(ctx)
	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "latest ID must be snap-2", string(latest.ID), "snap-2")
}

func Test_IndexStore_SaveSnapshot_With_ExistingID_Should_UpdateSnapshot(t *testing.T) {
	// Arrange
	store := outbound.NewInMemoryIndexStore()
	ctx := context.Background()

	files1 := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
	}
	snapshot1 := indexing.NewSnapshot("snap-1", files1)
	_ = store.SaveSnapshot(ctx, snapshot1)

	files2 := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
		indexing.NewFileInfo("/path/to/file2.go", time.Now(), 200),
	}
	snapshot2 := indexing.NewSnapshot("snap-1", files2)

	// Act
	err := store.SaveSnapshot(ctx, snapshot2)

	// Assert
	assert.That(t, "update error must be nil", err == nil, true)
	retrieved, err := store.GetSnapshot(ctx, "snap-1")
	assert.That(t, "get error must be nil", err == nil, true)
	assert.That(t, "files count must be 2", len(retrieved.Files), 2)
}
