package indexing_test

import (
	"context"
	"testing"

	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/event"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/cloud-native-utils/resource"
)

// Every test should follow the Test_<struct>_<method>_With_<condition>_Should_<result> pattern.
// This is important because we want the tests to be easy to read and understand.
// We use the Arrange-Act-Assert pattern for better readability.
// We use the assert package from the cloud-native-utils library for better readability.

// In order to test the IndexingService, we need to mock the dependencies.
// We need to create a simple mock implementation for the FileReader interface,
// but it has only one method, ReadFile, which returns a string and an error.
// We can use the resource.MockAccess implementation to mock the access to the resource,
// because the IndexRepository is of type resource.Repository[indexing.IndexID, indexing.Index].

func Test_IndexingService_CreateIndex_With_Mockup_Should_Return_Two_Entries(t *testing.T) {
	// Arrange
	sut, _ := setupIndexingService()
	path := "testdata/index.json"
	ctx := context.Background()

	// Act
	err := sut.CreateIndex(ctx, path)
	files, err2 := sut.IndexFiles(ctx, path)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "err2 must be nil", err2 == nil, true)
	assert.That(t, "index must have two entries", len(files) == 2, true)
}

func Test_IndexingService_CreateIndex_With_Mockup_Should_Be_Called(t *testing.T) {
	// Arrange
	sut, publisher := setupIndexingService()
	publisher = publisher.(*mockEventPublisher)
	path := "testdata/index.json"
	ctx := context.Background()

	// Act
	err := sut.CreateIndex(ctx, path)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "publisher must be called", publisher.(*mockEventPublisher).Published, true)
}

// mockFileReader is a simple mock implementation for the FileReader interface.
type mockFileReader struct {
	fileInfos []indexing.FileInfo
}

// ReadFileInfos returns the slice of file infos.
func (a *mockFileReader) ReadFileInfos(ctx context.Context, path string) ([]indexing.FileInfo, error) {
	return a.fileInfos, nil
}

// mockEventPublisher is a simple mock implementation for the EventPublisher interface.
type mockEventPublisher struct {
	Published bool
}

// Publish publishes the message.
func (a *mockEventPublisher) Publish(ctx context.Context, e event.Event) error {
	a.Published = true
	return nil
}

// setupIndexingService creates a new IndexingService with mocked dependencies.
func setupIndexingService() (*indexing.IndexingService, event.EventPublisher) {
	mockFileReader := &mockFileReader{
		fileInfos: []indexing.FileInfo{
			{AbsPath: "test/path/file1.txt", Size: 100},
			{AbsPath: "test/path/file2.txt", Size: 200},
		},
	}
	mockIndexRepository := resource.NewMockAccess[indexing.IndexID, indexing.Index]()
	mockIndexRepository.WithCreateFn(
		func(ctx context.Context, id indexing.IndexID, index indexing.Index) error {
			return nil
		})
	mockEventPublisher := &mockEventPublisher{}
	service := indexing.NewIndexingService(mockFileReader, mockIndexRepository, mockEventPublisher)
	return service, mockEventPublisher
}
