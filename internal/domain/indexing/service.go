package indexing

import (
	"context"
)

// The services in a domain context are responsible for managing the workflows.
// They are responsible for creating and managing aggregates.
// They are also responsible for interacting with inbound and outbound adapters.
// They accept context.Context and forward it to the inbound and outbound adapters.
// Normally, they do not create their own contexts or call context.Cancel() directly
// unless they start background goroutines on their own.

// We use the messaging.Dispatcher to dispatch domain events.
// Only the application/infrastructure layer should use the dispatcher.
// The domain events remain explicit Go-structs will be serialized and dispatched in
// the infrastructure layer.

// IndexingService represents the service for indexing operations.
// It is responsible for managing the index repository.
// It creates aggregates.
type IndexingService struct {
	fileReader      FileReader      // inbound
	indexRepository IndexRepository // outbound
	publisher       EventPublisher  // outbound
}

// NewIndexingService creates a new IndexingService instance.
func NewIndexingService(fileReader FileReader, indexRepository IndexRepository, publisher EventPublisher) *IndexingService {
	return &IndexingService{
		fileReader:      fileReader,
		indexRepository: indexRepository,
		publisher:       publisher,
	}
}

// CreateIndex creates a new index specified by the given ID.
// We pass the context to the file reader to ensure that the operation is cancellable.
// We also pass the context to the index repository to ensure that the operation is cancellable.
func (a *IndexingService) CreateIndex(ctx context.Context, path string) error {

	files, err := a.fileReader.ReadFileInfos(ctx, path)
	if err != nil {
		return err
	}

	// Create a new index with the given ID and files.
	id := IndexID(path)
	index := NewIndex(id, files)

	// Save the index to the repository.
	if err := a.indexRepository.Create(ctx, id, index); err != nil {
		return err
	}

	// Publish the event that the index was created by using the event publisher.
	// This ensures that the event is published asynchronously and does not block the main thread.
	// We use the event publisher to decouple the indexing service from other services.
	// Only services should be allowed to publish events via the event publisher,
	// because each service handles a single use case and should be aware of if the event can be published or not.
	// If there is a previous error than now event will be published.
	evt := NewEventFileIndexCreated(id, len(files))
	if err := a.publisher.Publish(ctx, evt); err != nil {
		return err
	}

	return nil
}

// IndexFiles indexes the files specified by the given ID.
func (a *IndexingService) IndexFiles(ctx context.Context, path string) ([]string, error) {

	// Read the file infos from the file reader.
	files, err := a.fileReader.ReadFileInfos(ctx, path)
	if err != nil {
		return nil, err
	}

	// Store the file paths in a slice.
	filePaths := make([]string, len(files))
	for i, file := range files {
		filePaths[i] = file.AbsPath
	}

	return filePaths, nil
}
