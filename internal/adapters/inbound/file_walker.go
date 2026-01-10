// Package inbound provides inbound adapters for the application.
package inbound

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/andygeiss/go-agent/internal/domain/indexing"
)

// FSWalker walks the file system to collect file information.
type FSWalker struct {
	computeHash bool // Whether to compute file hashes
}

// NewFSWalker creates a new FSWalker.
func NewFSWalker() *FSWalker {
	return &FSWalker{
		computeHash: false,
	}
}

// Walk traverses the given root directories and returns file information.
// Paths matching ignore patterns are excluded.
func (w *FSWalker) Walk(ctx context.Context, roots []string, ignore []string) ([]indexing.FileInfo, error) {
	var files []indexing.FileInfo

	for _, root := range roots {
		walked, err := w.walkRoot(ctx, root, ignore)
		if err != nil {
			return nil, err
		}
		files = append(files, walked...)
	}

	return files, nil
}

// WalkSingleFile returns FileInfo for a single file.
// Useful for getting info about a specific file without walking directories.
func (w *FSWalker) WalkSingleFile(path string) (indexing.FileInfo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return indexing.FileInfo{}, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return indexing.FileInfo{}, err
	}

	fileInfo := indexing.NewFileInfo(absPath, info.ModTime(), info.Size())

	if w.computeHash {
		hash, err := indexing.HashFile(absPath)
		if err == nil {
			fileInfo = fileInfo.WithHash(hash)
		}
	}

	return fileInfo, nil
}

// WithHash enables computing file content hashes during walks.
func (w *FSWalker) WithHash(enabled bool) *FSWalker {
	w.computeHash = enabled
	return w
}

// createFileInfo creates a FileInfo from a directory entry.
func (w *FSWalker) createFileInfo(path string, d fs.DirEntry) (indexing.FileInfo, error) {
	info, err := d.Info()
	if err != nil {
		return indexing.FileInfo{}, err
	}

	fileInfo := indexing.NewFileInfo(path, info.ModTime(), info.Size())

	if w.computeHash {
		if hash, err := indexing.HashFile(path); err == nil {
			fileInfo = fileInfo.WithHash(hash)
		}
	}

	return fileInfo, nil
}

// matchesPattern checks if a path matches a single pattern.
func matchesPattern(path, pattern string) bool {
	// Check if the path contains the pattern (simple substring match)
	if strings.Contains(path, pattern) {
		return true
	}

	// Check if the base name matches the pattern
	base := filepath.Base(path)
	if base == pattern {
		return true
	}

	// Try glob matching
	if matched, _ := filepath.Match(pattern, base); matched {
		return true
	}

	return false
}

// shouldIgnore checks if a path matches any ignore pattern.
func shouldIgnore(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchesPattern(path, pattern) {
			return true
		}
	}
	return false
}

// visitPath processes a single path during the walk.
func (w *FSWalker) visitPath(ctx context.Context, path string, d fs.DirEntry, ignore []string, files *[]indexing.FileInfo) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if d.IsDir() {
		if shouldIgnore(path, ignore) {
			return filepath.SkipDir
		}
		return nil
	}

	if shouldIgnore(path, ignore) {
		return nil
	}

	fileInfo, err := w.createFileInfo(path, d)
	if err != nil {
		return err
	}

	*files = append(*files, fileInfo)
	return nil
}

// walkRoot walks a single root directory.
func (w *FSWalker) walkRoot(ctx context.Context, root string, ignore []string) ([]indexing.FileInfo, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	var files []indexing.FileInfo
	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		return w.visitPath(ctx, path, d, ignore, &files)
	})

	return files, err
}
