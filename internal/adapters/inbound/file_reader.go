package inbound

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"
)

// The Inbound package creates (e.g. via context.Background()) or receives the context, e.g. via http.Request.Context().
// It is responsible for interacting with the filesystem and providing file information to the application.
// It passes the content into the application layer.

// FileReader represents the interface for interacting with the filesystem.
// It is responsible for reading file information from the filesystem.
type FileReader struct{}

// NewFileReader creates a new instance of indexing.FileReader.
// This ensures that the FileReader meets the requirements of the indexing.FileReader interface.
func NewFileReader() indexing.FileReader {
	return &FileReader{}
}

// ReadFileInfos reads file information from the filesystem.
// This methods implements the infrastructure layer for reading file information.
func (a *FileReader) ReadFileInfos(ctx context.Context, root string) ([]indexing.FileInfo, error) {
	var fileInfos []indexing.FileInfo

	// Walk the directory tree and collect file information.
	err := a.walk(root, &fileInfos)
	if err != nil {
		return nil, err
	}

	return fileInfos, nil
}

// walk recursively walks the directory tree and collects file information.
func (a *FileReader) walk(dir string, out *[]indexing.FileInfo) error {
	// Get the directory entries.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Recursively walk directories.
	for _, e := range entries {
		absPath := filepath.Join(dir, e.Name())

		// Skip hidden files and directories.
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}

		// Skip directories by walking them recursively.
		if e.IsDir() {
			if err := a.walk(absPath, out); err != nil {
				return err
			}
			continue
		}

		// Get the file info.
		info, err := e.Info()
		if err != nil {
			return err
		}

		// Create a new FileInfo entity.
		fileInfo := indexing.NewFileInfo(absPath, info.Size(), info.ModTime())

		// Append the FileInfo entity to the output slice.
		*out = append(*out, *fileInfo)
	}
	return nil
}
