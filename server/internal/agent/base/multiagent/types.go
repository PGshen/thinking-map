package multiagent

import (
	"context"
	"time"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// TaskComplexity represents the complexity level of a task
type TaskComplexity string

const (
	TaskComplexityUnknown     TaskComplexity = "unknown"
	TaskComplexitySimple      TaskComplexity = "simple"
	TaskComplexityModerate    TaskComplexity = "moderate"
	TaskComplexityComplex     TaskComplexity = "complex"
	TaskComplexityVeryComplex TaskComplexity = "very_complex"
)

// ActionType represents the type of action an agent can take
type ActionType string

const (
	ActionTypeUnknown ActionType = "unknown"
	ActionTypeThink   ActionType = "think"
	ActionTypePlan    ActionType = "plan"
	ActionTypeExecute ActionType = "execute"
	ActionTypeReflect ActionType = "reflect"
	ActionTypeAnswer  ActionType = "answer"
)

// StepStatus represents the status of a plan step
type StepStatus string

const (
	StepStatusUnknown   StepStatus = "unknown"
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// ExecutionStatus represents the overall execution status
type ExecutionStatus string

const (
	ExecutionStatusUnknown    ExecutionStatus = "unknown"
	ExecutionStatusPending    ExecutionStatus = "pending"
	ExecutionStatusAnalyzing  ExecutionStatus = "analyzing"
	ExecutionStatusPlanning   ExecutionStatus = "planning"
	ExecutionStatusStarted    ExecutionStatus = "started"
	ExecutionStatusRunning    ExecutionStatus = "running"
	ExecutionStatusExecuting  ExecutionStatus = "executing"
	ExecutionStatusCollecting ExecutionStatus = "collecting"
	ExecutionStatusCompleted  ExecutionStatus = "completed"
	ExecutionStatusSuccess    ExecutionStatus = "success"
	ExecutionStatusFailed     ExecutionStatus = "failed"
	ExecutionStatusTimeout    ExecutionStatus = "timeout"
	ExecutionStatusCancelled  ExecutionStatus = "cancelled"
)

// PlanUpdateType represents the type of plan update
type PlanUpdateType string

const (
	PlanUpdateTypeUnknown          PlanUpdateType = "unknown"
	PlanUpdateTypeStepAdd          PlanUpdateType = "step_add"
	PlanUpdateTypeStepModify       PlanUpdateType = "step_modify"
	PlanUpdateTypeStepRemove       PlanUpdateType = "step_remove"
	PlanUpdateTypeStepReorder      PlanUpdateType = "step_reorder"
	PlanUpdateTypePriorityChange   PlanUpdateType = "priority_change"
	PlanUpdateTypeDependencyChange PlanUpdateType = "dependency_change"
	PlanUpdateTypeResourceChange   PlanUpdateType = "resource_change"
	PlanUpdateTypeStrategyChange   PlanUpdateType = "strategy_change"
)

// ConversationContext contains conversation analysis results
type ConversationContext struct {
	UserIntent      string            `json:"userIntent"`
	RelevantHistory []*schema.Message `json:"relevantHistory"`
	KeyTopics       []string          `json:"keyTopics"`
	ContextSummary  string            `json:"contextSummary"`
	Complexity      TaskComplexity    `json:"complexity"`
	Metadata        map[string]any    `json:"metadata,omitempty"`
}

// ExecutionRecord represents a single execution step record
type ExecutionRecord struct {
	StepID    string            `json:"stepID"`
	Action    ActionType        `json:"action"`
	Input     []*schema.Message `json:"input"`
	Output    *schema.Message   `json:"output"`
	StartTime time.Time         `json:"startTime"`
	EndTime   time.Time         `json:"endTime"`
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
	QualityScore float64         `json:"qualityScore"`
	Metadata     map[string]any  `json:"metadata,omitempty"`
}

// PlanStep represents a single step in a task plan
type PlanStep struct {
	ID                 string         `json:"id"`
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	AssignedSpecialist string         `json:"assignedSpecialist"`
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
	PlanVersion int            `json:"planVersion"`
	UpdateType  PlanUpdateType `json:"updateType"`
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
	UpdateHistory []*PlanUpdate   `json:"updateHistory,omitempty"`
	Metadata      map[string]any  `json:"metadata,omitempty"`
}

// Feedback represents the feedback received from the user
type Feedback struct {
	ExecutionCompleted bool     `json:"execution_completed"`
	OverallQuality     float64  `json:"overall_quality"`
	PlanNeedsUpdate    bool     `json:"plan_needs_update"`
	Issues             []string `json:"issues"`
	Suggestions        []string `json:"suggestions"`
	Confidence         float64  `json:"confidence"`
	NextActionReason   string   `json:"next_action_reason"`
}

// MultiAgent represents the enhanced multi-agent system
type MultiAgent struct {
	Runnable         compose.Runnable[[]*schema.Message, *schema.Message]
	Graph            *compose.Graph[[]*schema.Message, *schema.Message]
	GraphAddNodeOpts []compose.GraphAddNodeOpt
	AgentOptions     []base.AgentOption
	Config           *MultiAgentConfig
}

// Generate executes the enhanced multi-agent system
func (ema *MultiAgent) Generate(ctx context.Context, input []*schema.Message, opts ...base.AgentOption) (*schema.Message, error) {
	options := base.GetComposeOptions(opts...)
	options = append(options, base.GetComposeOptions(ema.AgentOptions...)...) // 合并option

	return ema.Runnable.Invoke(ctx, input, options...)
}

// Stream executes the enhanced multi-agent system in streaming mode
func (ema *MultiAgent) Stream(ctx context.Context, input []*schema.Message, opts ...base.AgentOption) (*schema.StreamReader[*schema.Message], error) {
	options := base.GetComposeOptions(opts...)
	options = append(options, base.GetComposeOptions(ema.AgentOptions...)...) // 合并option

	return ema.Runnable.Stream(ctx, input, options...)
}

// ExportGraph exports the underlying graph
func (ema *MultiAgent) ExportGraph() (compose.AnyGraph, []compose.GraphAddNodeOpt) {
	return ema.Graph, ema.GraphAddNodeOpts
}

// GetConfig returns the configuration
func (ema *MultiAgent) GetConfig() *MultiAgentConfig {
	return ema.Config
}
