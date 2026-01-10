package tooling_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/indexing"
	"github.com/andygeiss/go-agent/internal/domain/tooling"
)

// mockIndexFileWalker is a test double for FileWalker.
type mockIndexFileWalker struct {
	err   error
	files []indexing.FileInfo
}

func (m *mockIndexFileWalker) Walk(_ context.Context, _ []string, _ []string) ([]indexing.FileInfo, error) {
	return m.files, m.err
}

// mockIndexingStore is a test double for IndexStore.
type mockIndexingStore struct {
	err       error
	snapshots map[indexing.SnapshotID]indexing.Snapshot
	latest    indexing.Snapshot
}

func newMockIndexingStore() *mockIndexingStore {
	return &mockIndexingStore{
		snapshots: make(map[indexing.SnapshotID]indexing.Snapshot),
	}
}

func (m *mockIndexingStore) SaveSnapshot(_ context.Context, snapshot indexing.Snapshot) error {
	if m.err != nil {
		return m.err
	}
	m.snapshots[snapshot.ID] = snapshot
	m.latest = snapshot
	return nil
}

func (m *mockIndexingStore) GetLatestSnapshot(_ context.Context) (indexing.Snapshot, error) {
	if m.err != nil {
		return indexing.Snapshot{}, m.err
	}
	return m.latest, nil
}

func (m *mockIndexingStore) GetSnapshot(_ context.Context, id indexing.SnapshotID) (indexing.Snapshot, error) {
	if m.err != nil {
		return indexing.Snapshot{}, m.err
	}
	if snapshot, ok := m.snapshots[id]; ok {
		return snapshot, nil
	}
	return indexing.Snapshot{}, indexing.ErrSnapshotNotFound
}

// indexScanResult matches the response structure from IndexScan.
type indexScanResult struct {
	SnapshotID   string `json:"snapshot_id"`
	Status       string `json:"status"`
	FilesIndexed int    `json:"files_indexed"`
}

// indexChangedSinceResult matches the response structure from IndexChangedSince.
type indexChangedSinceResult struct {
	Status       string   `json:"status"`
	ChangedFiles []string `json:"changed_files"`
	Count        int      `json:"count"`
}

// indexDiffSnapshotResult matches the response structure from IndexDiffSnapshot.
type indexDiffSnapshotResult struct {
	Status  string   `json:"status"`
	Added   []string `json:"added"`
	Changed []string `json:"changed"`
	Removed []string `json:"removed"`
}

func Test_IndexToolService_IndexScan_Should_ReturnSnapshotInfo(t *testing.T) {
	// Arrange
	now := time.Now()
	files := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/file1.go", now, 100),
		indexing.NewFileInfo("/path/to/file2.go", now, 200),
	}

	walker := &mockIndexFileWalker{files: files}
	store := newMockIndexingStore()
	idGen := func() string { return "snap-test" }
	svc := indexing.NewService(walker, store, idGen)
	toolSvc := tooling.NewIndexToolService(svc)

	args := `{"paths": ["/path/to/project"]}`

	// Act
	result, err := toolSvc.IndexScan(context.Background(), args)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)

	var response indexScanResult
	_ = json.Unmarshal([]byte(result), &response)

	assert.That(t, "status must be success", response.Status, "success")
	assert.That(t, "snapshot ID must match", response.SnapshotID, "snap-test")
	assert.That(t, "files indexed must be 2", response.FilesIndexed, 2)
}

func Test_IndexToolService_IndexScan_With_EmptyPaths_Should_ReturnError(t *testing.T) {
	// Arrange
	walker := &mockIndexFileWalker{}
	store := newMockIndexingStore()
	svc := indexing.NewService(walker, store, func() string { return "id" })
	toolSvc := tooling.NewIndexToolService(svc)

	args := `{"paths": []}`

	// Act
	_, err := toolSvc.IndexScan(context.Background(), args)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_IndexToolService_IndexChangedSince_Should_ReturnChangedFiles(t *testing.T) {
	// Arrange
	baseTime := time.Now().Add(-2 * time.Hour)
	files := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/old.go", baseTime.Add(-1*time.Hour), 100),
		indexing.NewFileInfo("/path/to/new.go", baseTime.Add(30*time.Minute), 200),
	}

	walker := &mockIndexFileWalker{}
	store := newMockIndexingStore()
	store.latest = indexing.NewSnapshot("snap-1", files)

	svc := indexing.NewService(walker, store, func() string { return "id" })
	toolSvc := tooling.NewIndexToolService(svc)

	args := `{"since": "` + baseTime.Format(time.RFC3339) + `"}`

	// Act
	result, err := toolSvc.IndexChangedSince(context.Background(), args)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)

	var response indexChangedSinceResult
	_ = json.Unmarshal([]byte(result), &response)

	assert.That(t, "status must be success", response.Status, "success")
	assert.That(t, "count must be 1", response.Count, 1)
}

func Test_IndexToolService_IndexChangedSince_With_InvalidTimestamp_Should_ReturnError(t *testing.T) {
	// Arrange
	walker := &mockIndexFileWalker{}
	store := newMockIndexingStore()
	svc := indexing.NewService(walker, store, func() string { return "id" })
	toolSvc := tooling.NewIndexToolService(svc)

	args := `{"since": "invalid-timestamp"}`

	// Act
	_, err := toolSvc.IndexChangedSince(context.Background(), args)

	// Assert
	assert.That(t, "error must not be nil", err != nil, true)
}

func Test_IndexToolService_IndexDiffSnapshot_Should_ReturnDiff(t *testing.T) {
	// Arrange
	now := time.Now()

	fromFiles := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/kept.go", now, 100).WithHash("hash1"),
		indexing.NewFileInfo("/path/to/removed.go", now, 200).WithHash("hash2"),
	}

	toFiles := []indexing.FileInfo{
		indexing.NewFileInfo("/path/to/kept.go", now, 100).WithHash("hash1"),
		indexing.NewFileInfo("/path/to/added.go", now, 300).WithHash("hash3"),
	}

	walker := &mockIndexFileWalker{}
	store := newMockIndexingStore()
	store.snapshots["snap-from"] = indexing.NewSnapshot("snap-from", fromFiles)
	store.snapshots["snap-to"] = indexing.NewSnapshot("snap-to", toFiles)

	svc := indexing.NewService(walker, store, func() string { return "id" })
	toolSvc := tooling.NewIndexToolService(svc)

	args := `{"from_id": "snap-from", "to_id": "snap-to"}`

	// Act
	result, err := toolSvc.IndexDiffSnapshot(context.Background(), args)

	// Assert
	assert.That(t, "error must be nil", err == nil, true)

	var response indexDiffSnapshotResult
	_ = json.Unmarshal([]byte(result), &response)

	assert.That(t, "status must be success", response.Status, "success")
	assert.That(t, "added count must be 1", len(response.Added), 1)
	assert.That(t, "removed count must be 1", len(response.Removed), 1)
}

func Test_NewIndexScanTool_Should_ReturnValidTool(t *testing.T) {
	// Arrange
	walker := &mockIndexFileWalker{}
	store := newMockIndexingStore()
	svc := indexing.NewService(walker, store, func() string { return "id" })
	toolSvc := tooling.NewIndexToolService(svc)

	// Act
	tool := tooling.NewIndexScanTool(toolSvc)

	// Assert
	assert.That(t, "tool ID must match", tool.ID, agent.ToolID("index.scan"))
	assert.That(t, "definition name must match", tool.Definition.Name, "index.scan")
	assert.That(t, "func must not be nil", tool.Func != nil, true)
}

func Test_NewIndexChangedSinceTool_Should_ReturnValidTool(t *testing.T) {
	// Arrange
	walker := &mockIndexFileWalker{}
	store := newMockIndexingStore()
	svc := indexing.NewService(walker, store, func() string { return "id" })
	toolSvc := tooling.NewIndexToolService(svc)

	// Act
	tool := tooling.NewIndexChangedSinceTool(toolSvc)

	// Assert
	assert.That(t, "tool ID must match", tool.ID, agent.ToolID("index.changed_since"))
	assert.That(t, "definition name must match", tool.Definition.Name, "index.changed_since")
	assert.That(t, "func must not be nil", tool.Func != nil, true)
}

func Test_NewIndexDiffSnapshotTool_Should_ReturnValidTool(t *testing.T) {
	// Arrange
	walker := &mockIndexFileWalker{}
	store := newMockIndexingStore()
	svc := indexing.NewService(walker, store, func() string { return "id" })
	toolSvc := tooling.NewIndexToolService(svc)

	// Act
	tool := tooling.NewIndexDiffSnapshotTool(toolSvc)

	// Assert
	assert.That(t, "tool ID must match", tool.ID, agent.ToolID("index.diff_snapshot"))
	assert.That(t, "definition name must match", tool.Definition.Name, "index.diff_snapshot")
	assert.That(t, "func must not be nil", tool.Func != nil, true)
}
