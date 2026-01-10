// Package indexing provides file system indexing capabilities for the agent.
// It enables tracking file changes, scanning directories, and comparing snapshots.
package indexing

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"time"
)

// SnapshotID is the unique identifier for a snapshot.
type SnapshotID string

// FileInfo represents metadata about a single file in the index.
type FileInfo struct {
	Hash    string    // SHA-256 hash of file contents (optional, for change detection)
	ModTime time.Time // Last modification time
	Path    string    // Absolute or relative file path
	Size    int64     // File size in bytes
}

// NewFileInfo creates a new FileInfo with the given path and metadata.
func NewFileInfo(path string, modTime time.Time, size int64) FileInfo {
	return FileInfo{
		ModTime: modTime,
		Path:    path,
		Size:    size,
	}
}

// WithHash sets the content hash on the FileInfo.
func (f FileInfo) WithHash(hash string) FileInfo {
	f.Hash = hash
	return f
}

// Snapshot represents a point-in-time capture of file system state.
type Snapshot struct {
	CreatedAt time.Time  // When the snapshot was created
	ID        SnapshotID // Unique identifier
	Files     []FileInfo // List of files in the snapshot
}

// NewSnapshot creates a new Snapshot with the given ID and files.
func NewSnapshot(id SnapshotID, files []FileInfo) Snapshot {
	return Snapshot{
		CreatedAt: time.Now(),
		Files:     files,
		ID:        id,
	}
}

// FileCount returns the number of files in the snapshot.
func (s Snapshot) FileCount() int {
	return len(s.Files)
}

// GetFileByPath returns the FileInfo for the given path, or nil if not found.
func (s Snapshot) GetFileByPath(path string) *FileInfo {
	for i := range s.Files {
		if s.Files[i].Path == path {
			return &s.Files[i]
		}
	}
	return nil
}

// DiffResult represents the difference between two snapshots.
type DiffResult struct {
	Added   []FileInfo // Files in the newer snapshot but not the older
	Changed []FileInfo // Files in both snapshots but with different hash/size
	Removed []FileInfo // Files in the older snapshot but not the newer
}

// HashFile computes the SHA-256 hash of a file's contents.
func HashFile(path string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // path is validated by caller
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
