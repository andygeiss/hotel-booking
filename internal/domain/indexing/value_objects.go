package indexing

// IndexID represents a value object and unique identifier for an index.
// It is used by the aggregate to identify the index.
type IndexID string

// SearchResult represents a single result from searching the index.
// It contains the file path and optional metadata about the match.
type SearchResult struct {
	FilePath string  `json:"file_path"`
	Snippet  string  `json:"snippet,omitempty"`
	Score    float64 `json:"score,omitempty"`
}

// NewSearchResult creates a new search result with the given file path.
func NewSearchResult(filePath string) SearchResult {
	return SearchResult{
		FilePath: filePath,
	}
}

// WithSnippet sets the matching snippet for the result.
func (r SearchResult) WithSnippet(snippet string) SearchResult {
	r.Snippet = snippet
	return r
}

// WithScore sets the relevance score for the result.
func (r SearchResult) WithScore(score float64) SearchResult {
	r.Score = score
	return r
}
