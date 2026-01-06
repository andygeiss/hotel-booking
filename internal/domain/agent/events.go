package agent

// This file contains the event types for the agent domain.

const (
	EventTopicAgentLoopFinished = "agent.loop_finished"
	EventTopicTaskCompleted     = "agent.task_completed"
	EventTopicTaskFailed        = "agent.task_failed"
	EventTopicTaskStarted       = "agent.task_started"
	EventTopicToolCallExecuted  = "agent.tool_call_executed"
)

// EventTaskStarted represents a task started event.
type EventTaskStarted struct {
	AgentID AgentID `json:"agent_id"`
	Name    string  `json:"name"`
	TaskID  TaskID  `json:"task_id"`
}

// NewEventTaskStarted creates a new EventTaskStarted instance.
func NewEventTaskStarted() *EventTaskStarted {
	return &EventTaskStarted{}
}

// Topic returns the topic for the event.
func (e *EventTaskStarted) Topic() string {
	return EventTopicTaskStarted
}

// WithAgentID sets the AgentID field.
func (e *EventTaskStarted) WithAgentID(id AgentID) *EventTaskStarted {
	e.AgentID = id
	return e
}

// WithName sets the Name field.
func (e *EventTaskStarted) WithName(name string) *EventTaskStarted {
	e.Name = name
	return e
}

// WithTaskID sets the TaskID field.
func (e *EventTaskStarted) WithTaskID(id TaskID) *EventTaskStarted {
	e.TaskID = id
	return e
}

// EventTaskCompleted represents a task completed event.
type EventTaskCompleted struct {
	Name       string  `json:"name"`
	Output     string  `json:"output"`
	AgentID    AgentID `json:"agent_id"`
	TaskID     TaskID  `json:"task_id"`
	Iterations int     `json:"iterations"`
}

// NewEventTaskCompleted creates a new EventTaskCompleted instance.
func NewEventTaskCompleted() *EventTaskCompleted {
	return &EventTaskCompleted{}
}

// Topic returns the topic for the event.
func (e *EventTaskCompleted) Topic() string {
	return EventTopicTaskCompleted
}

// WithAgentID sets the AgentID field.
func (e *EventTaskCompleted) WithAgentID(id AgentID) *EventTaskCompleted {
	e.AgentID = id
	return e
}

// WithIterations sets the Iterations field.
func (e *EventTaskCompleted) WithIterations(iterations int) *EventTaskCompleted {
	e.Iterations = iterations
	return e
}

// WithName sets the Name field.
func (e *EventTaskCompleted) WithName(name string) *EventTaskCompleted {
	e.Name = name
	return e
}

// WithOutput sets the Output field.
func (e *EventTaskCompleted) WithOutput(output string) *EventTaskCompleted {
	e.Output = output
	return e
}

// WithTaskID sets the TaskID field.
func (e *EventTaskCompleted) WithTaskID(id TaskID) *EventTaskCompleted {
	e.TaskID = id
	return e
}

// EventTaskFailed represents a task failed event.
type EventTaskFailed struct {
	AgentID    AgentID `json:"agent_id"`
	Error      string  `json:"error"`
	Name       string  `json:"name"`
	TaskID     TaskID  `json:"task_id"`
	Iterations int     `json:"iterations"`
}

// NewEventTaskFailed creates a new EventTaskFailed instance.
func NewEventTaskFailed() *EventTaskFailed {
	return &EventTaskFailed{}
}

// Topic returns the topic for the event.
func (e *EventTaskFailed) Topic() string {
	return EventTopicTaskFailed
}

// WithAgentID sets the AgentID field.
func (e *EventTaskFailed) WithAgentID(id AgentID) *EventTaskFailed {
	e.AgentID = id
	return e
}

// WithError sets the Error field.
func (e *EventTaskFailed) WithError(err string) *EventTaskFailed {
	e.Error = err
	return e
}

// WithIterations sets the Iterations field.
func (e *EventTaskFailed) WithIterations(iterations int) *EventTaskFailed {
	e.Iterations = iterations
	return e
}

// WithName sets the Name field.
func (e *EventTaskFailed) WithName(name string) *EventTaskFailed {
	e.Name = name
	return e
}

// WithTaskID sets the TaskID field.
func (e *EventTaskFailed) WithTaskID(id TaskID) *EventTaskFailed {
	e.TaskID = id
	return e
}

// EventToolCallExecuted represents a tool call executed event.
type EventToolCallExecuted struct {
	AgentID    AgentID    `json:"agent_id"`
	TaskID     TaskID     `json:"task_id"`
	ToolCallID ToolCallID `json:"tool_call_id"`
	ToolName   string     `json:"tool_name"`
	Success    bool       `json:"success"`
}

// NewEventToolCallExecuted creates a new EventToolCallExecuted instance.
func NewEventToolCallExecuted() *EventToolCallExecuted {
	return &EventToolCallExecuted{}
}

// Topic returns the topic for the event.
func (e *EventToolCallExecuted) Topic() string {
	return EventTopicToolCallExecuted
}

// WithAgentID sets the AgentID field.
func (e *EventToolCallExecuted) WithAgentID(id AgentID) *EventToolCallExecuted {
	e.AgentID = id
	return e
}

// WithSuccess sets the Success field.
func (e *EventToolCallExecuted) WithSuccess(success bool) *EventToolCallExecuted {
	e.Success = success
	return e
}

// WithTaskID sets the TaskID field.
func (e *EventToolCallExecuted) WithTaskID(id TaskID) *EventToolCallExecuted {
	e.TaskID = id
	return e
}

// WithToolCallID sets the ToolCallID field.
func (e *EventToolCallExecuted) WithToolCallID(id ToolCallID) *EventToolCallExecuted {
	e.ToolCallID = id
	return e
}

// WithToolName sets the ToolName field.
func (e *EventToolCallExecuted) WithToolName(name string) *EventToolCallExecuted {
	e.ToolName = name
	return e
}

// EventAgentLoopFinished represents an agent loop finished event.
type EventAgentLoopFinished struct {
	AgentID         AgentID `json:"agent_id"`
	TaskID          TaskID  `json:"task_id"`
	Reason          string  `json:"reason"`
	Success         bool    `json:"success"`
	TotalIterations int     `json:"total_iterations"`
}

// NewEventAgentLoopFinished creates a new EventAgentLoopFinished instance.
func NewEventAgentLoopFinished() *EventAgentLoopFinished {
	return &EventAgentLoopFinished{}
}

// Topic returns the topic for the event.
func (e *EventAgentLoopFinished) Topic() string {
	return EventTopicAgentLoopFinished
}

// WithAgentID sets the AgentID field.
func (e *EventAgentLoopFinished) WithAgentID(id AgentID) *EventAgentLoopFinished {
	e.AgentID = id
	return e
}

// WithReason sets the Reason field.
func (e *EventAgentLoopFinished) WithReason(reason string) *EventAgentLoopFinished {
	e.Reason = reason
	return e
}

// WithSuccess sets the Success field.
func (e *EventAgentLoopFinished) WithSuccess(success bool) *EventAgentLoopFinished {
	e.Success = success
	return e
}

// WithTaskID sets the TaskID field.
func (e *EventAgentLoopFinished) WithTaskID(id TaskID) *EventAgentLoopFinished {
	e.TaskID = id
	return e
}

// WithTotalIterations sets the TotalIterations field.
func (e *EventAgentLoopFinished) WithTotalIterations(total int) *EventAgentLoopFinished {
	e.TotalIterations = total
	return e
}
