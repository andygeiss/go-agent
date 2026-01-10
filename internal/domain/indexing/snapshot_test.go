package indexing_test

import (
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/indexing"
)

func Test_NewFileInfo_Should_SetAllFields(t *testing.T) {
	// Arrange
	modTime := time.Now()

	// Act
	fi := indexing.NewFileInfo("/path/to/file.go", modTime, 1024)

	// Assert
	assert.That(t, "path must match", fi.Path, "/path/to/file.go")
	assert.That(t, "modTime must match", fi.ModTime, modTime)
	assert.That(t, "size must match", fi.Size, int64(1024))
	assert.That(t, "hash must be empty", fi.Hash, "")
}

func Test_FileInfo_WithHash_Should_SetHash(t *testing.T) {
	// Arrange
	fi := indexing.NewFileInfo("/path/to/file.go", time.Now(), 1024)

	// Act
	fi = fi.WithHash("abc123")

	// Assert
	assert.That(t, "hash must match", fi.Hash, "abc123")
}

func Test_NewSnapshot_Should_SetIDAndFiles(t *testing.T) {
	// Arrange
	files := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
		indexing.NewFileInfo("/path/to/file2.go", time.Now(), 200),
	}

	// Act
	snapshot := indexing.NewSnapshot("snap-1", files)

	// Assert
	assert.That(t, "snapshot ID must match", string(snapshot.ID), "snap-1")
	assert.That(t, "files length must be 2", len(snapshot.Files), 2)
	assert.That(t, "createdAt must not be zero", snapshot.CreatedAt.IsZero(), false)
}

func Test_Snapshot_FileCount_Should_ReturnCorrectCount(t *testing.T) {
	// Arrange
	files := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
		indexing.NewFileInfo("/path/to/file2.go", time.Now(), 200),
		indexing.NewFileInfo("/path/to/file3.go", time.Now(), 300),
	}
	snapshot := indexing.NewSnapshot("snap-1", files)

	// Act
	count := snapshot.FileCount()

	// Assert
	assert.That(t, "file count must be 3", count, 3)
}

func Test_Snapshot_GetFileByPath_With_ExistingPath_Should_ReturnFile(t *testing.T) {
	// Arrange
	files := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
		indexing.NewFileInfo("/path/to/file2.go", time.Now(), 200),
	}
	snapshot := indexing.NewSnapshot("snap-1", files)

	// Act
	result := snapshot.GetFileByPath("/path/to/file1.go")

	// Assert
	assert.That(t, "result must not be nil", result != nil, true)
	assert.That(t, "path must match", result.Path, "/path/to/file1.go")
}

func Test_Snapshot_GetFileByPath_With_NonexistentPath_Should_ReturnNil(t *testing.T) {
	// Arrange
	files := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
		indexing.NewFileInfo("/path/to/file2.go", time.Now(), 200),
	}
	snapshot := indexing.NewSnapshot("snap-1", files)

	// Act
	result := snapshot.GetFileByPath("/path/to/nonexistent.go")

	// Assert
	assert.That(t, "result must be nil", result == nil, true)
}
