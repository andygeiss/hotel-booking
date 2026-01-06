package outbound

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/agent"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"
)

// IndexSearchToolExecutor implements the agent.ToolExecutor interface.
// It provides the search_index tool that searches indexed files and returns results.
type IndexSearchToolExecutor struct {
	tools           map[string]toolFunc
	indexingService *indexing.IndexingService
	indexID         indexing.IndexID
	verbose         bool
}

// searchIndexToolArgs represents the JSON arguments for search_index.
type searchIndexToolArgs struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// toolFunc is the signature for tool implementations.
type toolFunc func(ctx context.Context, arguments string) (string, error)

// NewIndexSearchToolExecutor creates a new IndexSearchToolExecutor.
// It requires an IndexingService to perform searches and an indexID to identify which index to search.
func NewIndexSearchToolExecutor(indexingService *indexing.IndexingService, indexID indexing.IndexID) *IndexSearchToolExecutor {
	executor := &IndexSearchToolExecutor{
		indexingService: indexingService,
		indexID:         indexID,
		tools:           make(map[string]toolFunc),
		verbose:         false,
	}
	executor.initTools()
	return executor
}

// WithVerbose enables verbose logging of tool executions.
func (e *IndexSearchToolExecutor) WithVerbose(verbose bool) *IndexSearchToolExecutor {
	e.verbose = verbose
	return e
}

// Execute runs the specified tool with the given input arguments.
// It returns the result of the tool execution as a JSON string.
func (e *IndexSearchToolExecutor) Execute(ctx context.Context, toolName string, arguments string) (string, error) {
	tool, ok := e.tools[toolName]
	if !ok {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}
	if e.verbose {
		fmt.Printf("  ↳ tool call: %s(%s)\n", toolName, arguments)
	}
	result, err := tool(ctx, arguments)
	if e.verbose {
		if err != nil {
			fmt.Printf("  ↳ tool error: %v\n", err)
		} else {
			// Truncate result for display
			display := result
			if len(display) > 150 {
				display = display[:150] + "..."
			}
			fmt.Printf("  ↳ tool result: %s\n", display)
		}
	}
	return result, err
}

// GetAvailableTools returns the list of available tool names.
func (e *IndexSearchToolExecutor) GetAvailableTools() []string {
	tools := make([]string, 0, len(e.tools))
	for name := range e.tools {
		tools = append(tools, name)
	}
	return tools
}

// GetToolDefinitions returns the tool definitions for the LLM.
// These definitions tell the LLM what tools are available and how to use them.
func (e *IndexSearchToolExecutor) GetToolDefinitions() []agent.ToolDefinition {
	return []agent.ToolDefinition{
		agent.NewToolDefinition(
			"search_index",
			"Search the indexed files for files matching a query. Returns file paths that match the search query.",
		).WithParameter("query", "The search query to find matching files (e.g., filename, extension, or path component)").
			WithParameter("limit", "Maximum number of results to return (optional, default: 10)"),
	}
}

// HasTool returns true if the specified tool is available.
func (e *IndexSearchToolExecutor) HasTool(toolName string) bool {
	_, ok := e.tools[toolName]
	return ok
}

// initTools registers all available tools.
func (e *IndexSearchToolExecutor) initTools() {
	e.tools["search_index"] = e.searchIndexTool
}

// searchIndexTool implements the search_index tool.
// It searches the index for files matching the given query.
func (e *IndexSearchToolExecutor) searchIndexTool(ctx context.Context, arguments string) (string, error) {
	var args searchIndexToolArgs
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Query == "" {
		return "", errors.New("query is required")
	}

	limit := args.Limit
	if limit <= 0 {
		limit = 10
	}

	results, err := e.indexingService.SearchFiles(ctx, e.indexID, args.Query, limit)
	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	// Convert results to JSON for the LLM
	output, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}

	return string(output), nil
}
