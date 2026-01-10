package indexing_test

import (
	"context"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/indexing"
)

// mockFileWalker is a test double for FileWalker.
type mockFileWalker struct {
	err   error
	files []indexing.FileInfo
}

func (m *mockFileWalker) Walk(_ context.Context, _ []string, _ []string) ([]indexing.FileInfo, error) {
	return m.files, m.err
}

// mockIndexStore is a test double for IndexStore.
type mockIndexStore struct {
	err       error
	snapshots map[indexing.SnapshotID]indexing.Snapshot
	latest    indexing.Snapshot
}

func newMockIndexStore() *mockIndexStore {
	return &mockIndexStore{
		snapshots: make(map[indexing.SnapshotID]indexing.Snapshot),
	}
}

func (m *mockIndexStore) SaveSnapshot(_ context.Context, snapshot indexing.Snapshot) error {
	if m.err != nil {
		return m.err
	}
	m.snapshots[snapshot.ID] = snapshot
	m.latest = snapshot
	return nil
}

func (m *mockIndexStore) GetLatestSnapshot(_ context.Context) (indexing.Snapshot, error) {
	if m.err != nil {
		return indexing.Snapshot{}, m.err
	}
	return m.latest, nil
}

func (m *mockIndexStore) GetSnapshot(_ context.Context, id indexing.SnapshotID) (indexing.Snapshot, error) {
	if m.err != nil {
		return indexing.Snapshot{}, m.err
	}
	if snapshot, ok := m.snapshots[id]; ok {
		return snapshot, nil
	}
	return indexing.Snapshot{}, indexing.ErrSnapshotNotFound
}

func TestService_Scan(t *testing.T) {
	now := time.Now()
	files := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", now, 100),
		indexing.NewFileInfo("/path/to/file2.go", now, 200),
	}

	walker := &mockFileWalker{files: files}
	store := newMockIndexStore()
	idCounter := 0
	idGen := func() string {
		idCounter++
		return "snap-" + string(rune('0'+idCounter))
	}

	svc := indexing.NewService(walker, store, idGen)

	snapshot, err := svc.Scan(context.Background(), []string{"/path"}, nil)

	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "file count must be 2", snapshot.FileCount(), 2)
	assert.That(t, "snapshot ID must match", string(snapshot.ID), "snap-1")
}

func TestService_ChangedSince(t *testing.T) {
	baseTime := time.Now().Add(-2 * time.Hour)
	files := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/old.go", baseTime.Add(-1*time.Hour), 100),    // 3 hours ago
		indexing.NewFileInfo("/path/to/new1.go", baseTime.Add(30*time.Minute), 200), // 1.5 hours ago
		indexing.NewFileInfo("/path/to/new2.go", baseTime.Add(1*time.Hour), 300),    // 1 hour ago
	}

	store := newMockIndexStore()
	store.latest = indexing.NewSnapshot("snap-1", files)
	store.latest.Files = files // Ensure files are set

	walker := &mockFileWalker{}
	svc := indexing.NewService(walker, store, func() string { return "id" })

	// Get files changed in the last 2 hours
	sinceTime := baseTime
	changed, err := svc.ChangedSince(context.Background(), sinceTime)

	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "changed count must be 2", len(changed), 2)
}

func TestService_DiffSnapshots(t *testing.T) {
	now := time.Now()

	fromFiles := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/kept.go", now, 100).WithHash("hash1"),
		indexing.NewFileInfo("/path/to/changed.go", now, 200).WithHash("hash2"),
		indexing.NewFileInfo("/path/to/removed.go", now, 300).WithHash("hash3"),
	}

	toFiles := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/kept.go", now, 100).WithHash("hash1"),                       // unchanged
		indexing.NewFileInfo("/path/to/changed.go", now.Add(time.Hour), 250).WithHash("hash2-new"), // changed
		indexing.NewFileInfo("/path/to/added.go", now, 400).WithHash("hash4"),                      // added
	}

	store := newMockIndexStore()
	fromSnapshot := indexing.NewSnapshot("snap-from", fromFiles)
	toSnapshot := indexing.NewSnapshot("snap-to", toFiles)
	store.snapshots["snap-from"] = fromSnapshot
	store.snapshots["snap-to"] = toSnapshot

	walker := &mockFileWalker{}
	svc := indexing.NewService(walker, store, func() string { return "id" })

	diff, err := svc.DiffSnapshots(context.Background(), "snap-from", "snap-to")

	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "added count must be 1", len(diff.Added), 1)
	assert.That(t, "changed count must be 1", len(diff.Changed), 1)
	assert.That(t, "removed count must be 1", len(diff.Removed), 1)

	// Verify the correct files
	assert.That(t, "added path must match", diff.Added[0].Path, "/path/to/added.go")
	assert.That(t, "changed path must match", diff.Changed[0].Path, "/path/to/changed.go")
	assert.That(t, "removed path must match", diff.Removed[0].Path, "/path/to/removed.go")
}

func TestService_DiffSnapshots_BySize(t *testing.T) {
	now := time.Now()

	fromFiles := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file.go", now, 100), // no hash
	}

	toFiles := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file.go", now, 200), // size changed, no hash
	}

	store := newMockIndexStore()
	store.snapshots["snap-from"] = indexing.NewSnapshot("snap-from", fromFiles)
	store.snapshots["snap-to"] = indexing.NewSnapshot("snap-to", toFiles)

	walker := &mockFileWalker{}
	svc := indexing.NewService(walker, store, func() string { return "id" })

	diff, err := svc.DiffSnapshots(context.Background(), "snap-from", "snap-to")

	assert.That(t, "error must be nil", err == nil, true)
	assert.That(t, "changed count must be 1", len(diff.Changed), 1)
}

func TestService_DiffSnapshots_SnapshotNotFound(t *testing.T) {
	store := newMockIndexStore()
	walker := &mockFileWalker{}
	svc := indexing.NewService(walker, store, func() string { return "id" })

	_, err := svc.DiffSnapshots(context.Background(), "nonexistent", "also-nonexistent")

	assert.That(t, "error must not be nil", err != nil, true)
}
