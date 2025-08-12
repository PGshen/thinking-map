package multiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	// Node keys
	conversationAnalyzerNodeKey = "conversation_analyzer"
	toComplexityBranchNodeKey   = "to_complexity_branch"
	complexityBranchNodeKey     = "complexity_branch"
	directAnswerNodeKey         = "direct_answer"
	planCreationNodeKey         = "plan_creation"
	planExecutionNodeKey        = "plan_execution"
	toSpecialistBranchNodeKey   = "to_specialist_branch"
	specialistBranchNodeKey     = "specialist_branch"
	resultCollectorNodeKey      = "result_collector"
	toFeedbackProcessorNodeKey  = "to_feedback_processor"
	feedbackProcessorNodeKey    = "feedback_processor"
	reflectionBranchNodeKey     = "reflection_branch"
	toPlanUpdateNodeKey         = "to_plan_update"
	planUpdateNodeKey           = "plan_update"
	toFinalAnswerNodeKey        = "to_final_answer"
	finalAnswerNodeKey          = "final_answer"
)

// NewMultiAgent creates a new multi-agent system
func NewMultiAgent(ctx context.Context, config *MultiAgentConfig) (*MultiAgent, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create the graph with state
	graph := compose.NewGraph[[]*schema.Message, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *MultiAgentState {
			return &MultiAgentState{
				RoundNumber:     1,
				StartTime:       time.Now(),
				ExecutionStatus: ExecutionStatusStarted,
				MaxRounds:       config.MaxRounds,
				ShouldContinue:  true,
				IsCompleted:     false,
			}
		}),
	)

	// Add conversation analyzer node
	conversationAnalyzer := NewConversationAnalyzerHandler(config)
	err := graph.AddChatModelNode(conversationAnalyzerNodeKey, config.Host.Model,
		compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
			return conversationAnalyzer.PreHandler(ctx, input, state)
		}),
		compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
			return conversationAnalyzer.PostHandler(ctx, output, state)
		}),
		compose.WithNodeName("conversation_analyzer"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add conversation analyzer node: %w", err)
	}

	toComplexityBranch := compose.ToList[*schema.Message]()
	graph.AddLambdaNode(toComplexityBranchNodeKey, toComplexityBranch, compose.WithNodeName("to_complexity_branch"))

	// Add complexity branch node
	complexityBranchHandler := NewComplexityBranchHandler(config)
	complexityBranch := compose.NewGraphBranch(func(ctx context.Context, input []*schema.Message) (string, error) {
		var result string
		err = compose.ProcessState(ctx, func(ctx context.Context, state *MultiAgentState) error {
			result, err = complexityBranchHandler.Evaluate(ctx, state)
			return err
		})
		return result, err
	}, map[string]bool{directAnswerNodeKey: true, planCreationNodeKey: true})

	// Add direct answer node for simple tasks
	err = graph.AddChatModelNode(directAnswerNodeKey, config.Host.Model,
		compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
			// Build direct answer prompt
			prompt := buildDirectAnswerPrompt(state)
			return []*schema.Message{prompt}, nil
		}),
		compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
			state.FinalAnswer = output
			state.IsCompleted = true
			return output, nil
		}),
		compose.WithNodeName("direct_answer"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add direct answer node: %w", err)
	}

	// Add plan creation node
	planCreationHandler := NewPlanCreationHandler(config)
	err = graph.AddChatModelNode(planCreationNodeKey, config.Host.Model,
		compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
			return planCreationHandler.PreHandler(ctx, input, state)
		}),
		compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
			return planCreationHandler.PostHandler(ctx, output, state)
		}),
		compose.WithNodeName("plan_creation"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add plan creation node: %w", err)
	}

	// Add plan execution node
	planExecutionHandler := NewPlanExecutionHandler(config)
	err = graph.AddLambdaNode(planExecutionNodeKey,
		compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
			var result *schema.Message
			err = compose.ProcessState(ctx, func(ctx context.Context, state *MultiAgentState) error {
				result, err = planExecutionHandler.Execute(ctx, input, state)
				return err
			})
			return result, err
		}),
		compose.WithNodeName("plan_execution"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add plan execution node: %w", err)
	}

	toSpecialistBranch := compose.ToList[*schema.Message]()
	graph.AddLambdaNode(toSpecialistBranchNodeKey, toSpecialistBranch, compose.WithNodeName("to_specialist_branch"))

	// Add specialist branch
	specialistBranchHandler := NewSpecialistBranchHandler(config)
	specialistBranch := compose.NewGraphBranch(func(ctx context.Context, input []*schema.Message) (string, error) {
		var result string
		err = compose.ProcessState(ctx, func(ctx context.Context, state *MultiAgentState) error {
			result, err = specialistBranchHandler.Evaluate(ctx, state)
			return err
		})
		return result, err
	}, buildSpecialistBranchMap(config.Specialists))

	// Add specialist nodes
	for _, specialist := range config.Specialists {
		if err = addSpecialist(graph, specialist); err != nil {
			return nil, fmt.Errorf("failed to add specialist node %s: %w", specialist.Name, err)
		}
	}

	// Add result collector node
	err = graph.AddLambdaNode(resultCollectorNodeKey,
		compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
			var result *schema.Message
			err = compose.ProcessState(ctx, func(ctx context.Context, state *MultiAgentState) error {
				// Convert single message to slice for ResultCollectorLambda
				messages := []*schema.Message{input}
				result, err = ResultCollectorLambda(ctx, messages, state)
				return err
			})
			return result, err
		}),
		compose.WithNodeName("result_collector"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add result collector node: %w", err)
	}

	toFeedbackProcessor := compose.ToList[*schema.Message]()
	graph.AddLambdaNode(toFeedbackProcessorNodeKey, toFeedbackProcessor, compose.WithNodeName("to_feedback_processor"))

	// Add feedback processor node
	err = graph.AddChatModelNode(feedbackProcessorNodeKey, config.Host.Model,
		compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
			// Set feedback processing state
			state.SetExecutionStatus(ExecutionStatusRunning)
			return buildFeedbackPrompt(state), nil
		}),
		compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
			err = processFeedbackResult(output, state)
			if err != nil {
				return output, err
			}
			// Update feedback history and reflection count
			state.IncrementReflection()
			return output, nil
		}),
		compose.WithNodeName("feedback_processor"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add feedback processor node: %w", err)
	}

	// Add reflection branch node
	reflectionBranch := compose.NewGraphBranch(func(ctx context.Context, input *schema.Message) (string, error) {
		var result string
		err = compose.ProcessState(ctx, func(ctx context.Context, state *MultiAgentState) error {
			result, err = evaluateReflectionDecision(state)
			return err
		})
		return result, err
	}, map[string]bool{
		planExecutionNodeKey: true,
		toPlanUpdateNodeKey:  true,
		toFinalAnswerNodeKey: true,
	})

	toPlanUpdate := compose.ToList[*schema.Message]()
	graph.AddLambdaNode(toPlanUpdateNodeKey, toPlanUpdate, compose.WithNodeName("to_plan_update"))

	// Add plan update node
	err = graph.AddChatModelNode(planUpdateNodeKey, config.Host.Model,
		compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
			// Set plan update state
			state.SetExecutionStatus(ExecutionStatusPlanning)
			return buildPlanUpdatePrompt(state), nil
		}),
		compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
			err = processPlanUpdate(output, state)
			if err != nil {
				return output, err
			}
			// After plan update, set status to execute the updated plan
			state.SetExecutionStatus(ExecutionStatusExecuting)
			return output, nil
		}),
		compose.WithNodeName("plan_update"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add plan update node: %w", err)
	}

	toFinalAnswer := compose.ToList[*schema.Message]()
	graph.AddLambdaNode(toFinalAnswerNodeKey, toFinalAnswer, compose.WithNodeName("to_final_answer"))

	// Add final answer node
	err = graph.AddChatModelNode(finalAnswerNodeKey, config.Host.Model,
		compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
			// Build final answer prompt
			prompt := buildFinalAnswerPrompt(state)
			return []*schema.Message{prompt}, nil
		}),
		compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
			state.FinalAnswer = output
			state.IsCompleted = true
			return output, nil
		}),
		compose.WithNodeName("final_answer"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add final answer node: %w", err)
	}

	// Define edges
	graph.AddEdge(compose.START, conversationAnalyzerNodeKey)
	graph.AddEdge(conversationAnalyzerNodeKey, toComplexityBranchNodeKey)

	// Complexity branch - directly from conversation analyzer
	graph.AddBranch(toComplexityBranchNodeKey, complexityBranch)

	// Direct answer path
	graph.AddEdge(directAnswerNodeKey, compose.END)

	// Plan and execute path
	graph.AddEdge(planCreationNodeKey, planExecutionNodeKey)

	// Plan execution branch
	graph.AddEdge(planExecutionNodeKey, toSpecialistBranchNodeKey)
	graph.AddBranch(toSpecialistBranchNodeKey, specialistBranch)

	// All specialists go to result collector
	for _, specialist := range config.Specialists {
		graph.AddEdge(specialist.Name, resultCollectorNodeKey)
	}

	// Result collector goes to feedback processor
	graph.AddEdge(resultCollectorNodeKey, toFeedbackProcessorNodeKey)
	graph.AddEdge(toFeedbackProcessorNodeKey, feedbackProcessorNodeKey)

	// Three-way reflection branch from feedback processor
	graph.AddBranch(feedbackProcessorNodeKey, reflectionBranch)

	// Plan update loop
	graph.AddEdge(toPlanUpdateNodeKey, planUpdateNodeKey)
	graph.AddEdge(planUpdateNodeKey, planExecutionNodeKey)

	// Final answer
	graph.AddEdge(toFinalAnswerNodeKey, finalAnswerNodeKey)
	graph.AddEdge(finalAnswerNodeKey, compose.END)

	// Compile the graph
	runnable, err := graph.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile graph: %w", err)
	}

	return &MultiAgent{
		runnable:         runnable,
		graph:            graph,
		graphAddNodeOpts: []compose.GraphAddNodeOpt{},
		config:           config,
	}, nil
}

func addSpecialist(graph *compose.Graph[[]*schema.Message, *schema.Message], specialist *Specialist) error {
	specialistHandler := NewSpecialistHandler(specialist)
	if specialist.Invokable != nil || specialist.Streamable != nil {
		lambda, err := compose.AnyLambda(specialist.Invokable, specialist.Streamable, nil, nil, compose.WithLambdaType("Specialist"))
		if err != nil {
			return err
		}
		if err := graph.AddLambdaNode(specialist.Name, lambda, compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
			return specialistHandler.PreHandler(ctx, input, state)
		}),
			compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
				return specialistHandler.PostHandler(ctx, output, state)
			}),
			compose.WithNodeName(specialist.Name)); err != nil {
			return err
		}
	} else if specialist.ChatModel != nil {
		if err := graph.AddChatModelNode(specialist.Name, specialist.ChatModel,
			compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
				return specialistHandler.PreHandler(ctx, input, state)
			}),
			compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
				return specialistHandler.PostHandler(ctx, output, state)
			}),
			compose.WithNodeName(specialist.Name),
		); err != nil {
			return err
		}
	}
	return nil
}

func processFeedbackResult(output *schema.Message, state *MultiAgentState) error {
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
	feedbackData := map[string]any{
		"content":             output.Content,
		"timestamp":           time.Now(),
		"execution_completed": feedback.ExecutionCompleted,
		"plan_needs_update":   feedback.PlanNeedsUpdate,
		"overall_quality":     feedback.OverallQuality,
		"confidence":          feedback.Confidence,
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

func evaluateReflectionDecision(state *MultiAgentState) (string, error) {
	// Get feedback decision from metadata
	executionCompleted, hasCompleted := state.GetMetadata("feedback_execution_completed")
	planNeedsUpdate, hasUpdate := state.GetMetadata("feedback_plan_needs_update")
	overallQuality, hasQuality := state.GetMetadata("feedback_overall_quality")
	confidence, hasConfidence := state.GetMetadata("feedback_confidence")

	// If feedback metadata is missing, default to continue execution
	if !hasCompleted || !hasUpdate {
		return planExecutionNodeKey, nil
	}

	// Convert metadata to appropriate types
	isCompleted, ok := executionCompleted.(bool)
	if !ok {
		return planExecutionNodeKey, fmt.Errorf("invalid execution_completed type")
	}

	needsUpdate, ok := planNeedsUpdate.(bool)
	if !ok {
		return planExecutionNodeKey, fmt.Errorf("invalid plan_needs_update type")
	}

	// Decision logic based on feedback
	if isCompleted {
		// Task is completed, proceed to final answer
		state.SetExecutionStatus(ExecutionStatusCompleted)
		return toFinalAnswerNodeKey, nil
	}

	if needsUpdate {
		// Plan needs update, go to plan update
		state.SetExecutionStatus(ExecutionStatusPlanning)
		return toPlanUpdateNodeKey, nil
	}

	// Check quality and confidence thresholds
	if hasQuality && hasConfidence {
		quality, qOk := overallQuality.(float64)
		conf, cOk := confidence.(float64)
		if qOk && cOk && (quality < 0.6 || conf < 0.7) {
			// Low quality or confidence, consider plan update
			state.SetExecutionStatus(ExecutionStatusPlanning)
			return toPlanUpdateNodeKey, nil
		}
	}

	if !isCompleted {
		// If not completed, continue execution
		state.SetExecutionStatus(ExecutionStatusExecuting)
		return planExecutionNodeKey, nil
	}

	// Check if we've reached max rounds
	if state.RoundNumber >= state.MaxRounds {
		// Force completion if max rounds reached
		state.SetExecutionStatus(ExecutionStatusCompleted)
		return toFinalAnswerNodeKey, nil
	}

	// Default: continue execution with current plan
	state.SetExecutionStatus(ExecutionStatusExecuting)
	return toFinalAnswerNodeKey, nil
}

func processPlanUpdate(output *schema.Message, state *MultiAgentState) error {
	// Parse updated plan
	var planData struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		UpdateReason string `json:"update_reason"`
		Steps        []struct {
			ID                 string         `json:"id"`
			Name               string         `json:"name"`
			Description        string         `json:"description"`
			AssignedSpecialist string         `json:"assigned_specialist"`
			Priority           int            `json:"priority"`
			Dependencies       []string       `json:"dependencies,omitempty"`
			Parameters         map[string]any `json:"parameters,omitempty"`
		} `json:"steps"`
	}

	err := json.Unmarshal([]byte(output.Content), &planData)
	if err != nil {
		return fmt.Errorf("failed to parse updated plan: %w", err)
	}

	// Create new plan with updated information
	updatedPlan := &TaskPlan{
		ID:          fmt.Sprintf("plan_%d", time.Now().Unix()),
		Version:     1,
		Name:        planData.Name,
		Description: planData.Description,
		Status:      ExecutionStatusPlanning,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Steps:       make([]*PlanStep, len(planData.Steps)),
		Metadata:    map[string]any{"update_reason": planData.UpdateReason},
	}

	// If there's a current plan, increment version and add to history
	if state.CurrentPlan != nil {
		updatedPlan.Version = state.CurrentPlan.Version + 1
		state.AddPlanToHistory(state.CurrentPlan)
	}

	// Convert steps
	for i, stepData := range planData.Steps {
		updatedPlan.Steps[i] = &PlanStep{
			ID:                 stepData.ID,
			Name:               stepData.Name,
			Description:        stepData.Description,
			AssignedSpecialist: stepData.AssignedSpecialist,
			Priority:           stepData.Priority,
			Status:             StepStatusPending,
			Dependencies:       stepData.Dependencies,
			Parameters:         stepData.Parameters,
			Metadata:           map[string]any{"created_at": time.Now()},
		}
	}

	// Update state with new plan
	state.SetCurrentPlan(updatedPlan)

	// Record the plan update
	planUpdate := &PlanUpdate{
		ID:          fmt.Sprintf("update_%d", time.Now().Unix()),
		PlanVersion: updatedPlan.Version,
		UpdateType:  PlanUpdateTypeStrategyChange,
		Description: planData.UpdateReason,
		Reason:      "Plan updated based on execution feedback",
		Timestamp:   time.Now(),
		Metadata:    map[string]any{"round": state.RoundNumber},
	}

	// Add update to plan history
	if updatedPlan.UpdateHistory == nil {
		updatedPlan.UpdateHistory = make([]*PlanUpdate, 0)
	}
	updatedPlan.UpdateHistory = append(updatedPlan.UpdateHistory, planUpdate)

	// Clear previous specialist results since plan has changed
	state.ClearSpecialistResults()

	state.RoundNumber++
	return nil
}
