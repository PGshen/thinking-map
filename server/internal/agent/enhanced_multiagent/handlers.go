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
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
)

// StatePreHandler handles state preparation before node execution
type StatePreHandler func(ctx context.Context, input any, state *EnhancedState) (any, error)

// StatePostHandler handles state updates after node execution
type StatePostHandler func(ctx context.Context, output any, state *EnhancedState) error

// ConversationAnalyzerHandler analyzes conversation context
type ConversationAnalyzerHandler struct {
	config *EnhancedMultiAgentConfig
}

// NewConversationAnalyzerHandler creates a new conversation analyzer handler
func NewConversationAnalyzerHandler(config *EnhancedMultiAgentConfig) *ConversationAnalyzerHandler {
	return &ConversationAnalyzerHandler{
		config: config,
	}
}

// PreHandler prepares input for conversation analysis
func (h *ConversationAnalyzerHandler) PreHandler(ctx context.Context, input any, state *EnhancedState) (any, error) {
	messages, ok := input.([]*schema.Message)
	if !ok {
		return nil, fmt.Errorf("expected []*schema.Message, got %T", input)
	}

	// Store original messages in state
	state.OriginalMessages = messages

	// Build conversation analysis prompt
	prompt := h.buildConversationAnalysisPrompt(messages)
	return []*schema.Message{prompt}, nil
}

// PostHandler processes conversation analysis results
func (h *ConversationAnalyzerHandler) PostHandler(ctx context.Context, output any, state *EnhancedState) error {
	msg, ok := output.(*schema.Message)
	if !ok {
		return fmt.Errorf("expected *schema.Message, got %T", output)
	}

	// Parse conversation context from LLM response
	context, err := h.parseConversationContext(msg.Content)
	if err != nil {
		return fmt.Errorf("failed to parse conversation context: %w", err)
	}

	state.ConversationContext = context
	return nil
}

func (h *ConversationAnalyzerHandler) buildConversationAnalysisPrompt(messages []*schema.Message) *schema.Message {
	prompt := `Analyze the following conversation and extract key information:

Conversation:
`
	for _, msg := range messages {
		prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	prompt += `
Please analyze and provide the following information in JSON format:
{
  "user_intent": "Brief description of what the user wants to achieve",
  "key_topics": ["topic1", "topic2", "topic3"],
  "context_summary": "Summary of the conversation context",
  "complexity": "simple|moderate|complex|very_complex",
  "metadata": {}
}`

	return &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}
}

func (h *ConversationAnalyzerHandler) parseConversationContext(content string) (*ConversationContext, error) {
	var result struct {
		UserIntent     string         `json:"user_intent"`
		KeyTopics      []string       `json:"key_topics"`
		ContextSummary string         `json:"context_summary"`
		Complexity     string         `json:"complexity"`
		Metadata       map[string]any `json:"metadata"`
	}

	err := json.Unmarshal([]byte(content), &result)
	if err != nil {
		return nil, err
	}

	// Convert complexity string to enum
	var complexity TaskComplexity
	switch result.Complexity {
	case "simple":
		complexity = TaskComplexitySimple
	case "moderate":
		complexity = TaskComplexityModerate
	case "complex":
		complexity = TaskComplexityComplex
	case "very_complex":
		complexity = TaskComplexityVeryComplex
	default:
		complexity = TaskComplexityUnknown
	}

	return &ConversationContext{
		UserIntent:     result.UserIntent,
		KeyTopics:      result.KeyTopics,
		ContextSummary: result.ContextSummary,
		Complexity:     complexity,
		Metadata:       result.Metadata,
	}, nil
}

// HostThinkHandler handles host agent thinking
type HostThinkHandler struct {
	config *EnhancedMultiAgentConfig
}

// NewHostThinkHandler creates a new host think handler
func NewHostThinkHandler(config *EnhancedMultiAgentConfig) *HostThinkHandler {
	return &HostThinkHandler{
		config: config,
	}
}

// PreHandler prepares input for host thinking
func (h *HostThinkHandler) PreHandler(ctx context.Context, input any, state *EnhancedState) (any, error) {
	// Build thinking prompt based on conversation context
	prompt := h.buildThinkingPrompt(state)
	return []*schema.Message{prompt}, nil
}

// PostHandler processes thinking results
func (h *HostThinkHandler) PostHandler(ctx context.Context, output any, state *EnhancedState) error {
	msg, ok := output.(*schema.Message)
	if !ok {
		return fmt.Errorf("expected *schema.Message, got %T", output)
	}

	// Create execution record for thinking
	record := &ExecutionRecord{
		StepID:    fmt.Sprintf("think_%d", len(state.ThinkingHistory)+1),
		Action:    ActionTypeThink,
		Output:    msg,
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Status:    ExecutionStatusSuccess,
	}

	state.ThinkingHistory = append(state.ThinkingHistory, record)
	return nil
}

func (h *HostThinkHandler) buildThinkingPrompt(state *EnhancedState) *schema.Message {
	prompt := fmt.Sprintf(`You are an intelligent host agent. Think step by step about how to handle this request.

Conversation Context:
- User Intent: %s
- Key Topics: %v
- Complexity: %s
- Summary: %s

Think about:
1. What is the user really asking for?
2. What information or capabilities do I need?
3. How complex is this task?
4. What approach should I take?

Provide your reasoning in a clear, structured way.`,
		state.ConversationContext.UserIntent,
		state.ConversationContext.KeyTopics,
		state.ConversationContext.Complexity.String(),
		state.ConversationContext.ContextSummary,
	)

	return &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}
}

// ComplexityBranchHandler handles complexity-based branching
type ComplexityBranchHandler struct {
	config *EnhancedMultiAgentConfig
}

// NewComplexityBranchHandler creates a new complexity branch handler
func NewComplexityBranchHandler(config *EnhancedMultiAgentConfig) *ComplexityBranchHandler {
	return &ComplexityBranchHandler{
		config: config,
	}
}

// Evaluate determines the branch based on task complexity
func (h *ComplexityBranchHandler) Evaluate(ctx context.Context, state *EnhancedState) (string, error) {
	if state.ConversationContext == nil {
		return "direct_answer", nil
	}

	switch state.ConversationContext.Complexity {
	case TaskComplexitySimple:
		return "direct_answer", nil
	case TaskComplexityModerate, TaskComplexityComplex, TaskComplexityVeryComplex:
		return "plan_and_execute", nil
	default:
		return "direct_answer", nil
	}
}

// PlanCreationHandler handles task plan creation
type PlanCreationHandler struct {
	config *EnhancedMultiAgentConfig
}

// NewPlanCreationHandler creates a new plan creation handler
func NewPlanCreationHandler(config *EnhancedMultiAgentConfig) *PlanCreationHandler {
	return &PlanCreationHandler{
		config: config,
	}
}

// PreHandler prepares input for plan creation
func (h *PlanCreationHandler) PreHandler(ctx context.Context, input any, state *EnhancedState) (any, error) {
	prompt := h.buildPlanCreationPrompt(state)
	return []*schema.Message{prompt}, nil
}

// PostHandler processes plan creation results
func (h *PlanCreationHandler) PostHandler(ctx context.Context, output any, state *EnhancedState) error {
	msg, ok := output.(*schema.Message)
	if !ok {
		return fmt.Errorf("expected *schema.Message, got %T", output)
	}

	// Parse task plan from LLM response
	plan, err := h.parseTaskPlan(msg.Content)
	if err != nil {
		return fmt.Errorf("failed to parse task plan: %w", err)
	}

	state.CurrentPlan = plan
	state.PlanHistory = append(state.PlanHistory, plan)
	return nil
}

func (h *PlanCreationHandler) buildPlanCreationPrompt(state *EnhancedState) *schema.Message {
	specialistList := ""
	for _, specialist := range h.config.Specialists {
		specialistList += fmt.Sprintf("- %s: %s\n", specialist.Name, specialist.IntendedUse)
	}

	prompt := fmt.Sprintf(`Create a detailed execution plan for the following task.

Task Context:
- User Intent: %s
- Complexity: %s
- Key Topics: %v

Available Specialists:
%s

Create a plan with the following JSON structure:
{
  "id": "unique_plan_id",
  "name": "Plan Name",
  "description": "Plan Description",
  "steps": [
    {
      "id": "step_1",
      "name": "Step Name",
      "description": "Step Description",
      "assigned_specialist": "specialist_name",
      "priority": 1,
      "estimated_time": "5m",
      "dependencies": [],
      "parameters": {}
    }
  ]
}`,
		state.ConversationContext.UserIntent,
		state.ConversationContext.Complexity.String(),
		state.ConversationContext.KeyTopics,
		specialistList,
	)

	return &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}
}

func (h *PlanCreationHandler) parseTaskPlan(content string) (*TaskPlan, error) {
	var planData struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Steps       []struct {
			ID                 string         `json:"id"`
			Name               string         `json:"name"`
			Description        string         `json:"description"`
			AssignedSpecialist string         `json:"assigned_specialist"`
			Priority           int            `json:"priority"`
			EstimatedTime      string         `json:"estimated_time"`
			Dependencies       []string       `json:"dependencies"`
			Parameters         map[string]any `json:"parameters"`
		} `json:"steps"`
	}

	err := json.Unmarshal([]byte(content), &planData)
	if err != nil {
		return nil, err
	}

	// Convert to TaskPlan
	plan := &TaskPlan{
		ID:          planData.ID,
		Version:     1,
		Name:        planData.Name,
		Description: planData.Description,
		Status:      ExecutionStatusStarted,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Steps:       make([]*PlanStep, len(planData.Steps)),
	}

	for i, stepData := range planData.Steps {
		// Parse estimated time
		estimatedTime, _ := time.ParseDuration(stepData.EstimatedTime)

		plan.Steps[i] = &PlanStep{
			ID:                 stepData.ID,
			Name:               stepData.Name,
			Description:        stepData.Description,
			AssignedSpecialist: stepData.AssignedSpecialist,
			Priority:           stepData.Priority,
			EstimatedTime:      estimatedTime,
			Status:             StepStatusPending,
			Dependencies:       stepData.Dependencies,
			Parameters:         stepData.Parameters,
		}
	}

	return plan, nil
}

// SpecialistHandler handles specialist execution
type SpecialistHandler struct {
	specialistName string
}

// NewSpecialistHandler creates a new specialist handler
func NewSpecialistHandler(specialistName string) *SpecialistHandler {
	return &SpecialistHandler{
		specialistName: specialistName,
	}
}

// PreHandler prepares input for specialist execution
func (h *SpecialistHandler) PreHandler(ctx context.Context, input any, state *EnhancedState) (any, error) {
	// Find the current step for this specialist
	currentStep := h.findCurrentStep(state)
	if currentStep == nil {
		return nil, fmt.Errorf("no current step found for specialist %s", h.specialistName)
	}

	// Build specialist prompt
	prompt := h.buildSpecialistPrompt(currentStep, state)
	return []*schema.Message{prompt}, nil
}

// PostHandler processes specialist execution results
func (h *SpecialistHandler) PostHandler(ctx context.Context, output any, state *EnhancedState) error {
	msg, ok := output.(*schema.Message)
	if !ok {
		return fmt.Errorf("expected *schema.Message, got %T", output)
	}

	// Create step result
	result := &StepResult{
		Success:      true,
		Output:       msg,
		Confidence:   0.8, // TODO: implement confidence calculation
		QualityScore: 0.8, // TODO: implement quality scoring
	}

	// Store result in state
	if state.SpecialistResults == nil {
		state.SpecialistResults = make(map[string]*StepResult)
	}
	state.SpecialistResults[h.specialistName] = result

	// Update step status
	currentStep := h.findCurrentStep(state)
	if currentStep != nil {
		currentStep.Status = StepStatusCompleted
		currentStep.Result = result
		now := time.Now()
		currentStep.EndTime = &now
		if currentStep.StartTime != nil {
			currentStep.ActualTime = now.Sub(*currentStep.StartTime)
		}
	}

	return nil
}

func (h *SpecialistHandler) findCurrentStep(state *EnhancedState) *PlanStep {
	if state.CurrentPlan == nil {
		return nil
	}

	for _, step := range state.CurrentPlan.Steps {
		if step.AssignedSpecialist == h.specialistName && step.Status == StepStatusPending {
			return step
		}
	}

	return nil
}

func (h *SpecialistHandler) buildSpecialistPrompt(step *PlanStep, state *EnhancedState) *schema.Message {
	prompt := fmt.Sprintf(`You are a %s specialist. Execute the following step:

Step: %s
Description: %s

Context:
- User Intent: %s
- Overall Plan: %s

Parameters: %v

Please complete this step and provide your result.`,
		h.specialistName,
		step.Name,
		step.Description,
		state.ConversationContext.UserIntent,
		state.CurrentPlan.Description,
		step.Parameters,
	)

	return &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}
}

// ResultCollectorLambda collects and summarizes specialist results
func ResultCollectorLambda(ctx context.Context, input any, state *EnhancedState) (*schema.Message, error) {
	if state.SpecialistResults == nil || len(state.SpecialistResults) == 0 {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: "No specialist results to collect.",
		}, nil
	}

	// Collect all results
	var results []*schema.Message
	for specialistName, result := range state.SpecialistResults {
		if result.Success && result.Output != nil {
			// Add specialist name as context
			msg := &schema.Message{
				Role:    result.Output.Role,
				Content: fmt.Sprintf("[%s]: %s", specialistName, result.Output.Content),
			}
			results = append(results, msg)
		}
	}

	state.CollectedResults = results

	// Create summary
	summary := "Specialist Results Summary:\n\n"
	for _, msg := range results {
		summary += msg.Content + "\n\n"
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: summary,
	}, nil
}
