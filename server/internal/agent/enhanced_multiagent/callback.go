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

package enhanced

import (
	"context"
	"log"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/schema"
)

// EnhancedMultiAgentCallback defines the callback interface for the enhanced multi-agent system
type EnhancedMultiAgentCallback interface {
	// System lifecycle callbacks
	OnSystemStart(ctx context.Context, config *EnhancedMultiAgentConfig) context.Context
	OnSystemEnd(ctx context.Context, result *schema.Message, err error) context.Context

	// Conversation analysis callbacks
	OnConversationAnalysisStart(ctx context.Context, messages []*schema.Message) context.Context
	OnConversationAnalysisEnd(ctx context.Context, context *ConversationContext) context.Context

	// Complexity decision callbacks
	OnComplexityDecision(ctx context.Context, complexity TaskComplexity, reasoning string) context.Context

	// Plan creation callbacks
	OnPlanCreationStart(ctx context.Context, state *EnhancedState) context.Context
	OnPlanCreationEnd(ctx context.Context, plan *TaskPlan) context.Context

	// Plan update callbacks
	OnPlanUpdateStart(ctx context.Context, state *EnhancedState, reason string) context.Context
	OnPlanUpdateEnd(ctx context.Context, state *EnhancedState, oldPlan, newPlan *TaskPlan, update *PlanUpdate) context.Context

	// Execution callbacks
	OnExecutionStart(ctx context.Context, plan *TaskPlan) context.Context
	OnExecutionEnd(ctx context.Context, results map[string]*StepResult) context.Context

	// Specialist execution callbacks
	OnSpecialistExecutionStart(ctx context.Context, specialistName string, step *PlanStep) context.Context
	OnSpecialistExecutionEnd(ctx context.Context, specialistName string, result *StepResult) context.Context
	OnSpecialistExecutionError(ctx context.Context, specialistName string, err error) context.Context

	// Result collection callbacks
	OnResultCollectionStart(ctx context.Context, results map[string]*StepResult) context.Context
	OnResultCollectionEnd(ctx context.Context, collectedResults []*schema.Message) context.Context

	// Feedback processing callbacks
	OnFeedbackProcessingStart(ctx context.Context, results []*schema.Message) context.Context
	OnFeedbackProcessingEnd(ctx context.Context, feedback map[string]any) context.Context

	// Reflection decision callbacks
	OnReflectionDecision(ctx context.Context, branch string, reasoning string) context.Context
	OnContinueExecution(ctx context.Context, state *EnhancedState, reasoning string) context.Context
	OnPlanUpdateDecision(ctx context.Context, state *EnhancedState, reason string) context.Context
	OnExecutionCompletion(ctx context.Context, state *EnhancedState) context.Context

	// State transition callbacks
	OnExecutionStatusChange(ctx context.Context, state *EnhancedState, oldStatus, newStatus ExecutionStatus) context.Context
	OnReflectionCountIncrement(ctx context.Context, state *EnhancedState, count int) context.Context

	// Round control callbacks
	OnRoundStart(ctx context.Context, roundNumber int, state *EnhancedState) context.Context
	OnRoundEnd(ctx context.Context, roundNumber int, state *EnhancedState) context.Context

	// Error handling callbacks
	OnError(ctx context.Context, stage string, err error) context.Context
	OnRecovery(ctx context.Context, stage string, recoveryAction string) context.Context

	// Performance monitoring callbacks
	OnPerformanceMetric(ctx context.Context, metric string, value any, timestamp time.Time) context.Context
}

// DefaultEnhancedCallback provides a default implementation of EnhancedMultiAgentCallback
type DefaultEnhancedCallback struct {
	logger *log.Logger
}

// NewDefaultEnhancedCallback creates a new default callback instance
func NewDefaultEnhancedCallback() *DefaultEnhancedCallback {
	return &DefaultEnhancedCallback{
		logger: log.Default(),
	}
}

// System lifecycle callbacks
func (cb *DefaultEnhancedCallback) OnSystemStart(ctx context.Context, config *EnhancedMultiAgentConfig) context.Context {
	cb.logger.Printf("[SYSTEM] Enhanced Multi-Agent System started: %s", config.Name)
	return ctx
}

func (cb *DefaultEnhancedCallback) OnSystemEnd(ctx context.Context, result *schema.Message, err error) context.Context {
	if err != nil {
		cb.logger.Printf("[SYSTEM] Enhanced Multi-Agent System ended with error: %v", err)
	} else {
		cb.logger.Printf("[SYSTEM] Enhanced Multi-Agent System completed successfully")
	}
	return ctx
}

// Conversation analysis callbacks
func (cb *DefaultEnhancedCallback) OnConversationAnalysisStart(ctx context.Context, messages []*schema.Message) context.Context {
	cb.logger.Printf("[CONVERSATION] Starting conversation analysis with %d messages", len(messages))
	return ctx
}

func (cb *DefaultEnhancedCallback) OnConversationAnalysisEnd(ctx context.Context, context *ConversationContext) context.Context {
	cb.logger.Printf("[CONVERSATION] Conversation analysis completed. Intent: %s, Complexity: %s",
		context.UserIntent, context.Complexity.String())
	return ctx
}

// Thinking callbacks
func (cb *DefaultEnhancedCallback) OnThinkingStart(ctx context.Context, state *EnhancedState) context.Context {
	cb.logger.Printf("[THINKING] Host agent started thinking for round %d", state.RoundNumber)
	return ctx
}

func (cb *DefaultEnhancedCallback) OnThinkingStep(ctx context.Context, step int, thought string) context.Context {
	cb.logger.Printf("[THINKING] Step %d: %s", step, thought)
	return ctx
}

func (cb *DefaultEnhancedCallback) OnThinkingEnd(ctx context.Context, thoughts []*ExecutionRecord) context.Context {
	cb.logger.Printf("[THINKING] Host agent completed thinking with %d thoughts", len(thoughts))
	return ctx
}

// Complexity decision callbacks
func (cb *DefaultEnhancedCallback) OnComplexityDecision(ctx context.Context, complexity TaskComplexity, reasoning string) context.Context {
	cb.logger.Printf("[COMPLEXITY] Task complexity determined as %s: %s", complexity.String(), reasoning)
	return ctx
}

// Plan creation callbacks
func (cb *DefaultEnhancedCallback) OnPlanCreationStart(ctx context.Context, state *EnhancedState) context.Context {
	cb.logger.Printf("[PLANNING] Starting plan creation for round %d", state.RoundNumber)
	return ctx
}

func (cb *DefaultEnhancedCallback) OnPlanCreationEnd(ctx context.Context, plan *TaskPlan) context.Context {
	cb.logger.Printf("[PLANNING] Plan created with %d steps: %s", len(plan.Steps), plan.Name)
	return ctx
}

// Plan update callbacks
func (cb *DefaultEnhancedCallback) OnPlanUpdateStart(ctx context.Context, state *EnhancedState, reason string) context.Context {
	cb.logger.Printf("[PLANNING] Starting plan update: %s", reason)
	return ctx
}

func (cb *DefaultEnhancedCallback) OnPlanUpdateEnd(ctx context.Context, state *EnhancedState, oldPlan, newPlan *TaskPlan, update *PlanUpdate) context.Context {
	cb.logger.Printf("[PLANNING] Plan updated from version %d to %d: %s", oldPlan.Version, newPlan.Version, update.Description)
	return ctx
}

// Execution callbacks
func (cb *DefaultEnhancedCallback) OnExecutionStart(ctx context.Context, plan *TaskPlan) context.Context {
	cb.logger.Printf("[EXECUTION] Starting execution of plan %s with %d steps", plan.Name, len(plan.Steps))
	return ctx
}

func (cb *DefaultEnhancedCallback) OnExecutionEnd(ctx context.Context, results map[string]*StepResult) context.Context {
	cb.logger.Printf("[EXECUTION] Execution completed with %d results", len(results))
	return ctx
}

// Specialist execution callbacks
func (cb *DefaultEnhancedCallback) OnSpecialistExecutionStart(ctx context.Context, specialistName string, step *PlanStep) context.Context {
	cb.logger.Printf("[SPECIALIST] %s started executing step: %s", specialistName, step.Name)
	return ctx
}

func (cb *DefaultEnhancedCallback) OnSpecialistExecutionEnd(ctx context.Context, specialistName string, result *StepResult) context.Context {
	status := "failed"
	if result.Success {
		status = "succeeded"
	}
	cb.logger.Printf("[SPECIALIST] %s %s with confidence %.2f", specialistName, status, result.Confidence)
	return ctx
}

func (cb *DefaultEnhancedCallback) OnSpecialistExecutionError(ctx context.Context, specialistName string, err error) context.Context {
	cb.logger.Printf("[SPECIALIST] %s encountered error: %v", specialistName, err)
	return ctx
}

// Result collection callbacks
func (cb *DefaultEnhancedCallback) OnResultCollectionStart(ctx context.Context, results map[string]*StepResult) context.Context {
	cb.logger.Printf("[COLLECTION] Starting result collection from %d specialists", len(results))
	return ctx
}

func (cb *DefaultEnhancedCallback) OnResultCollectionEnd(ctx context.Context, collectedResults []*schema.Message) context.Context {
	cb.logger.Printf("[COLLECTION] Result collection completed with %d messages", len(collectedResults))
	return ctx
}

// Feedback processing callbacks
func (cb *DefaultEnhancedCallback) OnFeedbackProcessingStart(ctx context.Context, results []*schema.Message) context.Context {
	cb.logger.Printf("[FEEDBACK] Starting feedback processing for %d results", len(results))
	return ctx
}

func (cb *DefaultEnhancedCallback) OnFeedbackProcessingEnd(ctx context.Context, feedback map[string]any) context.Context {
	cb.logger.Printf("[FEEDBACK] Feedback processing completed")
	return ctx
}

// Reflection decision callbacks
func (cb *DefaultEnhancedCallback) OnReflectionDecision(ctx context.Context, branch string, reasoning string) context.Context {
	cb.logger.Printf("[REFLECTION] Decision branch %s: %s", branch, reasoning)
	return ctx
}

func (cb *DefaultEnhancedCallback) OnContinueExecution(ctx context.Context, state *EnhancedState, reasoning string) context.Context {
	cb.logger.Printf("[CONTINUE] Continue execution: %s", reasoning)
	return ctx
}

func (cb *DefaultEnhancedCallback) OnPlanUpdateDecision(ctx context.Context, state *EnhancedState, reason string) context.Context {
	cb.logger.Printf("[PLAN_UPDATE] Plan update decision: %s", reason)
	return ctx
}

func (cb *DefaultEnhancedCallback) OnExecutionCompletion(ctx context.Context, state *EnhancedState) context.Context {
	cb.logger.Printf("[COMPLETION] Execution completed")
	return ctx
}

// State transition callbacks
func (cb *DefaultEnhancedCallback) OnExecutionStatusChange(ctx context.Context, state *EnhancedState, oldStatus, newStatus ExecutionStatus) context.Context {
	cb.logger.Printf("[STATUS] Execution status changed from %s to %s", oldStatus.String(), newStatus.String())
	return ctx
}

func (cb *DefaultEnhancedCallback) OnReflectionCountIncrement(ctx context.Context, state *EnhancedState, count int) context.Context {
	cb.logger.Printf("[REFLECTION] Reflection count incremented to %d", count)
	return ctx
}

// Round control callbacks
func (cb *DefaultEnhancedCallback) OnRoundStart(ctx context.Context, roundNumber int, state *EnhancedState) context.Context {
	cb.logger.Printf("[ROUND] Starting round %d", roundNumber)
	return ctx
}

func (cb *DefaultEnhancedCallback) OnRoundEnd(ctx context.Context, roundNumber int, state *EnhancedState) context.Context {
	cb.logger.Printf("[ROUND] Completed round %d with status %s", roundNumber, state.ExecutionStatus.String())
	return ctx
}

// Error handling callbacks
func (cb *DefaultEnhancedCallback) OnError(ctx context.Context, stage string, err error) context.Context {
	cb.logger.Printf("[ERROR] Error in %s: %v", stage, err)
	return ctx
}

func (cb *DefaultEnhancedCallback) OnRecovery(ctx context.Context, stage string, recoveryAction string) context.Context {
	cb.logger.Printf("[RECOVERY] Recovery in %s: %s", stage, recoveryAction)
	return ctx
}

// Performance monitoring callbacks
func (cb *DefaultEnhancedCallback) OnPerformanceMetric(ctx context.Context, metric string, value any, timestamp time.Time) context.Context {
	cb.logger.Printf("[METRICS] %s: %v at %s", metric, value, timestamp.Format(time.RFC3339))
	return ctx
}

// convertCallbacks converts agent options to callbacks handler
func convertCallbacks(opts ...agent.AgentOption) callbacks.Handler {
	// Extract enhanced multi-agent callbacks from options
	// This will be implemented when we create the options system
	return nil
}

// ConvertCallbackHandlers converts EnhancedMultiAgentCallback instances to callbacks.Handler
func ConvertCallbackHandlers(handlers ...EnhancedMultiAgentCallback) callbacks.Handler {
	if len(handlers) == 0 {
		return nil
	}

	// TODO: Implement proper callback handler conversion
	// This requires deeper integration with the eino callback system
	// For now, return nil and implement later when we have the full context
	return nil
}
