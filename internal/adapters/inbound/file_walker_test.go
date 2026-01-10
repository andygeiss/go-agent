package inbound_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/adapters/inbound"
)

func TestFSWalker_Walk(t *testing.T) {
	// Create a temp directory with test files
	tempDir := t.TempDir()

	// Create some test files
	file1 := filepath.Join(tempDir, "file1.go")
	file2 := filepath.Join(tempDir, "file2.txt")
	subDir := filepath.Join(tempDir, "subdir")
	file3 := filepath.Join(subDir, "file3.go")

	_ = os.WriteFile(file1, []byte("content1"), 0600)
	_ = os.WriteFile(file2, []byte("content2"), 0600)
	_ = os.MkdirAll(subDir, 0750)
	_ = os.WriteFile(file3, []byte("content3"), 0600)

	walker := inbound.NewFSWalker()
	files, err := walker.Walk(context.Background(), []string{tempDir}, nil)

	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "files count must be 3", len(files), 3)
}

func TestFSWalker_Walk_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "file1.go")
	_ = os.WriteFile(file1, []byte("content1"), 0600)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	walker := inbound.NewFSWalker()
	_, err := walker.Walk(ctx, []string{tempDir}, nil)

	assert.That(t, "error must not be nil", err != nil, true)
}

func TestFSWalker_Walk_IgnorePatterns(t *testing.T) {
	tests := []struct { //nolint:govet // anonymous struct, alignment acceptable
		ignorePattern []string
		name          string
		expected      int
	}{
		{
			name:          "no patterns includes all files",
			ignorePattern: nil,
			expected:      3, // .go, .txt, .log
		},
		{
			name:          "ignore log files",
			ignorePattern: []string{"*.log"},
			expected:      2, // excludes .log file
		},
		{
			name:          "ignore txt and log",
			ignorePattern: []string{"*.txt", "*.log"},
			expected:      1, // only .go remains
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			_ = os.WriteFile(filepath.Join(tempDir, "file.go"), []byte("go"), 0600)
			_ = os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("txt"), 0600)
			_ = os.WriteFile(filepath.Join(tempDir, "file.log"), []byte("log"), 0600)

			walker := inbound.NewFSWalker()
			files, err := walker.Walk(context.Background(), []string{tempDir}, tt.ignorePattern)

			assert.That(t, "error must be nil", err == nil, true)
			assert.That(t, "file count must match", len(files), tt.expected)
		})
	}
}

func TestFSWalker_Walk_WithHash(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "file1.go")
	_ = os.WriteFile(file1, []byte("content1"), 0600)

	walker := inbound.NewFSWalker().WithHash(true)
	files, err := walker.Walk(context.Background(), []string{tempDir}, nil)

	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "files count must be 1", len(files), 1)
	assert.That(t, "hash must not be empty", files[0].Hash != "", true)
}

func TestFSWalker_Walk_WithIgnore(t *testing.T) {
	// Create a temp directory with test files
	tempDir := t.TempDir()

	// Create some test files
	file1 := filepath.Join(tempDir, "file1.go")
	file2 := filepath.Join(tempDir, "file2.txt")
	ignoredDir := filepath.Join(tempDir, "node_modules")
	ignoredFile := filepath.Join(ignoredDir, "module.js")

	_ = os.WriteFile(file1, []byte("content1"), 0600)
	_ = os.WriteFile(file2, []byte("content2"), 0600)
	_ = os.MkdirAll(ignoredDir, 0750)
	_ = os.WriteFile(ignoredFile, []byte("ignored"), 0600)

	walker := inbound.NewFSWalker()
	files, err := walker.Walk(context.Background(), []string{tempDir}, []string{"node_modules"})

	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "files count must be 2", len(files), 2)
}

func TestFSWalker_WalkSingleFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "single.go")
	_ = os.WriteFile(filePath, []byte("single file content"), 0600)

	walker := inbound.NewFSWalker()
	info, err := walker.WalkSingleFile(filePath)

	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "size must be 19", info.Size, int64(19))
}

func TestFSWalker_WalkSingleFile_NotFound(t *testing.T) {
	walker := inbound.NewFSWalker()
	_, err := walker.WalkSingleFile("/nonexistent/path/file.go")

	assert.That(t, "error must not be nil", err != nil, true)
}

func TestFSWalker_WalkSingleFile_WithHash(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "single.go")
	_ = os.WriteFile(filePath, []byte("single file content"), 0600)

	walker := inbound.NewFSWalker().WithHash(true)
	info, err := walker.WalkSingleFile(filePath)

	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "hash must not be empty", info.Hash != "", true)
}
