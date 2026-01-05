package indexing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Index represents the aggregate for indexing.
// It is responsible for consistency and integrity of the index data.
// This ensures that the Index is a valid and consistent representation of the indexed files.
type Index struct {
	ID        IndexID
	CreatedAt time.Time
	UpdatedAt time.Time
	FileInfos []FileInfo
}

// NewIndex creates a new Index instance with the given ID and fileInfos.
func NewIndex(id IndexID, fileInfos []FileInfo) Index {
	return Index{
		ID:        id,
		FileInfos: fileInfos,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Hash returns a hash of the fileInfos.
// It is used to detect changes like file additions or deletions.
// It can also be used to verify the integrity of the index data.
func (a *Index) Hash() string {
	hasher := sha256.New()
	// The hash is calculated by concatenating the absolute path and size of each file info.
	// This ensures that the hash changes when the file info changes.
	// The hash does not include the IndexID because it is not part of the file info.
	// Thus even if the IndexID changes, the hash will remain the same.
	for _, fileInfo := range a.FileInfos {
		_, _ = fmt.Fprintf(hasher, "%s-%d|", fileInfo.AbsPath, fileInfo.Size)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}
