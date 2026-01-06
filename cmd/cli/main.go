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
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/agent"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/event"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"
)

// getEnvOrDefault returns the environment variable value or default if not set.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// runAgentAnalysis runs the agent to analyze the created index.
func runAgentAnalysis(ctx context.Context, taskService *agent.TaskService, evt *indexing.EventFileIndexCreated) {
	fmt.Printf("❯ agent: starting agent loop for index analysis...\n")

	// Create agent with system prompt that explains available tools
	agentInstance := agent.NewAgent(
		agent.AgentID("agent-"+string(evt.IndexID)),
		`You are a helpful assistant that analyzes file indexes. You have access to the search_index tool.
When given information about indexed files, use the search_index tool to find relevant files and provide a summary.
To search for files, call the search_index tool with a query parameter.
Example: to find Go files, search for ".go" or specific filenames.`,
	)
	agentInstance.WithMaxIterations(5)

	task := agent.NewTask(
		agent.TaskID("task-analyze-"+string(evt.IndexID)),
		"analyze-index",
		fmt.Sprintf("Analyze the file index with ID '%s' containing %d files. Use the search_index tool to find interesting files and provide a brief summary of the project structure.", evt.IndexID, evt.FileCount),
	)

	result, runErr := taskService.RunTask(ctx, &agentInstance, task)
	if runErr != nil {
		fmt.Printf("❯ agent: error running task - %v\n", runErr)
		return
	}

	if result.Success {
		fmt.Printf("❯ agent: task completed successfully\n")
		fmt.Printf("❯ agent: output - %s\n", result.Output)
	} else {
		fmt.Printf("❯ agent: task failed - %s\n", result.Error)
	}
}

// createAgentEventHandler creates the event handler for index created events.
func createAgentEventHandler(
	ctx context.Context,
	taskService *agent.TaskService,
	wg *sync.WaitGroup,
	once *sync.Once,
) event.EventHandlerFn {
	return func(e event.Event) error {
		// Ensure we only process once (in case of duplicate events)
		once.Do(func() {
			defer wg.Done()

			evt := e.(*indexing.EventFileIndexCreated)
			fmt.Printf("❯ event: received EventFileIndexCreated - IndexID: %s, FileCount: %d\n", evt.IndexID, evt.FileCount)

			runAgentAnalysis(ctx, taskService, evt)
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

// setupAgentComponents creates and returns the agent service components.
// It uses the IndexSearchToolExecutor with the indexing service to enable file searching.
func setupAgentComponents(eventPublisher event.EventPublisher, indexingService *indexing.IndexingService, indexID indexing.IndexID) *agent.TaskService {
	lmStudioURL := getEnvOrDefault("LM_STUDIO_URL", "http://localhost:1234")
	lmStudioModel := getEnvOrDefault("LM_STUDIO_MODEL", "default")
	llmClient := outbound.NewLMStudioClient(lmStudioURL, lmStudioModel)

	// Create the tool executor with search_index capability
	toolExecutor := outbound.NewIndexSearchToolExecutor(indexingService, indexID)

	return agent.NewTaskService(llmClient, toolExecutor, eventPublisher)
}

// waitForAgentCompletion waits for the agent to complete or timeout.
func waitForAgentCompletion(ctx context.Context, wg *sync.WaitGroup) {
	fmt.Printf("❯ main: waiting for agent to complete...\n")
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Printf("❯ main: agent completed\n")
	case <-ctx.Done():
		fmt.Printf("❯ main: agent timed out\n")
	}
}

func main() {
	// Create a context with timeout for the agent task
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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
	// This must be created before the agent components so the tool executor can use it.
	indexingService := indexing.NewIndexingService(fileReader, indexRepository, eventPublisher)

	// Use the service to create an index (this will also publish the event).
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Create the index ID from the working directory
	indexID := indexing.IndexID(wd)

	// Setup the agent components with the indexing service for search capability.
	taskService := setupAgentComponents(eventPublisher, indexingService, indexID)

	// WaitGroup and Once to coordinate agent completion and prevent duplicate processing
	var wg sync.WaitGroup
	var once sync.Once
	wg.Add(1)

	// Subscribe to the EventFileIndexCreated event to start the agent loop.
	err = eventSubscriber.Subscribe(
		ctx,
		indexing.EventTopicFileIndexCreated,
		func() event.Event { return indexing.NewEventFileIndexCreated() },
		createAgentEventHandler(ctx, taskService, &wg, &once),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("❯ main: creating index for path: %s\n", wd)

	createErr := indexingService.CreateIndex(ctx, wd)
	if createErr != nil {
		panic(createErr)
	}

	// Wait for agent task to complete
	waitForAgentCompletion(ctx, &wg)

	// Read the index back from the repository to demonstrate the full cycle.
	index, err := indexRepository.Read(context.Background(), indexID)
	if err != nil {
		panic(err)
	}

	printIndexSummary(index)
}
