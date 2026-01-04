package indexing

import "context"

// FileReader represents the interface for interacting with the filesystem.
// It is responsible for reading file information from the filesystem.
// We use context.Context to provide cancellation and timeout capabilities.
type FileReader interface {
	ReadFileInfos(ctx context.Context, path string) ([]FileInfo, error)
}
