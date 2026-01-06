package outbound_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/agent"
)

func Test_LMStudioClient_Run_With_ValidResponse_Should_ReturnMessage(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"id":      "chatcmpl-123",
			"object":  "chat.completion",
			"created": 1234567890,
			"model":   "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "Hello! How can I help you?",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     10,
				"completion_tokens": 20,
				"total_tokens":      30,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	sut := outbound.NewLMStudioClient(server.URL, "test-model")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
	}
	tools := []agent.ToolDefinition{}
	ctx := context.Background()

	// Act
	response, err := sut.Run(ctx, messages, tools)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "response message role must be assistant", response.Message.Role, agent.RoleAssistant)
	assert.That(t, "response message content must match", response.Message.Content, "Hello! How can I help you?")
	assert.That(t, "finish reason must be stop", response.FinishReason, "stop")
}

func Test_LMStudioClient_Run_With_ToolCalls_Should_ReturnToolCalls(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"id":      "chatcmpl-123",
			"object":  "chat.completion",
			"created": 1234567890,
			"model":   "test-model",
			"choices": []map[string]any{
				{
					"index": 0,
					"message": map[string]any{
						"role":    "assistant",
						"content": "",
						"tool_calls": []map[string]any{
							{
								"id":   "tc-123",
								"type": "function",
								"function": map[string]any{
									"name":      "search",
									"arguments": `{"query":"test"}`,
								},
							},
						},
					},
					"finish_reason": "tool_calls",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	sut := outbound.NewLMStudioClient(server.URL, "test-model")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Search for something"),
	}
	tools := []agent.ToolDefinition{
		agent.NewToolDefinition("search", "Search for items").WithParameter("query", "The search query"),
	}
	ctx := context.Background()

	// Act
	response, err := sut.Run(ctx, messages, tools)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "response must have tool calls", response.HasToolCalls(), true)
	assert.That(t, "tool call count must be 1", len(response.ToolCalls), 1)
	assert.That(t, "tool call name must be search", response.ToolCalls[0].Name, "search")
	assert.That(t, "tool call ID must match", string(response.ToolCalls[0].ID), "tc-123")
}

func Test_LMStudioClient_Run_With_ServerError_Should_ReturnError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	sut := outbound.NewLMStudioClient(server.URL, "test-model")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
	}
	tools := []agent.ToolDefinition{}
	ctx := context.Background()

	// Act
	_, err := sut.Run(ctx, messages, tools)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
}

func Test_LMStudioClient_Run_With_EmptyChoices_Should_ReturnError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"id":      "chatcmpl-123",
			"object":  "chat.completion",
			"created": 1234567890,
			"model":   "test-model",
			"choices": []map[string]any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	sut := outbound.NewLMStudioClient(server.URL, "test-model")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
	}
	tools := []agent.ToolDefinition{}
	ctx := context.Background()

	// Act
	_, err := sut.Run(ctx, messages, tools)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
}

func Test_LMStudioClient_Run_With_ConnectionError_Should_ReturnError(t *testing.T) {
	// Arrange
	sut := outbound.NewLMStudioClient("http://localhost:99999", "test-model")
	messages := []agent.Message{
		agent.NewMessage(agent.RoleUser, "Hello"),
	}
	tools := []agent.ToolDefinition{}
	ctx := context.Background()

	// Act
	_, err := sut.Run(ctx, messages, tools)

	// Assert
	assert.That(t, "err must not be nil", err != nil, true)
}
