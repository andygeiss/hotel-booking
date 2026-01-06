package agent

import "context"

// LLMClient represents the interface for interacting with a Large Language Model.
// It abstracts the communication with external LLM services (e.g., LM Studio, OpenAI).
// The domain layer uses this interface without knowing the implementation details.
type LLMClient interface {
	// Run sends the conversation messages to the LLM and returns the response.
	// The messages should include the system prompt, user messages, assistant responses,
	// and any tool call results from previous iterations.
	// The tools parameter provides the available tool definitions for the LLM to use.
	Run(ctx context.Context, messages []Message, tools []ToolDefinition) (LLMResponse, error)
}

// ToolExecutor represents the interface for executing tool calls.
// Tools are functions that the agent can use to interact with the outside world.
// The adapter layer implements this interface with actual tool implementations.
type ToolExecutor interface {
	// Execute runs the specified tool with the given input arguments.
	// It returns the result of the tool execution as a string.
	Execute(ctx context.Context, toolName string, arguments string) (string, error)

	// GetAvailableTools returns the list of available tool names.
	GetAvailableTools() []string

	// GetToolDefinitions returns the tool definitions for the LLM.
	// These definitions tell the LLM what tools are available and how to use them.
	GetToolDefinitions() []ToolDefinition

	// HasTool returns true if the specified tool is available.
	HasTool(toolName string) bool
}

// ToolDefinition represents the definition of a tool that can be used by the LLM.
// This is used to inform the LLM about available tools and their capabilities.
type ToolDefinition struct {
	Parameters  map[string]string `json:"parameters"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
}

// NewToolDefinition creates a new tool definition.
func NewToolDefinition(name, description string) ToolDefinition {
	return ToolDefinition{
		Name:        name,
		Description: description,
		Parameters:  make(map[string]string),
	}
}

// WithParameter adds a parameter to the tool definition.
func (td ToolDefinition) WithParameter(name, description string) ToolDefinition {
	td.Parameters[name] = description
	return td
}
