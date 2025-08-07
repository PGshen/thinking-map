/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package enhanced implements the enhanced multi-agent system with ReAct thinking,
// task planning, and continuous feedback capabilities.
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
	ExecutionStatusStarted
	ExecutionStatusRunning
	ExecutionStatusSuccess
	ExecutionStatusFailed
	ExecutionStatusTimeout
	ExecutionStatusCancelled
)

func (es ExecutionStatus) String() string {
	switch es {
	case ExecutionStatusStarted:
		return "started"
	case ExecutionStatusRunning:
		return "running"
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
	StepID      string                 `json:"step_id"`
	Action      ActionType             `json:"action"`
	Input       []*schema.Message      `json:"input"`
	Output      *schema.Message        `json:"output"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Duration    time.Duration          `json:"duration"`
	Status      ExecutionStatus        `json:"status"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]any         `json:"metadata,omitempty"`
}

// StepResult represents the result of executing a plan step
type StepResult struct {
	Success        bool               `json:"success"`
	Output         *schema.Message    `json:"output"`
	Error          string             `json:"error,omitempty"`
	Confidence     float64            `json:"confidence"`
	QualityScore   float64            `json:"quality_score"`
	Metadata       map[string]any     `json:"metadata,omitempty"`
}

// PlanStep represents a single step in a task plan
type PlanStep struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	AssignedSpecialist string                `json:"assigned_specialist"`
	Priority          int                    `json:"priority"`
	EstimatedTime     time.Duration          `json:"estimated_time"`
	ActualTime        time.Duration          `json:"actual_time"`
	Status            StepStatus             `json:"status"`
	StartTime         *time.Time             `json:"start_time,omitempty"`
	EndTime           *time.Time             `json:"end_time,omitempty"`
	Dependencies      []string               `json:"dependencies,omitempty"`
	Parameters        map[string]any         `json:"parameters,omitempty"`
	Result            *StepResult            `json:"result,omitempty"`
	Metadata          map[string]any         `json:"metadata,omitempty"`
}

// ImpactAssessment represents the impact assessment of a plan update
type ImpactAssessment struct {
	AffectedSteps   []string           `json:"affected_steps"`
	TimeImpact      time.Duration      `json:"time_impact"`
	ResourceImpact  map[string]any     `json:"resource_impact,omitempty"`
	RiskLevel       string             `json:"risk_level"`
	Mitigation      string             `json:"mitigation,omitempty"`
}

// PlanChange represents a specific change in a plan update
type PlanChange struct {
	Type      PlanUpdateType     `json:"type"`
	TargetID  string             `json:"target_id"`
	Field     string             `json:"field"`
	OldValue  any                `json:"old_value,omitempty"`
	NewValue  any                `json:"new_value,omitempty"`
	Metadata  map[string]any     `json:"metadata,omitempty"`
}

// PlanUpdate represents an update to a task plan
type PlanUpdate struct {
	ID               string             `json:"id"`
	PlanVersion      int                `json:"plan_version"`
	UpdateType       PlanUpdateType     `json:"update_type"`
	Description      string             `json:"description"`
	Reason           string             `json:"reason"`
	Changes          []PlanChange       `json:"changes"`
	Timestamp        time.Time          `json:"timestamp"`
	ImpactAssessment *ImpactAssessment  `json:"impact_assessment,omitempty"`
	Metadata         map[string]any     `json:"metadata,omitempty"`
}

// TaskPlan represents a complete task execution plan
type TaskPlan struct {
	ID                string                 `json:"id"`
	Version           int                    `json:"version"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	Status            ExecutionStatus        `json:"status"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	Steps             []*PlanStep            `json:"steps"`
	Dependencies      map[string][]string    `json:"dependencies,omitempty"`
	ResourceAllocation map[string]any        `json:"resource_allocation,omitempty"`
	UpdateHistory     []*PlanUpdate          `json:"update_history,omitempty"`
	Metadata          map[string]any         `json:"metadata,omitempty"`
}

// EnhancedState represents the complete state of the enhanced multi-agent system
type EnhancedState struct {
	// Session Information
	SessionID       string    `json:"session_id"`
	ConversationID  string    `json:"conversation_id"`
	RoundNumber     int       `json:"round_number"`
	StartTime       time.Time `json:"start_time"`

	// Conversation Context
	ConversationContext *ConversationContext `json:"conversation_context,omitempty"`
	OriginalMessages    []*schema.Message    `json:"original_messages"`

	// Thinking History
	ThinkingHistory []*ExecutionRecord `json:"thinking_history,omitempty"`

	// Task Planning
	CurrentPlan     *TaskPlan `json:"current_plan,omitempty"`
	PlanHistory     []*TaskPlan `json:"plan_history,omitempty"`

	// Execution Status
	ExecutionStatus ExecutionStatus `json:"execution_status"`
	CurrentStep     string          `json:"current_step,omitempty"`
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
	MaxRounds       int  `json:"max_rounds"`
	ShouldContinue  bool `json:"should_continue"`
	IsCompleted     bool `json:"is_completed"`

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
	
	var cloned EnhancedState
	err = cloned.FromJSON(data)
	if err != nil {
		return nil, err
	}
	
	return &cloned, nil
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