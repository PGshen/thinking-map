package enhanced

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"
)

// TaskComplexity represents the complexity level of a task
type TaskComplexity int

const (
	TaskComplexityUnknown TaskComplexity = iota
	TaskComplexitySimple
	TaskComplexityModerate
	TaskComplexityComplex
	TaskComplexityVeryComplex
)

func (tc TaskComplexity) String() string {
	switch tc {
	case TaskComplexitySimple:
		return "simple"
	case TaskComplexityModerate:
		return "moderate"
	case TaskComplexityComplex:
		return "complex"
	case TaskComplexityVeryComplex:
		return "very_complex"
	default:
		return "unknown"
	}
}

// ActionType represents the type of action an agent can take
type ActionType int

const (
	ActionTypeUnknown ActionType = iota
	ActionTypeThink
	ActionTypePlan
	ActionTypeExecute
	ActionTypeReflect
	ActionTypeAnswer
)

func (at ActionType) String() string {
	switch at {
	case ActionTypeThink:
		return "think"
	case ActionTypePlan:
		return "plan"
	case ActionTypeExecute:
		return "execute"
	case ActionTypeReflect:
		return "reflect"
	case ActionTypeAnswer:
		return "answer"
	default:
		return "unknown"
	}
}

// StepStatus represents the status of a plan step
type StepStatus int

const (
	StepStatusUnknown StepStatus = iota
	StepStatusPending
	StepStatusRunning
	StepStatusCompleted
	StepStatusFailed
	StepStatusSkipped
)

func (ss StepStatus) String() string {
	switch ss {
	case StepStatusPending:
		return "pending"
	case StepStatusRunning:
		return "running"
	case StepStatusCompleted:
		return "completed"
	case StepStatusFailed:
		return "failed"
	case StepStatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// ExecutionStatus represents the overall execution status
type ExecutionStatus int

const (
	ExecutionStatusUnknown ExecutionStatus = iota
	ExecutionStatusPending
	ExecutionStatusAnalyzing
	ExecutionStatusPlanning
	ExecutionStatusStarted
	ExecutionStatusRunning
	ExecutionStatusExecuting
	ExecutionStatusCollecting
	ExecutionStatusCompleted
	ExecutionStatusSuccess
	ExecutionStatusFailed
	ExecutionStatusTimeout
	ExecutionStatusCancelled
)

func (es ExecutionStatus) String() string {
	switch es {
	case ExecutionStatusPending:
		return "pending"
	case ExecutionStatusAnalyzing:
		return "analyzing"
	case ExecutionStatusPlanning:
		return "planning"
	case ExecutionStatusStarted:
		return "started"
	case ExecutionStatusRunning:
		return "running"
	case ExecutionStatusExecuting:
		return "executing"
	case ExecutionStatusCollecting:
		return "collecting"
	case ExecutionStatusCompleted:
		return "completed"
	case ExecutionStatusSuccess:
		return "success"
	case ExecutionStatusFailed:
		return "failed"
	case ExecutionStatusTimeout:
		return "timeout"
	case ExecutionStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// PlanUpdateType represents the type of plan update
type PlanUpdateType int

const (
	PlanUpdateTypeUnknown PlanUpdateType = iota
	PlanUpdateTypeStepAdd
	PlanUpdateTypeStepModify
	PlanUpdateTypeStepRemove
	PlanUpdateTypeStepReorder
	PlanUpdateTypePriorityChange
	PlanUpdateTypeDependencyChange
	PlanUpdateTypeResourceChange
	PlanUpdateTypeStrategyChange
)

func (put PlanUpdateType) String() string {
	switch put {
	case PlanUpdateTypeStepAdd:
		return "step_add"
	case PlanUpdateTypeStepModify:
		return "step_modify"
	case PlanUpdateTypeStepRemove:
		return "step_remove"
	case PlanUpdateTypeStepReorder:
		return "step_reorder"
	case PlanUpdateTypePriorityChange:
		return "priority_change"
	case PlanUpdateTypeDependencyChange:
		return "dependency_change"
	case PlanUpdateTypeResourceChange:
		return "resource_change"
	case PlanUpdateTypeStrategyChange:
		return "strategy_change"
	default:
		return "unknown"
	}
}

// ConversationContext contains conversation analysis results
type ConversationContext struct {
	UserIntent      string            `json:"user_intent"`
	RelevantHistory []*schema.Message `json:"relevant_history"`
	KeyTopics       []string          `json:"key_topics"`
	ContextSummary  string            `json:"context_summary"`
	Complexity      TaskComplexity    `json:"complexity"`
	Metadata        map[string]any    `json:"metadata,omitempty"`
}

// ExecutionRecord represents a single execution step record
type ExecutionRecord struct {
	StepID    string            `json:"step_id"`
	Action    ActionType        `json:"action"`
	Input     []*schema.Message `json:"input"`
	Output    *schema.Message   `json:"output"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time"`
	Duration  time.Duration     `json:"duration"`
	Status    ExecutionStatus   `json:"status"`
	Error     string            `json:"error,omitempty"`
	Metadata  map[string]any    `json:"metadata,omitempty"`
}

// StepResult represents the result of executing a plan step
type StepResult struct {
	Success      bool            `json:"success"`
	Output       *schema.Message `json:"output"`
	Error        string          `json:"error,omitempty"`
	Confidence   float64         `json:"confidence"`
	QualityScore float64         `json:"quality_score"`
	Metadata     map[string]any  `json:"metadata,omitempty"`
}

// PlanStep represents a single step in a task plan
type PlanStep struct {
	ID                 string         `json:"id"`
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	AssignedSpecialist string         `json:"assigned_specialist"`
	Priority           int            `json:"priority"`
	Status             StepStatus     `json:"status"`
	Dependencies       []string       `json:"dependencies,omitempty"`
	Parameters         map[string]any `json:"parameters,omitempty"`
	Result             *StepResult    `json:"result,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

// PlanUpdate represents an update to a task plan
type PlanUpdate struct {
	ID          string         `json:"id"`
	PlanVersion int            `json:"plan_version"`
	UpdateType  PlanUpdateType `json:"update_type"`
	Description string         `json:"description"`
	Reason      string         `json:"reason"`
	Timestamp   time.Time      `json:"timestamp"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// TaskPlan represents a complete task execution plan
type TaskPlan struct {
	ID            string          `json:"id"`
	Version       int             `json:"version"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Status        ExecutionStatus `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Steps         []*PlanStep     `json:"steps"`
	UpdateHistory []*PlanUpdate   `json:"update_history,omitempty"`
	Metadata      map[string]any  `json:"metadata,omitempty"`
}

// EnhancedState represents the complete state of the enhanced multi-agent system
type EnhancedState struct {
	RoundNumber int       `json:"round_number"`
	StartTime   time.Time `json:"start_time"`

	// Conversation Context
	ConversationContext *ConversationContext `json:"conversation_context,omitempty"`
	OriginalMessages    []*schema.Message    `json:"original_messages"`

	// Task Planning
	CurrentPlan *TaskPlan   `json:"current_plan,omitempty"`
	PlanHistory []*TaskPlan `json:"plan_history,omitempty"`

	// Execution Status
	ExecutionStatus  ExecutionStatus    `json:"execution_status"`
	CurrentStep      string             `json:"current_step,omitempty"`
	ExecutionHistory []*ExecutionRecord `json:"execution_history,omitempty"`

	// Specialist Results
	SpecialistResults map[string]*StepResult `json:"specialist_results,omitempty"`

	// Collected Results
	CollectedResults []*schema.Message `json:"collected_results,omitempty"`
	FinalResult      *schema.Message   `json:"final_result,omitempty"`

	// Feedback and Reflection
	FeedbackHistory []map[string]any `json:"feedback_history,omitempty"`
	ReflectionCount int              `json:"reflection_count"`

	// Execution Control
	MaxRounds      int  `json:"max_rounds"`
	ShouldContinue bool `json:"should_continue"`
	IsCompleted    bool `json:"is_completed"`

	// Final Answer
	FinalAnswer *schema.Message `json:"final_answer,omitempty"`

	// Metadata
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ToJSON serializes the state to JSON
func (es *EnhancedState) ToJSON() ([]byte, error) {
	return json.Marshal(es)
}

// FromJSON deserializes the state from JSON
func (es *EnhancedState) FromJSON(data []byte) error {
	return json.Unmarshal(data, es)
}

// Clone creates a deep copy of the state
func (es *EnhancedState) Clone() (*EnhancedState, error) {
	data, err := es.ToJSON()
	if err != nil {
		return nil, err
	}

	cloned := &EnhancedState{}
	err = cloned.FromJSON(data)
	if err != nil {
		return nil, err
	}

	return cloned, nil
}

// 基础字段管理方法
func (es *EnhancedState) SetRoundNumber(round int) {
	es.RoundNumber = round
}

func (es *EnhancedState) IncrementRound() {
	es.RoundNumber++
}

func (es *EnhancedState) SetStartTime(t time.Time) {
	es.StartTime = t
}

// 对话上下文管理方法
func (es *EnhancedState) UpdateConversationContext(ctx *ConversationContext) {
	es.ConversationContext = ctx
}

func (es *EnhancedState) SetOriginalMessages(messages []*schema.Message) {
	es.OriginalMessages = messages
}

// 计划管理方法
func (es *EnhancedState) SetCurrentPlan(plan *TaskPlan) {
	es.CurrentPlan = plan
}

func (es *EnhancedState) AddPlanToHistory(plan *TaskPlan) {
	if es.PlanHistory == nil {
		es.PlanHistory = make([]*TaskPlan, 0)
	}
	es.PlanHistory = append(es.PlanHistory, plan)
}

// 执行记录管理方法
func (es *EnhancedState) SetExecutionStatus(status ExecutionStatus) {
	es.ExecutionStatus = status
}

func (es *EnhancedState) SetCurrentStep(stepID string) {
	es.CurrentStep = stepID
}

func (es *EnhancedState) AddExecutionRecord(record *ExecutionRecord) {
	if es.ExecutionHistory == nil {
		es.ExecutionHistory = make([]*ExecutionRecord, 0)
	}
	es.ExecutionHistory = append(es.ExecutionHistory, record)
}

// 专家结果管理方法
func (es *EnhancedState) UpdateSpecialistResult(specialist string, result *StepResult) {
	if es.SpecialistResults == nil {
		es.SpecialistResults = make(map[string]*StepResult)
	}
	es.SpecialistResults[specialist] = result
}

func (es *EnhancedState) ClearSpecialistResults() {
	es.SpecialistResults = make(map[string]*StepResult)
}

// 结果收集管理方法
func (es *EnhancedState) AddCollectedResult(result *schema.Message) {
	if es.CollectedResults == nil {
		es.CollectedResults = make([]*schema.Message, 0)
	}
	es.CollectedResults = append(es.CollectedResults, result)
}

func (es *EnhancedState) SetFinalResult(result *schema.Message) {
	es.FinalResult = result
}

// 反馈管理方法
func (es *EnhancedState) AddFeedback(feedback map[string]any) {
	if es.FeedbackHistory == nil {
		es.FeedbackHistory = make([]map[string]any, 0)
	}
	es.FeedbackHistory = append(es.FeedbackHistory, feedback)
}

func (es *EnhancedState) IncrementReflection() {
	es.ReflectionCount++
}

// 元数据管理方法
func (es *EnhancedState) SetMaxRounds(max int) {
	es.MaxRounds = max
}

func (es *EnhancedState) SetShouldContinue(should bool) {
	es.ShouldContinue = should
}

func (es *EnhancedState) SetCompleted(completed bool) {
	es.IsCompleted = completed
}

func (es *EnhancedState) SetFinalAnswer(answer *schema.Message) {
	es.FinalAnswer = answer
}

func (es *EnhancedState) SetMetadata(key string, value any) {
	if es.Metadata == nil {
		es.Metadata = make(map[string]any)
	}
	es.Metadata[key] = value
}

func (es *EnhancedState) GetMetadata(key string) (any, bool) {
	if es.Metadata == nil {
		return nil, false
	}
	value, exists := es.Metadata[key]
	return value, exists
}

// EnhancedMultiAgent represents the enhanced multi-agent system
type EnhancedMultiAgent struct {
	runnable         compose.Runnable[[]*schema.Message, *schema.Message]
	graph            *compose.Graph[[]*schema.Message, *schema.Message]
	graphAddNodeOpts []compose.GraphAddNodeOpt
	config           *EnhancedMultiAgentConfig
}

// Generate executes the enhanced multi-agent system
func (ema *EnhancedMultiAgent) Generate(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*schema.Message, error) {
	composeOptions := agent.GetComposeOptions(opts...)

	// TODO: implement convertCallbacks function
	// handler := convertCallbacks(opts...)
	// if handler != nil {
	//	composeOptions = append(composeOptions, compose.WithCallbacks(handler))
	// }

	return ema.runnable.Invoke(ctx, input, composeOptions...)
}

// Stream executes the enhanced multi-agent system in streaming mode
func (ema *EnhancedMultiAgent) Stream(ctx context.Context, input []*schema.Message, opts ...agent.AgentOption) (*schema.StreamReader[*schema.Message], error) {
	composeOptions := agent.GetComposeOptions(opts...)

	// TODO: implement convertCallbacks function
	// handler := convertCallbacks(opts...)
	// if handler != nil {
	//	composeOptions = append(composeOptions, compose.WithCallbacks(handler))
	// }

	return ema.runnable.Stream(ctx, input, composeOptions...)
}

// ExportGraph exports the underlying graph
func (ema *EnhancedMultiAgent) ExportGraph() (compose.AnyGraph, []compose.GraphAddNodeOpt) {
	return ema.graph, ema.graphAddNodeOpts
}

// GetConfig returns the configuration
func (ema *EnhancedMultiAgent) GetConfig() *EnhancedMultiAgentConfig {
	return ema.config
}
