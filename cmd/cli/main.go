package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/event"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"
)

func main() {
	ctx := context.Background()

	// Setup the messaging dispatcher (external Kafka pub/sub).
	// Requires KAFKA_BROKERS environment variable to be set.
	// Example: KAFKA_BROKERS=localhost:9092 for local, kafka:9092 for Docker Compose.
	dispatcher := messaging.NewExternalDispatcher()

	// Setup the inbound adapters.
	fileReader := inbound.NewFileReader()
	eventSubscriber := inbound.NewEventSubscriber(dispatcher)

	// Setup the outbound adapters.
	indexPath := "./index.json"
	defer func() { _ = os.Remove(indexPath) }()
	indexRepository := outbound.NewFileIndexRepository(indexPath)
	eventPublisher := outbound.NewEventPublisher(dispatcher)

	// Subscribe to the EventFileIndexCreated event before creating the index.
	// This demonstrates the event-driven architecture where subscribers react to domain events.
	err := eventSubscriber.Subscribe(
		ctx,
		indexing.EventTopicFileIndexCreated,
		func() event.Event { return indexing.NewEventFileIndexCreated() },
		func(e event.Event) error {
			evt := e.(*indexing.EventFileIndexCreated)
			fmt.Printf("❯ event: received EventFileIndexCreated - IndexID: %s, FileCount: %d\n", evt.IndexID, evt.FileCount)
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	// Create the indexing service with all dependencies injected.
	indexingService := indexing.NewIndexingService(fileReader, indexRepository, eventPublisher)

	// Use the service to create an index (this will also publish the event).
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fmt.Printf("❯ main: creating index for path: %s\n", wd)

	if err := indexingService.CreateIndex(ctx, wd); err != nil {
		panic(err)
	}

	// Give a moment for async event processing.
	time.Sleep(100 * time.Millisecond)

	// Read the index back from the repository to demonstrate the full cycle.
	id := indexing.IndexID(wd)
	index, err := indexRepository.Read(ctx, id)
	if err != nil {
		panic(err)
	}

	fmt.Printf("❯ main: index created at %s with %d files\n", index.CreatedAt.Format(time.RFC3339), len(index.FileInfos))
	fmt.Printf("❯ main: index hash: %s\n", index.Hash())

	// Demonstrate listing some file infos from the index.
	fmt.Printf("❯ main: first 5 files in index:\n")
	for i, fi := range index.FileInfos {
		if i >= 5 {
			fmt.Printf("  ... and %d more files\n", len(index.FileInfos)-5)
			break
		}
		fmt.Printf("  - %s (%d bytes)\n", fi.AbsPath, fi.Size)
	}
}
