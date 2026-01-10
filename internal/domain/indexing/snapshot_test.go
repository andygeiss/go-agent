package indexing_test

import (
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/indexing"
)

func TestNewFileInfo(t *testing.T) {
	modTime := time.Now()
	fi := indexing.NewFileInfo("/path/to/file.go", modTime, 1024)

	assert.That(t, "path must match", fi.Path, "/path/to/file.go")
	assert.That(t, "modTime must match", fi.ModTime, modTime)
	assert.That(t, "size must match", fi.Size, int64(1024))
	assert.That(t, "hash must be empty", fi.Hash, "")
}

func TestFileInfo_WithHash(t *testing.T) {
	fi := indexing.NewFileInfo("/path/to/file.go", time.Now(), 1024).
		WithHash("abc123")

	assert.That(t, "hash must match", fi.Hash, "abc123")
}

func TestNewSnapshot(t *testing.T) {
	files := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
		indexing.NewFileInfo("/path/to/file2.go", time.Now(), 200),
	}
	snapshot := indexing.NewSnapshot("snap-1", files)

	assert.That(t, "snapshot ID must match", string(snapshot.ID), "snap-1")
	assert.That(t, "files length must be 2", len(snapshot.Files), 2)
	assert.That(t, "createdAt must not be zero", snapshot.CreatedAt.IsZero(), false)
}

func TestSnapshot_FileCount(t *testing.T) {
	files := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
		indexing.NewFileInfo("/path/to/file2.go", time.Now(), 200),
		indexing.NewFileInfo("/path/to/file3.go", time.Now(), 300),
	}
	snapshot := indexing.NewSnapshot("snap-1", files)

	assert.That(t, "file count must be 3", snapshot.FileCount(), 3)
}

func TestSnapshot_GetFileByPath(t *testing.T) {
	files := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", time.Now(), 100),
		indexing.NewFileInfo("/path/to/file2.go", time.Now(), 200),
	}
	snapshot := indexing.NewSnapshot("snap-1", files)

	tests := []struct {
		name      string
		path      string
		expectNil bool
	}{
		{
			name:      "file exists",
			path:      "/path/to/file1.go",
			expectNil: false,
		},
		{
			name:      "file does not exist",
			path:      "/path/to/nonexistent.go",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := snapshot.GetFileByPath(tt.path)
			if tt.expectNil {
				assert.That(t, "result must be nil", result == nil, true)
			} else {
				assert.That(t, "result must not be nil", result != nil, true)
				assert.That(t, "path must match", result.Path, tt.path)
			}
		})
	}
}
