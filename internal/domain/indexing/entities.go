package indexing

import (
	"time"
)

// FileInfo represents an entity that holds information about a file.
// It is used by the aggregate to store information about files in an index.
type FileInfo struct {
	ModTime time.Time
	AbsPath string
	Size    int64
}

// NewFileInfo creates a new FileInfo instance.
func NewFileInfo(absPath string, size int64, modTime time.Time) *FileInfo {
	return &FileInfo{
		AbsPath: absPath,
		Size:    size,
		ModTime: modTime,
	}
}
