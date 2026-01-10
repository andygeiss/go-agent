package tooling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/andygeiss/go-agent/internal/domain/agent"
	"github.com/andygeiss/go-agent/internal/domain/indexing"
)

// ErrPathsRequired indicates that paths argument is required.
var ErrPathsRequired = errors.New("paths is required")

// indexScanArgs represents the arguments for the index.scan tool.
type indexScanArgs struct {
	Ignore []string `json:"ignore,omitempty"`
	Paths  []string `json:"paths"`
}

// indexChangedSinceArgs represents the arguments for the index.changed_since tool.
type indexChangedSinceArgs struct {
	Since string `json:"since"`
}

// indexDiffSnapshotArgs represents the arguments for the index.diff_snapshot tool.
type indexDiffSnapshotArgs struct {
	FromID string `json:"from_id"`
	ToID   string `json:"to_id"`
}

// indexScanResult represents the result of the index.scan tool.
type indexScanResult struct {
	IndexedAt    string `json:"indexed_at"`
	SnapshotID   string `json:"snapshot_id"`
	Status       string `json:"status"`
	FilesIndexed int    `json:"files_indexed"`
	FilesTotal   int    `json:"files_total"`
}

// indexChangedSinceResult represents the result of the index.changed_since tool.
type indexChangedSinceResult struct {
	Status string            `json:"status"`
	Files  []indexFileResult `json:"files"`
	Count  int               `json:"count"`
}

// indexFileResult represents a single file in the result.
type indexFileResult struct {
	ModTime string `json:"mod_time"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
}

// indexDiffSnapshotResult represents the result of the index.diff_snapshot tool.
type indexDiffSnapshotResult struct {
	Status  string            `json:"status"`
	Added   []indexFileResult `json:"added"`
	Changed []indexFileResult `json:"changed"`
	Removed []indexFileResult `json:"removed"`
}

// IndexToolService provides indexing tool implementations.
type IndexToolService struct {
	svc *indexing.Service
}

// NewIndexToolService creates a new index tool service.
func NewIndexToolService(svc *indexing.Service) *IndexToolService {
	return &IndexToolService{svc: svc}
}

// IndexScan scans the given paths and creates a new snapshot.
func (s *IndexToolService) IndexScan(ctx context.Context, arguments string) (string, error) {
	var args indexScanArgs
	if err := agent.DecodeArgs(arguments, &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	if len(args.Paths) == 0 {
		return "", ErrPathsRequired
	}

	snapshot, err := s.svc.Scan(ctx, args.Paths, args.Ignore)
	if err != nil {
		return "", fmt.Errorf("failed to scan: %w", err)
	}

	result := indexScanResult{
		FilesIndexed: snapshot.FileCount(),
		FilesTotal:   snapshot.FileCount(),
		IndexedAt:    snapshot.CreatedAt.Format(time.RFC3339),
		SnapshotID:   string(snapshot.ID),
		Status:       "success",
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// IndexChangedSince returns files changed since the given timestamp.
func (s *IndexToolService) IndexChangedSince(ctx context.Context, arguments string) (string, error) {
	var args indexChangedSinceArgs
	if err := agent.DecodeArgs(arguments, &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	since, err := time.Parse(time.RFC3339, args.Since)
	if err != nil {
		return "", fmt.Errorf("failed to parse since timestamp: %w", err)
	}

	files, err := s.svc.ChangedSince(ctx, since)
	if err != nil {
		return "", fmt.Errorf("failed to get changed files: %w", err)
	}

	result := indexChangedSinceResult{
		Count:  len(files),
		Files:  convertFileInfosToResults(files),
		Status: "success",
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// IndexDiffSnapshot compares two snapshots and returns the differences.
func (s *IndexToolService) IndexDiffSnapshot(ctx context.Context, arguments string) (string, error) {
	var args indexDiffSnapshotArgs
	if err := agent.DecodeArgs(arguments, &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	diff, err := s.svc.DiffSnapshots(ctx, indexing.SnapshotID(args.FromID), indexing.SnapshotID(args.ToID))
	if err != nil {
		return "", fmt.Errorf("failed to diff snapshots: %w", err)
	}

	result := indexDiffSnapshotResult{
		Added:   convertFileInfosToResults(diff.Added),
		Changed: convertFileInfosToResults(diff.Changed),
		Removed: convertFileInfosToResults(diff.Removed),
		Status:  "success",
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// convertFileInfosToResults converts FileInfo slices to result format.
func convertFileInfosToResults(files []indexing.FileInfo) []indexFileResult {
	results := make([]indexFileResult, len(files))
	for i, f := range files {
		results[i] = indexFileResult{
			ModTime: f.ModTime.Format(time.RFC3339),
			Path:    f.Path,
			Size:    f.Size,
		}
	}
	return results
}

// NewIndexScanTool creates the index.scan tool definition.
func NewIndexScanTool(svc *IndexToolService) agent.Tool {
	return agent.Tool{
		ID: "index.scan",
		Definition: agent.NewToolDefinition("index.scan", "Scan directories and create a snapshot of all files. Use this to index a codebase before analyzing changes.").
			WithParameterDef(agent.NewParameterDefinition("paths", agent.ParamTypeArray).
				WithDescription("List of directory paths to scan (absolute or relative)").
				WithRequired()).
			WithParameterDef(agent.NewParameterDefinition("ignore", agent.ParamTypeArray).
				WithDescription("Patterns to ignore (e.g., node_modules, .git, *.log)")),
		Func: svc.IndexScan,
	}
}

// NewIndexChangedSinceTool creates the index.changed_since tool definition.
func NewIndexChangedSinceTool(svc *IndexToolService) agent.Tool {
	return agent.Tool{
		ID: "index.changed_since",
		Definition: agent.NewToolDefinition("index.changed_since", "Get files changed since a given timestamp from the latest snapshot.").
			WithParameterDef(agent.NewParameterDefinition("since", agent.ParamTypeString).
				WithDescription("RFC3339 timestamp (e.g., 2024-01-15T10:00:00Z)").
				WithRequired()),
		Func: svc.IndexChangedSince,
	}
}

// NewIndexDiffSnapshotTool creates the index.diff_snapshot tool definition.
func NewIndexDiffSnapshotTool(svc *IndexToolService) agent.Tool {
	return agent.Tool{
		ID: "index.diff_snapshot",
		Definition: agent.NewToolDefinition("index.diff_snapshot", "Compare two snapshots and return added, removed, and changed files.").
			WithParameterDef(agent.NewParameterDefinition("from_id", agent.ParamTypeString).
				WithDescription("ID of the older snapshot").
				WithRequired()).
			WithParameterDef(agent.NewParameterDefinition("to_id", agent.ParamTypeString).
				WithDescription("ID of the newer snapshot").
				WithRequired()),
		Func: svc.IndexDiffSnapshot,
	}
}
