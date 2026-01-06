package indexing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
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

// Search searches the index for files matching the given query.
// It performs a case-insensitive substring match on file paths.
// Returns up to limit results, sorted by relevance (exact matches first).
func (a *Index) Search(query string, limit int) []SearchResult {
	if query == "" || limit <= 0 {
		return []SearchResult{}
	}

	results := a.findMatchingFiles(strings.ToLower(query))
	sortByScoreDescending(results)

	if len(results) > limit {
		return results[:limit]
	}
	return results
}

// findMatchingFiles finds all files matching the query and calculates their scores.
func (a *Index) findMatchingFiles(queryLower string) []SearchResult {
	var results []SearchResult
	for _, fileInfo := range a.FileInfos {
		if result, ok := a.matchFile(fileInfo, queryLower); ok {
			results = append(results, result)
		}
	}
	return results
}

// matchFile checks if a file matches the query and returns its score.
func (a *Index) matchFile(fileInfo FileInfo, queryLower string) (SearchResult, bool) {
	pathLower := strings.ToLower(fileInfo.AbsPath)
	if !strings.Contains(pathLower, queryLower) {
		return SearchResult{}, false
	}
	score := calculateScore(pathLower, queryLower)
	return NewSearchResult(fileInfo.AbsPath).WithScore(score), true
}

// calculateScore calculates the relevance score for a match.
func calculateScore(pathLower, queryLower string) float64 {
	fileNameLower := strings.ToLower(filepath.Base(pathLower))
	if fileNameLower == queryLower {
		return 2.0 // Exact filename match
	}
	if strings.Contains(fileNameLower, queryLower) {
		return 1.0 // Filename contains query
	}
	return 0.5 // Path-only match
}

// sortByScoreDescending sorts results by score in descending order.
func sortByScoreDescending(results []SearchResult) {
	for i := range len(results) - 1 {
		for j := range len(results) - i - 1 {
			if results[j].Score < results[j+1].Score {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}
