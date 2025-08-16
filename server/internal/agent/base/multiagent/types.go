package multiagent

import (
	"context"
	"time"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/cloudwego/eino/compose"
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

// MultiAgent represents the enhanced multi-agent system
type MultiAgent struct {
	Runnable         compose.Runnable[[]*schema.Message, *schema.Message]
	Graph            *compose.Graph[[]*schema.Message, *schema.Message]
	GraphAddNodeOpts []compose.GraphAddNodeOpt
	Config           *MultiAgentConfig
}

// Generate executes the enhanced multi-agent system
func (ema *MultiAgent) Generate(ctx context.Context, input []*schema.Message, opts ...base.AgentOption) (*schema.Message, error) {
	composeOptions := base.GetComposeOptions(opts...)

	// TODO: implement convertCallbacks function
	// handler := convertCallbacks(opts...)
	// if handler != nil {
	//	composeOptions = append(composeOptions, compose.WithCallbacks(handler))
	// }

	return ema.Runnable.Invoke(ctx, input, composeOptions...)
}

// Stream executes the enhanced multi-agent system in streaming mode
func (ema *MultiAgent) Stream(ctx context.Context, input []*schema.Message, opts ...base.AgentOption) (*schema.StreamReader[*schema.Message], error) {
	composeOptions := base.GetComposeOptions(opts...)

	// TODO: implement convertCallbacks function
	// handler := convertCallbacks(opts...)
	// if handler != nil {
	//	composeOptions = append(composeOptions, compose.WithCallbacks(handler))
	// }

	return ema.Runnable.Stream(ctx, input, composeOptions...)
}

// ExportGraph exports the underlying graph
func (ema *MultiAgent) ExportGraph() (compose.AnyGraph, []compose.GraphAddNodeOpt) {
	return ema.Graph, ema.GraphAddNodeOpts
}

// GetConfig returns the configuration
func (ema *MultiAgent) GetConfig() *MultiAgentConfig {
	return ema.Config
}
