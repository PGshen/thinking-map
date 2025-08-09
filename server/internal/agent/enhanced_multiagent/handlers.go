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
type StatePreHandler[I any] func(ctx context.Context, input I, state *EnhancedState) (I, error)

// StatePostHandler handles state updates after node execution
type StatePostHandler[O any] func(ctx context.Context, output O, state *EnhancedState) (O, error)

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
func (h *ConversationAnalyzerHandler) PreHandler(ctx context.Context, input []*schema.Message, state *EnhancedState) ([]*schema.Message, error) {
	// Store original messages in state
	state.OriginalMessages = input

	// Build conversation analysis prompt
	prompt := h.buildConversationAnalysisPrompt(input)
	return []*schema.Message{prompt}, nil
}

// PostHandler processes conversation analysis results
func (h *ConversationAnalyzerHandler) PostHandler(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// Parse conversation context from LLM response
	context, err := h.parseConversationContext(output.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse conversation context: %w", err)
	}

	state.ConversationContext = context
	return output, nil
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
func (h *PlanCreationHandler) PreHandler(ctx context.Context, input []*schema.Message, state *EnhancedState) ([]*schema.Message, error) {
	prompt := h.buildPlanCreationPrompt(state)
	return []*schema.Message{prompt}, nil
}

// PostHandler processes plan creation results
func (h *PlanCreationHandler) PostHandler(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// Parse task plan from LLM response
	plan, err := h.parseTaskPlan(output.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task plan: %w", err)
	}

	state.CurrentPlan = plan
	state.PlanHistory = append(state.PlanHistory, plan)
	return output, nil
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
		plan.Steps[i] = &PlanStep{
			ID:                 stepData.ID,
			Name:               stepData.Name,
			Description:        stepData.Description,
			AssignedSpecialist: stepData.AssignedSpecialist,
			Priority:           stepData.Priority,
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
func (h *SpecialistHandler) PreHandler(ctx context.Context, input []*schema.Message, state *EnhancedState) ([]*schema.Message, error) {
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
func (h *SpecialistHandler) PostHandler(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
	// Create step result
	result := &StepResult{
		Success:      true,
		Output:       output,
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
	}

	return output, nil
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
func ResultCollectorLambda(ctx context.Context, input []*schema.Message, state *EnhancedState) (*schema.Message, error) {
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

// PlanExecutionHandler handles plan execution coordination
type PlanExecutionHandler struct {
	config *EnhancedMultiAgentConfig
}

// NewPlanExecutionHandler creates a new plan execution handler
func NewPlanExecutionHandler(config *EnhancedMultiAgentConfig) *PlanExecutionHandler {
	return &PlanExecutionHandler{
		config: config,
	}
}

// Execute coordinates the execution of the current plan
func (h *PlanExecutionHandler) Execute(ctx context.Context, input *schema.Message, state *EnhancedState) (*schema.Message, error) {
	if state.CurrentPlan == nil {
		return nil, fmt.Errorf("no current plan to execute")
	}

	// Find the next step to execute
	nextStep := h.findNextStep(state)
	if nextStep == nil {
		// All steps completed, proceed to result collection
		return &schema.Message{
			Role:    schema.Assistant,
			Content: "Plan execution completed. All steps have been executed.",
		}, nil
	}

	// Mark step as executing
	nextStep.Status = StepStatusRunning
	now := time.Now()

	// Update state
	state.CurrentStep = nextStep.ID

	// Create execution record
	record := &ExecutionRecord{
		StepID:    nextStep.ID,
		Action:    ActionTypeExecute,
		Output:    input,
		StartTime: now,
		Status:    ExecutionStatusStarted,
	}
	state.ExecutionHistory = append(state.ExecutionHistory, record)

	return &schema.Message{
		Role:    schema.Assistant,
		Content: fmt.Sprintf("Executing step: %s - %s", nextStep.Name, nextStep.Description),
	}, nil
}

// findNextStep finds the next step to execute based on dependencies and status
func (h *PlanExecutionHandler) findNextStep(state *EnhancedState) *PlanStep {
	for _, step := range state.CurrentPlan.Steps {
		if step.Status == StepStatusPending {
			// Check if all dependencies are completed
			if h.areDependenciesCompleted(step, state) {
				return step
			}
		}
	}
	return nil
}

// areDependenciesCompleted checks if all dependencies for a step are completed
func (h *PlanExecutionHandler) areDependenciesCompleted(step *PlanStep, state *EnhancedState) bool {
	for _, depID := range step.Dependencies {
		for _, planStep := range state.CurrentPlan.Steps {
			if planStep.ID == depID && planStep.Status != StepStatusCompleted {
				return false
			}
		}
	}
	return true
}

// SpecialistBranchHandler handles specialist selection and branching
type SpecialistBranchHandler struct {
	config *EnhancedMultiAgentConfig
}

// NewSpecialistBranchHandler creates a new specialist branch handler
func NewSpecialistBranchHandler(config *EnhancedMultiAgentConfig) *SpecialistBranchHandler {
	return &SpecialistBranchHandler{
		config: config,
	}
}

// Evaluate determines which specialist should handle the current step
func (h *SpecialistBranchHandler) Evaluate(ctx context.Context, state *EnhancedState) (string, error) {
	if state.CurrentStep == "" {
		return "result_collector", nil // No current step, go to result collection
	}

	// Find the current step by ID
	currentStep := h.findStepByID(state.CurrentStep, state)
	if currentStep == nil {
		return "result_collector", nil // Step not found, go to result collection
	}

	// Return the assigned specialist for the current step
	assignedSpecialist := currentStep.AssignedSpecialist
	if assignedSpecialist == "" {
		return "result_collector", nil // No specialist assigned, go to result collection
	}

	// Verify the specialist exists in config
	for _, specialist := range h.config.Specialists {
		if specialist.Name == assignedSpecialist {
			return assignedSpecialist, nil
		}
	}

	// Specialist not found, go to result collection
	return "result_collector", nil
}

// findStepByID finds a step by its ID in the current plan
func (h *SpecialistBranchHandler) findStepByID(stepID string, state *EnhancedState) *PlanStep {
	if state.CurrentPlan == nil {
		return nil
	}

	for _, step := range state.CurrentPlan.Steps {
		if step.ID == stepID {
			return step
		}
	}
	return nil
}

// buildSpecialistBranchMap creates a map of specialist names to branch conditions
func buildSpecialistBranchMap(specialists []*EnhancedSpecialist) map[string]bool {
	branchMap := make(map[string]bool)

	// Add all specialist names as valid branches
	for _, specialist := range specialists {
		branchMap[specialist.Name] = true
	}

	// Add result collector as a fallback branch
	branchMap["result_collector"] = true

	return branchMap
}
