package multiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/cloudwego/eino/schema"
)

// StatePreHandler handles state preparation before node execution
type StatePreHandler[I any] func(ctx context.Context, input I, state *MultiAgentState) (I, error)

// StatePostHandler handles state updates after node execution
type StatePostHandler[O any] func(ctx context.Context, output O, state *MultiAgentState) (O, error)

// ConversationAnalyzerHandler analyzes conversation context
type ConversationAnalyzerHandler struct {
	config *MultiAgentConfig
}

// NewConversationAnalyzerHandler creates a new conversation analyzer handler
func NewConversationAnalyzerHandler(config *MultiAgentConfig) *ConversationAnalyzerHandler {
	return &ConversationAnalyzerHandler{
		config: config,
	}
}

// PreHandler prepares input for conversation analysis
func (h *ConversationAnalyzerHandler) PreHandler(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
	// Store original messages in state
	state.OriginalMessages = input

	// Set analysis state
	state.SetExecutionStatus(ExecutionStatusAnalyzing)
	// Build conversation analysis prompt
	prompt := h.buildConversationAnalysisPrompt(input)
	return []*schema.Message{prompt}, nil
}

// PostHandler processes conversation analysis results
func (h *ConversationAnalyzerHandler) PostHandler(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
	// Parse conversation context from LLM response
	context, err := h.parseConversationContext(output.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse conversation context: %w", err)
	}

	// Update state using unified method
	state.UpdateConversationContext(context)
	state.SetExecutionStatus(ExecutionStatusPlanning)
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
IMPORTANT: You MUST respond with ONLY a valid JSON object. Do not include any explanations, comments, or additional text before or after the JSON. Your response should start with { and end with }.

Please analyze and provide the following information in JSON format:
{
  "userIntent": "Brief description of what the user wants to achieve",
  "keyTopics": ["topic1", "topic2", "topic3"],
  "contextSummary": "Summary of the conversation context",
  "complexity": "simple|moderate|complex|very_complex",
  "metadata": {}
}

Remember: Output ONLY the JSON object, no other text.`

	return &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}
}

func (h *ConversationAnalyzerHandler) parseConversationContext(content string) (*ConversationContext, error) {
	var result struct {
		UserIntent     string         `json:"userIntent"`
		KeyTopics      []string       `json:"keyTopics"`
		ContextSummary string         `json:"contextSummary"`
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

// ComplexityBranchHandler handles complexity-based branching
type ComplexityBranchHandler struct {
	config *MultiAgentConfig
}

// NewComplexityBranchHandler creates a new complexity branch handler
func NewComplexityBranchHandler(config *MultiAgentConfig) *ComplexityBranchHandler {
	return &ComplexityBranchHandler{
		config: config,
	}
}

// Evaluate determines the branch based on task complexity
func (h *ComplexityBranchHandler) Evaluate(ctx context.Context, state *MultiAgentState) (string, error) {
	if state.ConversationContext == nil {
		return directAnswerNodeKey, nil
	}

	switch state.ConversationContext.Complexity {
	case TaskComplexitySimple:
		return directAnswerNodeKey, nil
	case TaskComplexityModerate, TaskComplexityComplex, TaskComplexityVeryComplex:
		return planCreationNodeKey, nil
	default:
		return directAnswerNodeKey, nil
	}
}

// PlanCreationHandler handles task plan creation
type PlanCreationHandler struct {
	config *MultiAgentConfig
}

// NewPlanCreationHandler creates a new plan creation handler
func NewPlanCreationHandler(config *MultiAgentConfig) *PlanCreationHandler {
	return &PlanCreationHandler{
		config: config,
	}
}

// PreHandler prepares input for plan creation
func (h *PlanCreationHandler) PreHandler(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
	// Set planning state
	state.SetExecutionStatus(ExecutionStatusPlanning)

	prompt := buildPlanCreationPrompt(state, h.config)
	return []*schema.Message{prompt}, nil
}

// PostHandler processes plan creation results
func (h *PlanCreationHandler) PostHandler(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
	// Parse task plan from LLM response
	plan, err := h.parseTaskPlan(output.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task plan: %w", err)
	}

	// Update state using unified methods
	state.SetCurrentPlan(plan)
	state.SetExecutionStatus(ExecutionStatusExecuting)
	return output, nil
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
			AssignedSpecialist string         `json:"assignedSpecialist"`
			Priority           int            `json:"priority"`
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
	specialist *Specialist
}

// NewSpecialistHandler creates a new specialist handler
func NewSpecialistHandler(specialist *Specialist) *SpecialistHandler {
	return &SpecialistHandler{
		specialist: specialist,
	}
}

// PreHandler prepares input for specialist execution
func (h *SpecialistHandler) PreHandler(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
	// Find the current step for this specialist
	currentStep := h.findCurrentStep(state)
	if currentStep == nil {
		return nil, fmt.Errorf("no current step found for specialist %s", h.specialist.Name)
	}

	// Build specialist prompt
	prompt := buildSpecialistPrompt(h.specialist, currentStep, state)
	return prompt, nil
}

// PostHandler processes specialist execution results
func (h *SpecialistHandler) PostHandler(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
	// Create step result
	result := &StepResult{
		Success:      true,
		Output:       output,
		Confidence:   0.8, // TODO: implement confidence calculation
		QualityScore: 0.8, // TODO: implement quality scoring
	}

	// Update state using unified method
	state.UpdateSpecialistResult(h.specialist.Name, result)

	// Create execution record
	record := &ExecutionRecord{
		StepID:    state.CurrentStep,
		Action:    ActionTypeExecute,
		Output:    output,
		StartTime: time.Now(),
		Status:    ExecutionStatusCompleted,
	}
	state.AddExecutionRecord(record)

	// Update step status
	currentStep := h.findCurrentStep(state)
	if currentStep != nil {
		currentStep.Status = StepStatusCompleted
		currentStep.Result = result
	}

	return output, nil
}

func (h *SpecialistHandler) findCurrentStep(state *MultiAgentState) *PlanStep {
	if state.CurrentPlan == nil {
		return nil
	}

	for _, step := range state.CurrentPlan.Steps {
		if step.AssignedSpecialist == h.specialist.Name && step.Status == StepStatusRunning {
			return step
		}
	}

	return nil
}

// PlanExecutionHandler handles plan execution coordination
type PlanExecutionHandler struct {
	config *MultiAgentConfig
}

// NewPlanExecutionHandler creates a new plan execution handler
func NewPlanExecutionHandler(config *MultiAgentConfig) *PlanExecutionHandler {
	return &PlanExecutionHandler{
		config: config,
	}
}

// Execute coordinates the execution of the current plan
func (h *PlanExecutionHandler) Execute(ctx context.Context, input *schema.Message, state *MultiAgentState) (*schema.Message, error) {
	if state.CurrentPlan == nil {
		return nil, fmt.Errorf("no current plan to execute")
	}

	// Set execution state
	state.SetExecutionStatus(ExecutionStatusExecuting)
	state.ClearSpecialistResults() // Clear previous round results

	// Find the next step to execute
	nextStep := findNextStep(state)
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

	// Update current step using unified method
	state.SetCurrentStep(nextStep.ID)

	// Create execution record
	record := &ExecutionRecord{
		StepID:    nextStep.ID,
		Action:    ActionTypeExecute,
		Output:    input,
		StartTime: now,
		Status:    ExecutionStatusStarted,
	}
	state.AddExecutionRecord(record)

	return &schema.Message{
		Role:    schema.Assistant,
		Content: fmt.Sprintf("Executing step: %s - %s", nextStep.Name, nextStep.Description),
	}, nil
}

// findNextStep finds the next step to execute based on dependencies and status
func findNextStep(state *MultiAgentState) *PlanStep {
	for _, step := range state.CurrentPlan.Steps {
		if step.Status == StepStatusPending {
			// Check if all dependencies are completed
			if areDependenciesCompleted(step, state) {
				return step
			}
		}
	}
	return nil
}

// areDependenciesCompleted checks if all dependencies for a step are completed
func areDependenciesCompleted(step *PlanStep, state *MultiAgentState) bool {
	for _, depID := range step.Dependencies {
		for _, planStep := range state.CurrentPlan.Steps {
			if planStep.ID == depID && planStep.Status != StepStatusCompleted {
				return false
			}
		}
	}
	return true
}

func findCurrentStep(state *MultiAgentState) *PlanStep {
	if state.CurrentPlan == nil {
		return nil
	}

	for _, step := range state.CurrentPlan.Steps {
		if step.ID == state.CurrentStep {
			return step
		}
	}

	return nil
}

// SpecialistBranchHandler handles specialist selection and branching
type SpecialistBranchHandler struct {
	config *MultiAgentConfig
}

// NewSpecialistBranchHandler creates a new specialist branch handler
func NewSpecialistBranchHandler(config *MultiAgentConfig) *SpecialistBranchHandler {
	return &SpecialistBranchHandler{
		config: config,
	}
}

// Evaluate determines which specialist should handle the current step
func (h *SpecialistBranchHandler) Evaluate(ctx context.Context, state *MultiAgentState) (string, error) {
	if state.CurrentStep == "" {
		return generalSpecialistNodeKey, nil // No current step, go to common specialist
	}

	// Find the current step by ID
	currentStep := h.findStepByID(state.CurrentStep, state)
	if currentStep == nil {
		return generalSpecialistNodeKey, nil // Step not found, go to common specialist
	}

	// Return the assigned specialist for the current step
	assignedSpecialist := currentStep.AssignedSpecialist
	if assignedSpecialist == "" {
		return generalSpecialistNodeKey, nil // No specialist assigned, go to common specialist
	}

	// Verify the specialist exists in config
	for _, specialist := range h.config.Specialists {
		if specialist.Name == assignedSpecialist {
			return assignedSpecialist, nil
		}
	}

	// Specialist not found, go to common specialist
	return generalSpecialistNodeKey, nil
}

// findStepByID finds a step by its ID in the current plan
func (h *SpecialistBranchHandler) findStepByID(stepID string, state *MultiAgentState) *PlanStep {
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
func buildSpecialistBranchMap(specialists []*Specialist) map[string]bool {
	branchMap := make(map[string]bool)

	// Add all specialist names as valid branches
	for _, specialist := range specialists {
		branchMap[specialist.Name] = true
	}

	return branchMap
}

// ResultCollectorHandler handles result collection and summarization
type ResultCollectorHandler struct {
	config *MultiAgentConfig
}

// NewResultCollectorHandler creates a new result collector handler
func NewResultCollectorHandler(config *MultiAgentConfig) *ResultCollectorHandler {
	return &ResultCollectorHandler{
		config: config,
	}
}

// ResultCollectorLambda collects and summarizes specialist results
func (h *ResultCollectorHandler) ResultCollector(ctx context.Context, input []*schema.Message, state *MultiAgentState) (*schema.Message, error) {
	// Set collecting state
	state.SetExecutionStatus(ExecutionStatusCollecting)

	if len(state.SpecialistResults) == 0 {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: "No specialist results to collect.",
		}, nil
	}

	currentStep := findCurrentStep(state)
	if currentStep == nil {
		return nil, fmt.Errorf("current step not found")
	}

	// Collect all results
	var results []*schema.Message
	for specialistName, result := range state.SpecialistResults {
		if result.Success && result.Output != nil {
			// Add specialist name as context
			msg := &schema.Message{
				Role:    result.Output.Role,
				Content: fmt.Sprintf("%s\n[%s]: %s", currentStep.Description, specialistName, result.Output.Content),
			}
			results = append(results, msg)
			state.AddCollectedResult(msg)
		}
	}

	// Create summary
	summary := "Specialist Results Summary:\n\n"
	for _, msg := range results {
		summary += msg.Content + "\n\n"
	}

	finalResult := &schema.Message{
		Role:    schema.Assistant,
		Content: summary,
	}

	return finalResult, nil
}

type FeedbackProcessorHandler struct {
	config *MultiAgentConfig
}

func NewFeedbackProcessorHandler(config *MultiAgentConfig) *FeedbackProcessorHandler {
	return &FeedbackProcessorHandler{
		config: config,
	}
}

func (h *FeedbackProcessorHandler) PreHandler(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
	// Set feedback processing state
	state.SetExecutionStatus(ExecutionStatusRunning)
	return buildFeedbackPrompt(state), nil
}

func (h *FeedbackProcessorHandler) PostHandler(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
	err := h.processFeedbackResult(output, state)
	if err != nil {
		return output, err
	}
	// Update feedback history and reflection count
	state.IncrementReflection()
	return output, nil
}

func (h *FeedbackProcessorHandler) processFeedbackResult(output *schema.Message, state *MultiAgentState) error {
	// Parse feedback result
	var feedback struct {
		ExecutionCompleted bool     `json:"execution_completed"`
		OverallQuality     float64  `json:"overall_quality"`
		PlanNeedsUpdate    bool     `json:"plan_needs_update"`
		Issues             []string `json:"issues"`
		Suggestions        []string `json:"suggestions"`
		Confidence         float64  `json:"confidence"`
		NextActionReason   string   `json:"next_action_reason"`
	}

	err := json.Unmarshal([]byte(output.Content), &feedback)
	if err != nil {
		return fmt.Errorf("failed to parse feedback result: %w", err)
	}

	// Update state with feedback
	feedbackData := &Feedback{
		ExecutionCompleted: feedback.ExecutionCompleted,
		OverallQuality:     feedback.OverallQuality,
		PlanNeedsUpdate:    feedback.PlanNeedsUpdate,
		Issues:             feedback.Issues,
		Suggestions:        feedback.Suggestions,
		Confidence:         feedback.Confidence,
		NextActionReason:   feedback.NextActionReason,
	}
	state.AddFeedback(feedbackData)

	// Store feedback decision for branch evaluation
	state.SetMetadata("feedback_execution_completed", feedback.ExecutionCompleted)
	state.SetMetadata("feedback_plan_needs_update", feedback.PlanNeedsUpdate)
	state.SetMetadata("feedback_overall_quality", feedback.OverallQuality)
	state.SetMetadata("feedback_confidence", feedback.Confidence)
	state.SetMetadata("feedback_next_action_reason", feedback.NextActionReason)

	return nil
}

// ReflectionBranchHandler handles the decision logic for reflection branches
type ReflectionBranchHandler struct {
	config *MultiAgentConfig
}

func NewReflectionBranchHandler(config *MultiAgentConfig) *ReflectionBranchHandler {
	return &ReflectionBranchHandler{
		config: config,
	}
}

func (h *ReflectionBranchHandler) evaluateReflectionDecision(state *MultiAgentState) (decision string) {
	defer func() {
		if decision == planExecutionNodeKey && findNextStep(state) == nil {
			// 如果计划已经执行完毕，那么就不能再到planExecutionNode,而是应该到finalAnswerNode
			state.SetExecutionStatus(ExecutionStatusCompleted)
			decision = toFinalAnswerNodeKey
		}
	}()
	// Get feedback decision from metadata
	executionCompleted, hasCompleted := state.GetMetadata("feedback_execution_completed")
	planNeedsUpdate, hasUpdate := state.GetMetadata("feedback_plan_needs_update")
	overallQuality, hasQuality := state.GetMetadata("feedback_overall_quality")
	confidence, hasConfidence := state.GetMetadata("feedback_confidence")

	// If feedback metadata is missing, default to continue execution
	if !hasCompleted || !hasUpdate {
		decision = planExecutionNodeKey
		return
	}

	// Convert metadata to appropriate types
	isCompleted, ok := executionCompleted.(bool)
	if !ok {
		logger.Error("invalid execution_completed type")
		decision = planExecutionNodeKey
		return
	}

	needsUpdate, ok := planNeedsUpdate.(bool)
	if !ok {
		logger.Error("invalid plan_needs_update type")
		decision = planExecutionNodeKey
		return
	}

	// Decision logic based on feedback
	if isCompleted {
		// Task is completed, proceed to final answer
		state.SetExecutionStatus(ExecutionStatusCompleted)
		decision = toFinalAnswerNodeKey
		return
	}

	if needsUpdate {
		// Plan needs update, go to plan update
		state.SetExecutionStatus(ExecutionStatusPlanning)
		decision = toPlanUpdateNodeKey
		return
	}

	// Check quality and confidence thresholds
	if hasQuality && hasConfidence {
		quality, qOk := overallQuality.(float64)
		conf, cOk := confidence.(float64)
		if qOk && cOk && (quality < 0.6 || conf < 0.7) {
			// Low quality or confidence, consider plan update
			logger.Info("low quality or confidence, consider plan update")
			state.SetExecutionStatus(ExecutionStatusPlanning)
			decision = toPlanUpdateNodeKey
			return
		}
	}

	if !isCompleted {
		// If not completed, continue execution
		state.SetExecutionStatus(ExecutionStatusExecuting)
		decision = planExecutionNodeKey
		return
	}

	// Check if we've reached max rounds
	if state.RoundNumber >= state.MaxRounds {
		// Force completion if max rounds reached
		logger.Info("max rounds reached, force completion")
		state.SetExecutionStatus(ExecutionStatusCompleted)
		decision = toFinalAnswerNodeKey
		return
	}

	// Default: continue execution with current plan
	state.SetExecutionStatus(ExecutionStatusExecuting)
	decision = planExecutionNodeKey
	return
}

// PlanUpdateHandler handles the plan update process
type PlanUpdateHandler struct {
	config *MultiAgentConfig
}

func NewPlanUpdateHandler(config *MultiAgentConfig) *PlanUpdateHandler {
	return &PlanUpdateHandler{
		config: config,
	}
}

func (h *PlanUpdateHandler) PreHandler(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
	// Set plan update state
	state.SetExecutionStatus(ExecutionStatusPlanning)
	return buildPlanUpdatePrompt(state), nil
}

func (h *PlanUpdateHandler) PostHandler(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
	err := h.processPlanUpdate(output, state)
	if err != nil {
		return output, err
	}
	// After plan update, set status to execute the updated plan
	state.SetExecutionStatus(ExecutionStatusExecuting)
	return output, nil
}

func (h *PlanUpdateHandler) processPlanUpdate(output *schema.Message, state *MultiAgentState) error {
	// Parse incremental update operations
	var updateData struct {
		UpdateReason string `json:"update_reason"`
		Operations   []struct {
			Type     string    `json:"type"`
			StepID   string    `json:"stepID"`
			StepData *StepData `json:"step_data,omitempty"`
			Position string    `json:"position,omitempty"`
			Reason   string    `json:"reason,omitempty"`
		} `json:"operations"`
		PlanMetadata *struct {
			Name        string `json:"name,omitempty"`
			Description string `json:"description,omitempty"`
		} `json:"plan_metadata,omitempty"`
	}

	err := json.Unmarshal([]byte(output.Content), &updateData)
	if err != nil {
		return fmt.Errorf("failed to parse plan update operations: %w", err)
	}

	if state.CurrentPlan == nil {
		return fmt.Errorf("no current plan to update")
	}

	// Create a copy of the current plan for incremental updates
	updatedPlan := h.clonePlan(state.CurrentPlan)
	updatedPlan.Version++
	updatedPlan.UpdatedAt = time.Now()

	// Update plan metadata if provided
	if updateData.PlanMetadata != nil {
		if updateData.PlanMetadata.Name != "" {
			updatedPlan.Name = updateData.PlanMetadata.Name
		}
		if updateData.PlanMetadata.Description != "" {
			updatedPlan.Description = updateData.PlanMetadata.Description
		}
	}

	// Apply operations in order
	var appliedOperations []string
	var operationDataList []OperationData
	for _, op := range updateData.Operations {
		opData := &OperationData{
			Type:     op.Type,
			StepID:   op.StepID,
			StepData: op.StepData,
			Position: op.Position,
			Reason:   op.Reason,
		}
		err := h.applyOperation(updatedPlan, opData)
		if err != nil {
			return fmt.Errorf("failed to apply operation %s: %w", op.Type, err)
		}
		appliedOperations = append(appliedOperations, fmt.Sprintf("%s:%s", op.Type, op.StepID))
		operationDataList = append(operationDataList, *opData)
	}

	// Add old plan to history
	state.AddPlanToHistory(state.CurrentPlan)

	// Update state with modified plan
	state.SetCurrentPlan(updatedPlan)

	// Record the plan update
	planUpdate := &PlanUpdate{
		ID:          fmt.Sprintf("update_%d", time.Now().Unix()),
		PlanVersion: updatedPlan.Version,
		UpdateType:  h.determineUpdateType(operationDataList),
		Description: updateData.UpdateReason,
		Reason:      "Plan updated incrementally based on execution feedback",
		Timestamp:   time.Now(),
		Metadata: map[string]any{
			"round":              state.RoundNumber,
			"applied_operations": appliedOperations,
			"operation_count":    len(updateData.Operations),
		},
	}

	// Add update to plan history
	if updatedPlan.UpdateHistory == nil {
		updatedPlan.UpdateHistory = make([]*PlanUpdate, 0)
	}
	updatedPlan.UpdateHistory = append(updatedPlan.UpdateHistory, planUpdate)

	// Only clear specialist results for steps that were modified/removed
	// This preserves results for completed and unmodified steps
	h.selectiveClearSpecialistResults(state, operationDataList)

	state.RoundNumber++
	return nil
}

// clonePlan creates a deep copy of the task plan
func (h *PlanUpdateHandler) clonePlan(plan *TaskPlan) *TaskPlan {
	cloned := &TaskPlan{
		ID:            plan.ID,
		Version:       plan.Version,
		Name:          plan.Name,
		Description:   plan.Description,
		Status:        plan.Status,
		CreatedAt:     plan.CreatedAt,
		UpdatedAt:     plan.UpdatedAt,
		Steps:         make([]*PlanStep, len(plan.Steps)),
		UpdateHistory: make([]*PlanUpdate, len(plan.UpdateHistory)),
		Metadata:      make(map[string]any),
	}

	// Clone steps
	for i, step := range plan.Steps {
		cloned.Steps[i] = &PlanStep{
			ID:                 step.ID,
			Name:               step.Name,
			Description:        step.Description,
			AssignedSpecialist: step.AssignedSpecialist,
			Priority:           step.Priority,
			Status:             step.Status,
			Dependencies:       make([]string, len(step.Dependencies)),
			Parameters:         make(map[string]any),
			Result:             step.Result, // Shallow copy is fine for Result
			Metadata:           make(map[string]any),
		}
		copy(cloned.Steps[i].Dependencies, step.Dependencies)
		for k, v := range step.Parameters {
			cloned.Steps[i].Parameters[k] = v
		}
		for k, v := range step.Metadata {
			cloned.Steps[i].Metadata[k] = v
		}
	}

	// Clone update history
	copy(cloned.UpdateHistory, plan.UpdateHistory)

	// Clone metadata
	for k, v := range plan.Metadata {
		cloned.Metadata[k] = v
	}

	return cloned
}

// OperationData defines the structure for plan update operations
type OperationData struct {
	Type     string    `json:"type"`
	StepID   string    `json:"stepID"`
	StepData *StepData `json:"step_data,omitempty"`
	Position string    `json:"position,omitempty"`
	Reason   string    `json:"reason,omitempty"`
}

type StepData struct {
	ID                 string         `json:"id"`
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	AssignedSpecialist string         `json:"assignedSpecialist"`
	Priority           int            `json:"priority"`
	Dependencies       []string       `json:"dependencies,omitempty"`
	Parameters         map[string]any `json:"parameters,omitempty"`
}

// applyOperation applies a single update operation to the plan
func (h *PlanUpdateHandler) applyOperation(plan *TaskPlan, op *OperationData) error {
	switch op.Type {
	case "add":
		return h.addStep(plan, op)
	case "modify":
		return h.modifyStep(plan, op)
	case "remove":
		return h.removeStep(plan, op)
	case "reorder":
		return h.reorderStep(plan, op)
	default:
		return fmt.Errorf("unknown operation type: %s", op.Type)
	}
}

// addStep adds a new step to the plan
func (h *PlanUpdateHandler) addStep(plan *TaskPlan, op *OperationData) error {
	if op.StepData == nil {
		return fmt.Errorf("step_data is required for add operation")
	}

	newStep := &PlanStep{
		ID:                 op.StepData.ID,
		Name:               op.StepData.Name,
		Description:        op.StepData.Description,
		AssignedSpecialist: op.StepData.AssignedSpecialist,
		Priority:           op.StepData.Priority,
		Status:             StepStatusPending,
		Dependencies:       op.StepData.Dependencies,
		Parameters:         op.StepData.Parameters,
		Metadata:           map[string]any{"created_at": time.Now(), "operation": "add"},
	}

	// Determine insertion position
	insertIndex := len(plan.Steps) // Default: append at end

	if op.Position != "" {
		switch op.Position {
		case "before":
			if idx := h.findStepIndex(plan, op.StepID); idx >= 0 {
				insertIndex = idx
			}
		case "after":
			if idx := h.findStepIndex(plan, op.StepID); idx >= 0 {
				insertIndex = idx + 1
			}
		default:
			// Try to parse as index
			if idx, err := fmt.Sscanf(op.Position, "%d", &insertIndex); err == nil && idx == 1 {
				if insertIndex < 0 || insertIndex > len(plan.Steps) {
					insertIndex = len(plan.Steps)
				}
			}
		}
	}

	// Insert step at the determined position
	plan.Steps = append(plan.Steps[:insertIndex], append([]*PlanStep{newStep}, plan.Steps[insertIndex:]...)...)
	return nil
}

// modifyStep modifies an existing step
func (h *PlanUpdateHandler) modifyStep(plan *TaskPlan, op *OperationData) error {
	stepIndex := h.findStepIndex(plan, op.StepID)
	if stepIndex < 0 {
		return fmt.Errorf("step %s not found", op.StepID)
	}

	step := plan.Steps[stepIndex]

	// Don't modify completed steps unless explicitly allowed
	if step.Status == StepStatusCompleted {
		return fmt.Errorf("cannot modify completed step %s", op.StepID)
	}

	if op.StepData != nil {
		if op.StepData.Name != "" {
			step.Name = op.StepData.Name
		}
		if op.StepData.Description != "" {
			step.Description = op.StepData.Description
		}
		if op.StepData.AssignedSpecialist != "" {
			step.AssignedSpecialist = op.StepData.AssignedSpecialist
		}
		if op.StepData.Priority > 0 {
			step.Priority = op.StepData.Priority
		}
		if op.StepData.Dependencies != nil {
			step.Dependencies = op.StepData.Dependencies
		}
		if op.StepData.Parameters != nil {
			step.Parameters = op.StepData.Parameters
		}
	}

	// Update metadata
	if step.Metadata == nil {
		step.Metadata = make(map[string]any)
	}
	step.Metadata["last_modified"] = time.Now()
	step.Metadata["operation"] = "modify"

	return nil
}

// removeStep removes a step from the plan
func (h *PlanUpdateHandler) removeStep(plan *TaskPlan, op *OperationData) error {
	stepIndex := h.findStepIndex(plan, op.StepID)
	if stepIndex < 0 {
		return fmt.Errorf("step %s not found", op.StepID)
	}

	step := plan.Steps[stepIndex]

	// Don't remove completed steps
	if step.Status == StepStatusCompleted {
		return fmt.Errorf("cannot remove completed step %s", op.StepID)
	}

	// Check if other steps depend on this step
	for _, otherStep := range plan.Steps {
		for _, dep := range otherStep.Dependencies {
			if dep == op.StepID {
				return fmt.Errorf("cannot remove step %s: step %s depends on it", op.StepID, otherStep.ID)
			}
		}
	}

	// Remove the step
	plan.Steps = append(plan.Steps[:stepIndex], plan.Steps[stepIndex+1:]...)
	return nil
}

// reorderStep changes the position of a step
func (h *PlanUpdateHandler) reorderStep(plan *TaskPlan, op *OperationData) error {
	stepIndex := h.findStepIndex(plan, op.StepID)
	if stepIndex < 0 {
		return fmt.Errorf("step %s not found", op.StepID)
	}

	step := plan.Steps[stepIndex]

	// Don't reorder completed steps
	if step.Status == StepStatusCompleted {
		return fmt.Errorf("cannot reorder completed step %s", op.StepID)
	}

	// Remove step from current position
	plan.Steps = append(plan.Steps[:stepIndex], plan.Steps[stepIndex+1:]...)

	// Determine new position
	newIndex := len(plan.Steps) // Default: append at end

	if op.Position != "" {
		switch op.Position {
		case "before":
			// Position is relative to another step, but we need the target step ID
			// This would need additional logic to specify the target step
		case "after":
			// Similar to "before"
		default:
			// Try to parse as index
			if idx, err := fmt.Sscanf(op.Position, "%d", &newIndex); err == nil && idx == 1 {
				if newIndex < 0 || newIndex > len(plan.Steps) {
					newIndex = len(plan.Steps)
				}
			}
		}
	}

	// Insert step at new position
	plan.Steps = append(plan.Steps[:newIndex], append([]*PlanStep{step}, plan.Steps[newIndex:]...)...)
	return nil
}

// findStepIndex finds the index of a step by ID
func (h *PlanUpdateHandler) findStepIndex(plan *TaskPlan, stepID string) int {
	for i, step := range plan.Steps {
		if step.ID == stepID {
			return i
		}
	}
	return -1
}

// determineUpdateType determines the primary update type based on operations
func (h *PlanUpdateHandler) determineUpdateType(operations []OperationData) PlanUpdateType {
	if len(operations) == 0 {
		return PlanUpdateTypeUnknown
	}

	// Count operation types
	typeCounts := make(map[string]int)
	for _, op := range operations {
		typeCounts[op.Type]++
	}

	// Determine primary type based on most frequent operation
	if typeCounts["add"] > 0 {
		return PlanUpdateTypeStepAdd
	}
	if typeCounts["modify"] > 0 {
		return PlanUpdateTypeStepModify
	}
	if typeCounts["remove"] > 0 {
		return PlanUpdateTypeStepRemove
	}
	if typeCounts["reorder"] > 0 {
		return PlanUpdateTypeStepReorder
	}

	return PlanUpdateTypeStrategyChange
}

// selectiveClearSpecialistResults clears results only for affected steps
func (h *PlanUpdateHandler) selectiveClearSpecialistResults(state *MultiAgentState, operations []OperationData) {
	// Collect step IDs that were modified or removed
	affectedSteps := make(map[string]bool)
	for _, op := range operations {
		switch op.Type {
		case "modify", "remove":
			affectedSteps[op.StepID] = true
		case "add":
			// New steps don't have existing results to clear
		case "reorder":
			// Reordering doesn't affect results, but might affect dependencies
			// For now, we'll be conservative and clear reordered steps
			affectedSteps[op.StepID] = true
		}
	}

	// Clear specialist results only for affected steps
	if len(affectedSteps) > 0 {
		newResults := make(map[string]*StepResult)
		for stepID, result := range state.SpecialistResults {
			if !affectedSteps[stepID] {
				newResults[stepID] = result
			}
		}
		state.SpecialistResults = newResults
	}
}

type FinalAnswerHandler struct {
	config *MultiAgentConfig
}

func NewFinalAnswerHandler(config *MultiAgentConfig) *FinalAnswerHandler {
	return &FinalAnswerHandler{
		config: config,
	}
}

func (h *FinalAnswerHandler) PreHandler(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
	// Build final answer prompt
	prompt := buildFinalAnswerPrompt(state)
	return []*schema.Message{prompt}, nil
}

func (h *FinalAnswerHandler) PostHandler(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
	state.FinalAnswer = output
	state.IsCompleted = true
	return output, nil
}
