package indexing_test

import (
	"testing"

	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"

	"github.com/andygeiss/cloud-native-utils/assert"
)

// Every test should follow the Test_<struct>_<method>_With_<condition>_Should_<result> pattern.
// This is important because we want the tests to be easy to read and understand.
// We use the Arrange-Act-Assert pattern for better readability.
// We use the assert package from the cloud-native-utils library for better readability.

func Test_Index_Hash_With_No_FileInfos_Should_Return_Valid_Hash(t *testing.T) {
	// Arrange
	index := indexing.Index{
		ID:        "empty-index",
		FileInfos: []indexing.FileInfo{},
	}

	// Act
	hash := index.Hash()

	// Assert
	assert.That(t, "empty index must have a valid hash (size of 64 bytes)", len(hash), 64)
}

func Test_Index_Hash_With_One_FileInfo_Should_Return_Valid_Hash(t *testing.T) {
	// Arrange
	index := indexing.Index{
		ID: "single-file-index",
		FileInfos: []indexing.FileInfo{
			{AbsPath: "file.txt", Size: 1024},
		},
	}

	// Act
	hash := index.Hash()

	// Assert
	assert.That(t, "single file index must have a valid hash (size of 64 bytes)", len(hash), 64)
}

func Test_Index_Hash_With_Multiple_FileInfos_Should_Return_Valid_Hash(t *testing.T) {
	// Arrange
	index := indexing.Index{
		ID: "multiple-files-index",
		FileInfos: []indexing.FileInfo{
			{AbsPath: "file1.txt", Size: 1024},
			{AbsPath: "file2.txt", Size: 2048},
		},
	}

	// Act
	hash := index.Hash()

	// Assert
	assert.That(t, "multiple files index must have a valid hash (size of 64 bytes)", len(hash), 64)
}

func Test_Index_Hash_With_Same_FileInfos_Should_Return_Same_Hash(t *testing.T) {
	// Arrange
	index1 := indexing.Index{
		ID: "same-files-index-1",
		FileInfos: []indexing.FileInfo{
			{AbsPath: "file1.txt", Size: 1024},
			{AbsPath: "file2.txt", Size: 2048},
		},
	}

	index2 := indexing.Index{
		ID: "same-files-index-2",
		FileInfos: []indexing.FileInfo{
			{AbsPath: "file1.txt", Size: 1024},
			{AbsPath: "file2.txt", Size: 2048},
		},
	}

	// Act
	hash1 := index1.Hash()
	hash2 := index2.Hash()

	// Assert
	assert.That(t, "same file infos must produce the same hash", hash1 == hash2, true)
}

func Test_Index_Hash_With_Different_FileInfos_Should_Return_Different_Hash(t *testing.T) {
	// Arrange
	index1 := indexing.Index{
		ID: "different-files-index-1",
		FileInfos: []indexing.FileInfo{
			{AbsPath: "file1.txt", Size: 1024},
			{AbsPath: "file2.txt", Size: 2048},
		},
	}

	index2 := indexing.Index{
		ID: "different-files-index-2",
		FileInfos: []indexing.FileInfo{
			{AbsPath: "file1.txt", Size: 1024},
			{AbsPath: "file3.txt", Size: 3072},
		},
	}

	// Act
	hash1 := index1.Hash()
	hash2 := index2.Hash()

	// Assert
	assert.That(t, "different file infos must produce different hashes", hash1 != hash2, true)
}

func Test_Index_Search_With_EmptyQuery_Should_ReturnEmptyResults(t *testing.T) {
	// Arrange
	index := indexing.Index{
		ID: "search-test-index",
		FileInfos: []indexing.FileInfo{
			{AbsPath: "/path/to/file.go", Size: 1024},
			{AbsPath: "/path/to/main.go", Size: 2048},
		},
	}

	// Act
	results := index.Search("", 10)

	// Assert
	assert.That(t, "empty query must return empty results", len(results), 0)
}

func Test_Index_Search_With_MatchingQuery_Should_ReturnMatchingFiles(t *testing.T) {
	// Arrange
	index := indexing.Index{
		ID: "search-test-index",
		FileInfos: []indexing.FileInfo{
			{AbsPath: "/path/to/file.go", Size: 1024},
			{AbsPath: "/path/to/main.go", Size: 2048},
			{AbsPath: "/path/to/readme.md", Size: 512},
		},
	}

	// Act
	results := index.Search(".go", 10)

	// Assert
	assert.That(t, "search must return matching files", len(results), 2)
}

func Test_Index_Search_With_Limit_Should_RespectLimit(t *testing.T) {
	// Arrange
	index := indexing.Index{
		ID: "search-test-index",
		FileInfos: []indexing.FileInfo{
			{AbsPath: "/path/to/file1.go", Size: 1024},
			{AbsPath: "/path/to/file2.go", Size: 2048},
			{AbsPath: "/path/to/file3.go", Size: 512},
		},
	}

	// Act
	results := index.Search(".go", 2)

	// Assert
	assert.That(t, "search must respect limit", len(results), 2)
}

func Test_Index_Search_With_CaseInsensitiveQuery_Should_ReturnMatches(t *testing.T) {
	// Arrange
	index := indexing.Index{
		ID: "search-test-index",
		FileInfos: []indexing.FileInfo{
			{AbsPath: "/path/to/README.md", Size: 1024},
			{AbsPath: "/path/to/readme.txt", Size: 2048},
		},
	}

	// Act
	results := index.Search("readme", 10)

	// Assert
	assert.That(t, "case-insensitive search must find both files", len(results), 2)
}

func Test_Index_Search_With_ExactFilename_Should_HaveHigherScore(t *testing.T) {
	// Arrange
	index := indexing.Index{
		ID: "search-test-index",
		FileInfos: []indexing.FileInfo{
			{AbsPath: "/path/to/utils/main.go.bak", Size: 1024}, // contains "main.go" but not exact
			{AbsPath: "/path/to/main.go", Size: 2048},           // exact filename match
		},
	}

	// Act
	results := index.Search("main.go", 10)

	// Assert
	assert.That(t, "search must return results", len(results), 2)
	assert.That(t, "exact filename must have higher score", results[0].FilePath, "/path/to/main.go")
}
