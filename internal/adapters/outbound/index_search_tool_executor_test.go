package outbound_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/event"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"
)

// errIndexNotFound is a sentinel error for when an index is not found.
var errIndexNotFound = errors.New("index not found")

// mockFileReader implements the indexing.FileReader interface for testing.
type mockFileReader struct {
	err   error
	files []indexing.FileInfo
}

func (m *mockFileReader) ReadFileInfos(_ context.Context, _ string) ([]indexing.FileInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.files, nil
}

// mockEventPublisher implements the event.EventPublisher interface for testing.
type mockEventPublisher struct {
	events []event.Event
}

func (m *mockEventPublisher) Publish(_ context.Context, e event.Event) error {
	m.events = append(m.events, e)
	return nil
}

// mockIndexRepository implements the indexing.IndexRepository interface for testing.
type mockIndexRepository struct {
	indexes map[indexing.IndexID]*indexing.Index
}

func newMockIndexRepository() *mockIndexRepository {
	return &mockIndexRepository{
		indexes: make(map[indexing.IndexID]*indexing.Index),
	}
}

func (m *mockIndexRepository) Create(_ context.Context, id indexing.IndexID, index indexing.Index) error {
	m.indexes[id] = &index
	return nil
}

func (m *mockIndexRepository) Read(_ context.Context, id indexing.IndexID) (*indexing.Index, error) {
	if idx, ok := m.indexes[id]; ok {
		return idx, nil
	}
	return nil, errIndexNotFound
}

func (m *mockIndexRepository) ReadAll(_ context.Context) ([]indexing.Index, error) {
	indexes := make([]indexing.Index, 0, len(m.indexes))
	for _, idx := range m.indexes {
		indexes = append(indexes, *idx)
	}
	return indexes, nil
}

func (m *mockIndexRepository) Update(_ context.Context, id indexing.IndexID, index indexing.Index) error {
	m.indexes[id] = &index
	return nil
}

func (m *mockIndexRepository) Delete(_ context.Context, id indexing.IndexID) error {
	delete(m.indexes, id)
	return nil
}

func Test_IndexSearchToolExecutor_Execute_With_ValidSearchQuery_Should_ReturnResults(t *testing.T) {
	// Arrange
	fileReader := &mockFileReader{
		files: []indexing.FileInfo{
			{AbsPath: "/path/to/main.go", Size: 1024, ModTime: time.Now()},
			{AbsPath: "/path/to/service.go", Size: 2048, ModTime: time.Now()},
			{AbsPath: "/path/to/readme.md", Size: 512, ModTime: time.Now()},
		},
	}
	indexRepo := newMockIndexRepository()
	publisher := &mockEventPublisher{}
	indexingService := indexing.NewIndexingService(fileReader, indexRepo, publisher)

	indexID := indexing.IndexID("/path/to")
	ctx := context.Background()

	// Create the index first
	err := indexingService.CreateIndex(ctx, string(indexID))
	assert.That(t, "create index must succeed", err == nil, true)

	executor := outbound.NewIndexSearchToolExecutor(indexingService, indexID)

	// Act
	result, err := executor.Execute(ctx, "search_index", `{"query": ".go", "limit": 10}`)

	// Assert
	assert.That(t, "execute must succeed", err == nil, true)
	assert.That(t, "result must contain search results", len(result) > 0, true)
}

func Test_IndexSearchToolExecutor_Execute_With_EmptyQuery_Should_ReturnError(t *testing.T) {
	// Arrange
	fileReader := &mockFileReader{files: []indexing.FileInfo{}}
	indexRepo := newMockIndexRepository()
	publisher := &mockEventPublisher{}
	indexingService := indexing.NewIndexingService(fileReader, indexRepo, publisher)

	indexID := indexing.IndexID("/path/to")
	executor := outbound.NewIndexSearchToolExecutor(indexingService, indexID)
	ctx := context.Background()

	// Act
	_, err := executor.Execute(ctx, "search_index", `{"query": ""}`)

	// Assert
	assert.That(t, "execute must return error for empty query", err != nil, true)
}

func Test_IndexSearchToolExecutor_Execute_With_UnknownTool_Should_ReturnError(t *testing.T) {
	// Arrange
	fileReader := &mockFileReader{files: []indexing.FileInfo{}}
	indexRepo := newMockIndexRepository()
	publisher := &mockEventPublisher{}
	indexingService := indexing.NewIndexingService(fileReader, indexRepo, publisher)

	indexID := indexing.IndexID("/path/to")
	executor := outbound.NewIndexSearchToolExecutor(indexingService, indexID)
	ctx := context.Background()

	// Act
	_, err := executor.Execute(ctx, "unknown_tool", `{}`)

	// Assert
	assert.That(t, "execute must return error for unknown tool", err != nil, true)
}

func Test_IndexSearchToolExecutor_GetAvailableTools_Should_ReturnSearchIndex(t *testing.T) {
	// Arrange
	fileReader := &mockFileReader{files: []indexing.FileInfo{}}
	indexRepo := newMockIndexRepository()
	publisher := &mockEventPublisher{}
	indexingService := indexing.NewIndexingService(fileReader, indexRepo, publisher)

	indexID := indexing.IndexID("/path/to")
	executor := outbound.NewIndexSearchToolExecutor(indexingService, indexID)

	// Act
	tools := executor.GetAvailableTools()

	// Assert
	assert.That(t, "must have one tool", len(tools), 1)
	assert.That(t, "must contain search_index", tools[0], "search_index")
}

func Test_IndexSearchToolExecutor_GetToolDefinitions_Should_ReturnDefinitions(t *testing.T) {
	// Arrange
	fileReader := &mockFileReader{files: []indexing.FileInfo{}}
	indexRepo := newMockIndexRepository()
	publisher := &mockEventPublisher{}
	indexingService := indexing.NewIndexingService(fileReader, indexRepo, publisher)

	indexID := indexing.IndexID("/path/to")
	executor := outbound.NewIndexSearchToolExecutor(indexingService, indexID)

	// Act
	definitions := executor.GetToolDefinitions()

	// Assert
	assert.That(t, "must have one definition", len(definitions), 1)
	assert.That(t, "definition name must be search_index", definitions[0].Name, "search_index")
	assert.That(t, "definition must have query parameter", definitions[0].Parameters["query"] != "", true)
}

func Test_IndexSearchToolExecutor_HasTool_With_SearchIndex_Should_ReturnTrue(t *testing.T) {
	// Arrange
	fileReader := &mockFileReader{files: []indexing.FileInfo{}}
	indexRepo := newMockIndexRepository()
	publisher := &mockEventPublisher{}
	indexingService := indexing.NewIndexingService(fileReader, indexRepo, publisher)

	indexID := indexing.IndexID("/path/to")
	executor := outbound.NewIndexSearchToolExecutor(indexingService, indexID)

	// Act
	hasSearchIndex := executor.HasTool("search_index")
	hasUnknown := executor.HasTool("unknown")

	// Assert
	assert.That(t, "must have search_index tool", hasSearchIndex, true)
	assert.That(t, "must not have unknown tool", hasUnknown, false)
}
