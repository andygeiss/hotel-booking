package agent

import (
	"context"
	"errors"

	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/event"
)

// TaskService represents the service for managing agent tasks.
// It orchestrates the agent loop: observe → decide → act → update.
// The service is responsible for coordinating between the LLM and tool execution.
type TaskService struct {
	executor  ToolExecutor
	llmClient LLMClient
	publisher event.EventPublisher
}

// NewTaskService creates a new TaskService instance.
func NewTaskService(llmClient LLMClient, executor ToolExecutor, publisher event.EventPublisher) *TaskService {
	return &TaskService{
		executor:  executor,
		llmClient: llmClient,
		publisher: publisher,
	}
}

// RunTask executes a task using the agent loop pattern.
// The loop follows: observe → decide → act → update
// It continues until the task is completed, failed, or max iterations reached.
func (s *TaskService) RunTask(ctx context.Context, agent *Agent, task Task) (Result, error) {
	currentTask, err := s.initializeTask(ctx, agent, task)
	if err != nil {
		return Result{}, err
	}

	// Run the agent loop
	for agent.CanContinue() && !currentTask.IsTerminal() {
		if err := s.executeIteration(ctx, agent, currentTask); err != nil {
			return Result{}, err
		}
	}

	return s.finalizeTask(ctx, agent, currentTask)
}

// initializeTask sets up the task for execution.
func (s *TaskService) initializeTask(ctx context.Context, agent *Agent, task Task) (*Task, error) {
	agent.AddTask(task)
	currentTask := agent.GetCurrentTask()
	if currentTask == nil {
		return nil, errors.New("no current task found")
	}

	currentTask.Start()
	agent.ResetIteration()

	if err := s.publishTaskStarted(ctx, agent.ID, currentTask); err != nil {
		return nil, err
	}

	agent.AddMessage(NewMessage(RoleUser, currentTask.Input))
	return currentTask, nil
}

// executeIteration performs one iteration of the agent loop.
func (s *TaskService) executeIteration(ctx context.Context, agent *Agent, currentTask *Task) error {
	agent.IncrementIteration()

	// OBSERVE & DECIDE: Get state and call LLM with available tools
	tools := s.executor.GetToolDefinitions()
	response, err := s.llmClient.Run(ctx, agent.GetMessages(), tools)
	if err != nil {
		return s.handleLLMError(ctx, agent, currentTask, err)
	}

	agent.AddMessage(response.Message)

	// ACT & UPDATE: Execute tools or complete
	if response.HasToolCalls() {
		return s.executeToolCalls(ctx, agent, currentTask, response.ToolCalls)
	}

	currentTask.Complete(response.Message.Content)
	return nil
}

// handleLLMError handles LLM errors during task execution.
func (s *TaskService) handleLLMError(ctx context.Context, agent *Agent, task *Task, err error) error {
	task.Fail(err.Error())
	return s.publishTaskFailed(ctx, agent, task, err.Error())
}

// executeToolCalls executes all tool calls and adds results to conversation.
func (s *TaskService) executeToolCalls(ctx context.Context, agent *Agent, task *Task, toolCalls []ToolCall) error {
	for _, toolCall := range toolCalls {
		tc := NewToolCall(toolCall.ID, toolCall.Name, toolCall.Arguments)
		tc.Execute()

		result, execErr := s.executor.Execute(ctx, toolCall.Name, toolCall.Arguments)
		if execErr != nil {
			tc.Fail(execErr.Error())
		} else {
			tc.Complete(result)
		}

		if err := s.publishToolCallExecuted(ctx, agent.ID, task.ID, &tc); err != nil {
			return err
		}

		agent.AddMessage(tc.ToMessage())
	}
	return nil
}

// finalizeTask handles the completion or failure of a task.
func (s *TaskService) finalizeTask(ctx context.Context, agent *Agent, task *Task) (Result, error) {
	if !task.IsTerminal() {
		errMsg := "max iterations reached"
		task.Fail(errMsg)
		if err := s.publishTaskFailed(ctx, agent, task, errMsg); err != nil {
			return Result{}, err
		}
		return NewResult(task.ID, false, "").WithError(errMsg), nil
	}

	if task.IsCompleted() {
		if err := s.publishTaskCompleted(ctx, agent, task); err != nil {
			return Result{}, err
		}
		return NewResult(task.ID, true, task.Output), nil
	}

	return NewResult(task.ID, false, "").WithError(task.Error), nil
}

// publishTaskCompleted publishes a task completed event.
func (s *TaskService) publishTaskCompleted(ctx context.Context, agent *Agent, task *Task) error {
	evt := NewEventTaskCompleted().
		WithAgentID(agent.ID).
		WithTaskID(task.ID).
		WithIterations(agent.CurrentIteration).
		WithName(task.Name).
		WithOutput(task.Output)
	return s.publisher.Publish(ctx, evt)
}

// publishTaskFailed publishes a task failed event.
func (s *TaskService) publishTaskFailed(ctx context.Context, agent *Agent, task *Task, errMsg string) error {
	evt := NewEventTaskFailed().
		WithAgentID(agent.ID).
		WithTaskID(task.ID).
		WithError(errMsg).
		WithIterations(agent.CurrentIteration).
		WithName(task.Name)
	return s.publisher.Publish(ctx, evt)
}

// publishTaskStarted publishes a task started event.
func (s *TaskService) publishTaskStarted(ctx context.Context, agentID AgentID, task *Task) error {
	evt := NewEventTaskStarted().
		WithAgentID(agentID).
		WithTaskID(task.ID).
		WithName(task.Name)
	return s.publisher.Publish(ctx, evt)
}

// publishToolCallExecuted publishes a tool call executed event.
func (s *TaskService) publishToolCallExecuted(ctx context.Context, agentID AgentID, taskID TaskID, tc *ToolCall) error {
	evt := NewEventToolCallExecuted().
		WithAgentID(agentID).
		WithSuccess(tc.IsCompleted()).
		WithTaskID(taskID).
		WithToolCallID(tc.ID)
	return s.publisher.Publish(ctx, evt)
}
