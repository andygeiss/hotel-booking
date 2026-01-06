package agent_test

import (
	"testing"

	"github.com/andygeiss/cloud-native-utils/assert"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/agent"
)

func Test_Task_NewTask_With_ValidParams_Should_ReturnPendingTask(t *testing.T) {
	// Arrange
	id := agent.TaskID("task-1")
	name := "Test Task"
	input := "Do something"

	// Act
	task := agent.NewTask(id, name, input)

	// Assert
	assert.That(t, "task ID must match", task.ID, id)
	assert.That(t, "task name must match", task.Name, name)
	assert.That(t, "task input must match", task.Input, input)
	assert.That(t, "task status must be pending", task.Status, agent.TaskStatusPending)
}

func Test_Task_Complete_With_Output_Should_BeCompleted(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Task", "input")
	task.Start()

	// Act
	task.Complete("output result")

	// Assert
	assert.That(t, "task status must be completed", task.Status, agent.TaskStatusCompleted)
	assert.That(t, "task output must match", task.Output, "output result")
}

func Test_Task_Fail_With_Error_Should_BeFailed(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Task", "input")
	task.Start()

	// Act
	task.Fail("something went wrong")

	// Assert
	assert.That(t, "task status must be failed", task.Status, agent.TaskStatusFailed)
	assert.That(t, "task error must match", task.Error, "something went wrong")
}

func Test_Task_IsTerminal_With_CompletedTask_Should_ReturnTrue(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Task", "input")
	task.Complete("done")

	// Act
	isTerminal := task.IsTerminal()

	// Assert
	assert.That(t, "completed task must be terminal", isTerminal, true)
}

func Test_Task_IsTerminal_With_FailedTask_Should_ReturnTrue(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Task", "input")
	task.Fail("error")

	// Act
	isTerminal := task.IsTerminal()

	// Assert
	assert.That(t, "failed task must be terminal", isTerminal, true)
}

func Test_Task_IsTerminal_With_PendingTask_Should_ReturnFalse(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Task", "input")

	// Act
	isTerminal := task.IsTerminal()

	// Assert
	assert.That(t, "pending task must not be terminal", isTerminal, false)
}

func Test_Task_Start_With_PendingTask_Should_BeInProgress(t *testing.T) {
	// Arrange
	task := agent.NewTask("task-1", "Task", "input")

	// Act
	task.Start()

	// Assert
	assert.That(t, "task status must be in progress", task.Status, agent.TaskStatusInProgress)
}

func Test_ToolCall_NewToolCall_With_ValidParams_Should_ReturnPendingToolCall(t *testing.T) {
	// Arrange
	id := agent.ToolCallID("tc-1")
	name := "search"
	args := `{"query": "test"}`

	// Act
	tc := agent.NewToolCall(id, name, args)

	// Assert
	assert.That(t, "tool call ID must match", tc.ID, id)
	assert.That(t, "tool call name must match", tc.Name, name)
	assert.That(t, "tool call arguments must match", tc.Arguments, args)
	assert.That(t, "tool call status must be pending", tc.Status, agent.ToolCallStatusPending)
}

func Test_ToolCall_Complete_With_Result_Should_BeCompleted(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)
	tc.Execute()

	// Act
	tc.Complete("search result")

	// Assert
	assert.That(t, "tool call status must be completed", tc.Status, agent.ToolCallStatusCompleted)
	assert.That(t, "tool call result must match", tc.Result, "search result")
}

func Test_ToolCall_Execute_With_PendingToolCall_Should_BeExecuting(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)

	// Act
	tc.Execute()

	// Assert
	assert.That(t, "tool call status must be executing", tc.Status, agent.ToolCallStatusExecuting)
}

func Test_ToolCall_Fail_With_Error_Should_BeFailed(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)
	tc.Execute()

	// Act
	tc.Fail("tool error")

	// Assert
	assert.That(t, "tool call status must be failed", tc.Status, agent.ToolCallStatusFailed)
	assert.That(t, "tool call error must match", tc.Error, "tool error")
}

func Test_ToolCall_ToMessage_With_CompletedToolCall_Should_ReturnToolMessage(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)
	tc.Complete("result")

	// Act
	msg := tc.ToMessage()

	// Assert
	assert.That(t, "message role must be tool", msg.Role, agent.RoleTool)
	assert.That(t, "message content must be result", msg.Content, "result")
	assert.That(t, "message tool call ID must match", msg.ToolCallID, "tc-1")
}

func Test_ToolCall_ToMessage_With_FailedToolCall_Should_ReturnErrorMessage(t *testing.T) {
	// Arrange
	tc := agent.NewToolCall("tc-1", "search", `{}`)
	tc.Fail("tool failed")

	// Act
	msg := tc.ToMessage()

	// Assert
	assert.That(t, "message role must be tool", msg.Role, agent.RoleTool)
	assert.That(t, "message content must contain error", msg.Content, "Error: tool failed")
}

func Test_SearchIndexToolArgs_NewSearchIndexToolArgs_Should_HaveDefaultLimit(t *testing.T) {
	// Arrange & Act
	args := agent.NewSearchIndexToolArgs("test query")

	// Assert
	assert.That(t, "query must match", args.Query, "test query")
	assert.That(t, "default limit must be 10", args.Limit, 10)
}

func Test_SearchIndexToolArgs_WithLimit_Should_SetLimit(t *testing.T) {
	// Arrange
	args := agent.NewSearchIndexToolArgs("test query")

	// Act
	args = args.WithLimit(5)

	// Assert
	assert.That(t, "limit must be 5", args.Limit, 5)
}

func Test_SearchResult_NewSearchResult_Should_SetFilePath(t *testing.T) {
	// Arrange & Act
	result := agent.NewSearchResult("/path/to/file.go")

	// Assert
	assert.That(t, "file path must match", result.FilePath, "/path/to/file.go")
}

func Test_SearchResult_WithSnippet_Should_SetSnippet(t *testing.T) {
	// Arrange
	result := agent.NewSearchResult("/path/to/file.go")

	// Act
	result = result.WithSnippet("matching content")

	// Assert
	assert.That(t, "snippet must match", result.Snippet, "matching content")
}

func Test_SearchResult_WithScore_Should_SetScore(t *testing.T) {
	// Arrange
	result := agent.NewSearchResult("/path/to/file.go")

	// Act
	result = result.WithScore(0.95)

	// Assert
	assert.That(t, "score must match", result.Score, 0.95)
}
