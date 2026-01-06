package agent_test

import (
	"context"
	"errors"
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/agent"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/event"
)

func Test_TaskService_RunTask_With_DirectCompletion_Should_Succeed(t *testing.T) {
	// Arrange
	mockLLM := &mockLLMClient{
		response: agent.NewLLMResponse(
			agent.NewMessage(agent.RoleAssistant, "Here is the answer"),
			"stop",
		),
	}
	mockExecutor := &mockToolExecutor{}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher)

	ag := agent.NewAgent("agent-1", "You are helpful")
	task := agent.NewTask("task-1", "Answer Question", "What is 2+2?")
	ctx := context.Background()

	// Act
	result, err := sut.RunTask(ctx, &ag, task)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "result must be successful", result.Success, true)
	assert.That(t, "result output must match", result.Output, "Here is the answer")
}

func Test_TaskService_RunTask_With_ToolCall_Should_ExecuteToolAndComplete(t *testing.T) {
	// Arrange
	callCount := 0
	mockLLM := &mockLLMClient{
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			callCount++
			if callCount == 1 {
				return agent.NewLLMResponse(
					agent.NewMessage(agent.RoleAssistant, ""),
					"tool_calls",
				).WithToolCalls([]agent.ToolCall{
					agent.NewToolCall("tc-1", "search", `{"query":"test"}`),
				})
			}
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, "Based on the search: answer"),
				"stop",
			)
		},
	}
	mockExecutor := &mockToolExecutor{
		result: "search result",
	}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher)

	ag := agent.NewAgent("agent-1", "You are helpful")
	task := agent.NewTask("task-1", "Search Task", "Find something")
	ctx := context.Background()

	// Act
	result, err := sut.RunTask(ctx, &ag, task)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "result must be successful", result.Success, true)
	assert.That(t, "tool executor must be called", mockExecutor.called, true)
	assert.That(t, "LLM must be called twice", callCount, 2)
}

func Test_TaskService_RunTask_With_MaxIterations_Should_Fail(t *testing.T) {
	// Arrange
	mockLLM := &mockLLMClient{
		responseFn: func(_ []agent.Message) agent.LLMResponse {
			return agent.NewLLMResponse(
				agent.NewMessage(agent.RoleAssistant, ""),
				"tool_calls",
			).WithToolCalls([]agent.ToolCall{
				agent.NewToolCall("tc-1", "loop_tool", `{}`),
			})
		},
	}
	mockExecutor := &mockToolExecutor{result: "loop result"}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher)

	ag := agent.NewAgent("agent-1", "prompt")
	ag.WithMaxIterations(3)
	task := agent.NewTask("task-1", "Infinite Loop", "loop")
	ctx := context.Background()

	// Act
	result, err := sut.RunTask(ctx, &ag, task)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "result must not be successful", result.Success, false)
	assert.That(t, "error must indicate max iterations", result.Error, "max iterations reached")
}

func Test_TaskService_RunTask_With_LLMError_Should_FailTask(t *testing.T) {
	// Arrange
	mockLLM := &mockLLMClient{
		err: errors.New("LLM connection failed"),
	}
	mockExecutor := &mockToolExecutor{}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher)

	ag := agent.NewAgent("agent-1", "prompt")
	task := agent.NewTask("task-1", "Fail Task", "input")
	ctx := context.Background()

	// Act
	result, err := sut.RunTask(ctx, &ag, task)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "result must not be successful", result.Success, false)
	assert.That(t, "error must contain LLM error", result.Error, "LLM connection failed")
}

func Test_TaskService_RunTask_Should_PublishEvents(t *testing.T) {
	// Arrange
	mockLLM := &mockLLMClient{
		response: agent.NewLLMResponse(
			agent.NewMessage(agent.RoleAssistant, "done"),
			"stop",
		),
	}
	mockExecutor := &mockToolExecutor{}
	mockPublisher := &mockEventPublisher{}
	sut := agent.NewTaskService(mockLLM, mockExecutor, mockPublisher)

	ag := agent.NewAgent("agent-1", "prompt")
	task := agent.NewTask("task-1", "Task", "input")
	ctx := context.Background()

	// Act
	_, err := sut.RunTask(ctx, &ag, task)

	// Assert
	assert.That(t, "err must be nil", err == nil, true)
	assert.That(t, "events must be published", len(mockPublisher.events) >= 2, true)
}

// Mock implementations

type mockLLMClient struct {
	err        error
	responseFn func(messages []agent.Message) agent.LLMResponse
	response   agent.LLMResponse
}

func (m *mockLLMClient) Run(_ context.Context, messages []agent.Message, _ []agent.ToolDefinition) (agent.LLMResponse, error) {
	if m.err != nil {
		return agent.LLMResponse{}, m.err
	}
	if m.responseFn != nil {
		return m.responseFn(messages), nil
	}
	return m.response, nil
}

type mockToolExecutor struct {
	err    error
	result string
	called bool
}

func (m *mockToolExecutor) Execute(_ context.Context, _ string, _ string) (string, error) {
	m.called = true
	if m.err != nil {
		return "", m.err
	}
	return m.result, nil
}

func (m *mockToolExecutor) GetAvailableTools() []string {
	return []string{"search", "loop_tool"}
}

func (m *mockToolExecutor) GetToolDefinitions() []agent.ToolDefinition {
	return []agent.ToolDefinition{
		agent.NewToolDefinition("search", "Search for items").WithParameter("query", "The search query"),
		agent.NewToolDefinition("loop_tool", "A tool that loops"),
	}
}

func (m *mockToolExecutor) HasTool(_ string) bool {
	return true
}

type mockEventPublisher struct {
	events []event.Event
}

func (m *mockEventPublisher) Publish(_ context.Context, e event.Event) error {
	m.events = append(m.events, e)
	return nil
}
