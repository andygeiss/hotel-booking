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

func Test_IndexID_With_String_Value_Should_Be_Assignable(t *testing.T) {
	// Arrange
	expected := "test-index-id"

	// Act
	id := indexing.IndexID(expected)

	// Assert
	assert.That(t, "IndexID must match string value", string(id), expected)
}

func Test_IndexID_With_Empty_String_Should_Be_Valid(t *testing.T) {
	// Arrange
	expected := ""

	// Act
	id := indexing.IndexID(expected)

	// Assert
	assert.That(t, "IndexID must be empty", string(id), expected)
}

func Test_IndexID_With_UUID_Format_Should_Be_Valid(t *testing.T) {
	// Arrange
	expected := "550e8400-e29b-41d4-a716-446655440000"

	// Act
	id := indexing.IndexID(expected)

	// Assert
	assert.That(t, "IndexID must handle UUID format", string(id), expected)
}

func Test_SearchResult_NewSearchResult_Should_SetFilePath(t *testing.T) {
	// Arrange & Act
	result := indexing.NewSearchResult("/path/to/file.go")

	// Assert
	assert.That(t, "file path must match", result.FilePath, "/path/to/file.go")
}

func Test_SearchResult_WithSnippet_Should_SetSnippet(t *testing.T) {
	// Arrange
	result := indexing.NewSearchResult("/path/to/file.go")

	// Act
	result = result.WithSnippet("matching content")

	// Assert
	assert.That(t, "snippet must match", result.Snippet, "matching content")
}

func Test_SearchResult_WithScore_Should_SetScore(t *testing.T) {
	// Arrange
	result := indexing.NewSearchResult("/path/to/file.go")

	// Act
	result = result.WithScore(0.95)

	// Assert
	assert.That(t, "score must match", result.Score, 0.95)
}
