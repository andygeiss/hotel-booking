package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/event"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"
)

func main() {
	// Create a context with timeout for the indexing task
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup the messaging dispatcher (external Kafka pub/sub).
	dispatcher := messaging.NewExternalDispatcher()

	// Setup the inbound adapters.
	fileReader := inbound.NewFileReader()
	eventSubscriber := inbound.NewEventSubscriber(dispatcher)

	// Setup the outbound adapters.
	indexPath := "./index.json"
	defer func() { _ = os.Remove(indexPath) }()
	indexRepository := outbound.NewFileIndexRepository(indexPath)
	eventPublisher := outbound.NewEventPublisher(dispatcher)

	// Create the indexing service with all dependencies injected.
	indexingService := indexing.NewIndexingService(fileReader, indexRepository, eventPublisher)

	// Use the service to create an index (this will also publish the event).
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Create the index ID from the working directory
	indexID := indexing.IndexID(wd)

	// WaitGroup and Once to coordinate event processing and prevent duplicate handling
	var wg sync.WaitGroup
	var once sync.Once
	wg.Add(1)

	// Subscribe to the EventFileIndexCreated event.
	err = eventSubscriber.Subscribe(
		ctx,
		indexing.EventTopicFileIndexCreated,
		func() event.Event { return indexing.NewEventFileIndexCreated() },
		createIndexEventHandler(&wg, &once),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("❯ main: creating index for path: %s\n", wd)

	createErr := indexingService.CreateIndex(ctx, wd)
	if createErr != nil {
		panic(createErr)
	}

	// Wait for event processing to complete
	waitForCompletion(ctx, &wg)

	// Read the index back from the repository to demonstrate the full cycle.
	index, err := indexRepository.Read(context.Background(), indexID)
	if err != nil {
		panic(err)
	}

	printIndexSummary(index)
}

// createIndexEventHandler creates the event handler for index created events.
func createIndexEventHandler(wg *sync.WaitGroup, once *sync.Once) event.EventHandlerFn {
	return func(e event.Event) error {
		// Ensure we only process once (in case of duplicate events)
		once.Do(func() {
			defer wg.Done()

			evt := e.(*indexing.EventFileIndexCreated)
			fmt.Printf("❯ event: received EventFileIndexCreated - IndexID: %s, FileCount: %d\n", evt.IndexID, evt.FileCount)
		})
		return nil
	}
}

// printIndexSummary prints a summary of the index.
func printIndexSummary(index *indexing.Index) {
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

// waitForCompletion waits for event processing to complete or timeout.
func waitForCompletion(ctx context.Context, wg *sync.WaitGroup) {
	fmt.Printf("❯ main: waiting for event processing to complete...\n")
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Printf("❯ main: event processing completed\n")
	case <-ctx.Done():
		fmt.Printf("❯ main: event processing timed out\n")
	}
}
