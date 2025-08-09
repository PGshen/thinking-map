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
	"fmt"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	// Node keys
	conversationAnalyzerNodeKey = "conversation_analyzer"
	complexityBranchNodeKey     = "complexity_branch"
	directAnswerNodeKey         = "direct_answer"
	planCreationNodeKey         = "plan_creation"
	planExecutionNodeKey        = "plan_execution"
	specialistBranchNodeKey     = "specialist_branch"
	resultCollectorNodeKey      = "result_collector"
	feedbackProcessorNodeKey    = "feedback_processor"
	reflectionBranchNodeKey     = "reflection_branch"
	planUpdateNodeKey           = "plan_update"
	finalAnswerNodeKey          = "final_answer"

	// Branch conditions
	directAnswerBranch      = "direct_answer"
	planAndExecuteBranch    = "plan_and_execute"
	continueExecutionBranch = "continue"
	finishExecutionBranch   = "finish"
)

// NewEnhancedMultiAgent creates a new enhanced multi-agent system
func NewEnhancedMultiAgent(ctx context.Context, config *EnhancedMultiAgentConfig) (*EnhancedMultiAgent, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create the graph with state
	graph := compose.NewGraph[[]*schema.Message, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *EnhancedState {
			return &EnhancedState{
				RoundNumber:     1,
				StartTime:       time.Now(),
				ExecutionStatus: ExecutionStatusStarted,
				MaxRounds:       config.ExecutionControl.MaxRounds,
				ShouldContinue:  true,
				IsCompleted:     false,
			}
		}),
	)

	// Add conversation analyzer node
	conversationAnalyzer := NewConversationAnalyzerHandler(config)
	err := graph.AddChatModelNode(conversationAnalyzerNodeKey, config.Host.Model,
		compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *EnhancedState) ([]*schema.Message, error) {
			return conversationAnalyzer.PreHandler(ctx, input, state)
		}),
		compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
			return conversationAnalyzer.PostHandler(ctx, output, state)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add conversation analyzer node: %w", err)
	}



	// Add complexity branch node
	complexityBranchHandler := NewComplexityBranchHandler(config)
	complexityBranch := compose.NewGraphBranch(func(ctx context.Context, input *schema.Message) (string, error) {
		var result string
		err = compose.ProcessState(ctx, func(ctx context.Context, state *EnhancedState) error {
			result, err = complexityBranchHandler.Evaluate(ctx, state)
			return err
		})
		return result, err
	}, map[string]bool{directAnswerBranch: true, planAndExecuteBranch: true})

	// Add direct answer node for simple tasks
	err = graph.AddChatModelNode(directAnswerNodeKey, config.Host.Model,
		compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *EnhancedState) ([]*schema.Message, error) {
			// Build direct answer prompt
			prompt := buildDirectAnswerPrompt(state)
			return []*schema.Message{prompt}, nil
		}),
		compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
			state.FinalAnswer = output
			state.IsCompleted = true
			return output, nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add direct answer node: %w", err)
	}

	// Add plan creation node
	planCreationHandler := NewPlanCreationHandler(config)
	err = graph.AddChatModelNode(planCreationNodeKey, config.Host.Model,
		compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *EnhancedState) ([]*schema.Message, error) {
			return planCreationHandler.PreHandler(ctx, input, state)
		}),
		compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
			return planCreationHandler.PostHandler(ctx, output, state)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add plan creation node: %w", err)
	}

	// Add plan execution node
	planExecutionHandler := NewPlanExecutionHandler(config)
	err = graph.AddLambdaNode(planExecutionNodeKey,
		compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
			var result *schema.Message
			err = compose.ProcessState(ctx, func(ctx context.Context, state *EnhancedState) error {
				result, err = planExecutionHandler.Execute(ctx, input, state)
				return err
			})
			return result, err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add plan execution node: %w", err)
	}

	// Add specialist branch
	specialistBranchHandler := NewSpecialistBranchHandler(config)
	specialistBranch := compose.NewGraphBranch(func(ctx context.Context, input *schema.Message) (string, error) {
		var result string
		err = compose.ProcessState(ctx, func(ctx context.Context, state *EnhancedState) error {
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
			err = compose.ProcessState(ctx, func(ctx context.Context, state *EnhancedState) error {
				// Convert single message to slice for ResultCollectorLambda
				messages := []*schema.Message{input}
				result, err = ResultCollectorLambda(ctx, messages, state)
				return err
			})
			return result, err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add result collector node: %w", err)
	}

	// Add feedback processor node
	err = graph.AddChatModelNode(feedbackProcessorNodeKey, config.Host.Model,
		compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *EnhancedState) ([]*schema.Message, error) {
			return buildFeedbackPrompt(state), nil
		}),
		compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
			err = processFeedbackResult(output, state)
			return output, err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add feedback processor node: %w", err)
	}

	// Add reflection branch node
	reflectionBranch := compose.NewGraphBranch(func(ctx context.Context, input *schema.Message) (string, error) {
		var result string
		err = compose.ProcessState(ctx, func(ctx context.Context, state *EnhancedState) error {
			result, err = evaluateReflectionDecision(state)
			return err
		})
		return result, err
	}, map[string]bool{continueExecutionBranch: true, finishExecutionBranch: true})

	err = graph.AddBranch(feedbackProcessorNodeKey, reflectionBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to add reflection branch: %w", err)
	}

	// Add plan update node
	err = graph.AddChatModelNode(planUpdateNodeKey, config.Host.Model,
		compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *EnhancedState) ([]*schema.Message, error) {
			return buildPlanUpdatePrompt(state), nil
		}),
		compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
			err = processPlanUpdate(output, state)
			return output, err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add plan update node: %w", err)
	}

	// Add final answer node
	err = graph.AddLambdaNode(finalAnswerNodeKey,
		compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
			var result *schema.Message
			err = compose.ProcessState(ctx, func(ctx context.Context, state *EnhancedState) error {
				result = generateFinalAnswer(state)
				return nil
			})
			return result, err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add final answer node: %w", err)
	}

	// Define edges
	graph.AddEdge(compose.START, conversationAnalyzerNodeKey)

	// Complexity branch - directly from conversation analyzer
	graph.AddBranch(conversationAnalyzerNodeKey, complexityBranch)

	// Direct answer path
	graph.AddEdge(directAnswerNodeKey, compose.END)

	// Plan and execute path
	graph.AddEdge(planCreationNodeKey, planExecutionNodeKey)

	// Plan execution branch
	graph.AddBranch(planExecutionNodeKey, specialistBranch)

	// All specialists go to result collector
	for _, specialist := range config.Specialists {
		graph.AddEdge(specialist.Name, resultCollectorNodeKey)
	}

	// Result collector goes to feedback processor
	graph.AddEdge(resultCollectorNodeKey, feedbackProcessorNodeKey)

	// Plan update loop
	graph.AddEdge(planUpdateNodeKey, resultCollectorNodeKey)

	// Final answer
	graph.AddEdge(finalAnswerNodeKey, compose.END)

	// Compile the graph
	runnable, err := graph.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile graph: %w", err)
	}

	return &EnhancedMultiAgent{
		runnable:         runnable,
		graph:            graph,
		graphAddNodeOpts: []compose.GraphAddNodeOpt{},
		config:           config,
	}, nil
}

func addSpecialist(graph *compose.Graph[[]*schema.Message, *schema.Message], specialist *EnhancedSpecialist) error {
	specialistHandler := NewSpecialistHandler(specialist.Name)
	if specialist.Invokable != nil || specialist.Streamable != nil {
		lambda, err := compose.AnyLambda(specialist.Invokable, specialist.Streamable, nil, nil, compose.WithLambdaType("Specialist"))
		if err != nil {
			return err
		}
		if err := graph.AddLambdaNode(specialist.Name, lambda, compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *EnhancedState) ([]*schema.Message, error) {
			return specialistHandler.PreHandler(ctx, input, state)
		}),
			compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
				return specialistHandler.PostHandler(ctx, output, state)
			}),
			compose.WithNodeName(specialist.Name), compose.WithOutputKey(specialist.Name)); err != nil {
			return err
		}
	} else if specialist.ChatModel != nil {
		if err := graph.AddChatModelNode(specialist.Name, specialist.ChatModel,
			compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *EnhancedState) ([]*schema.Message, error) {
				return specialistHandler.PreHandler(ctx, input, state)
			}),
			compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *EnhancedState) (*schema.Message, error) {
				return specialistHandler.PostHandler(ctx, output, state)
			}),
			compose.WithNodeName(specialist.Name), compose.WithOutputKey(specialist.Name),
		); err != nil {
			return err
		}
	}
	return nil
}

// Helper functions

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

func generateConversationID() string {
	return fmt.Sprintf("conv_%d", time.Now().UnixNano())
}

func buildDirectAnswerPrompt(state *EnhancedState) *schema.Message {
	prompt := fmt.Sprintf(`Provide a direct answer to the user's request.

User Intent: %s
Context: %s

Please provide a clear, helpful response.`,
		state.ConversationContext.UserIntent,
		state.ConversationContext.ContextSummary,
	)

	return &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}
}

func buildFeedbackPrompt(state *EnhancedState) []*schema.Message {
	prompt := `Analyze the execution results and provide feedback.

Results:
`
	for _, result := range state.CollectedResults {
		prompt += result.Content + "\n\n"
	}

	prompt += `
Provide feedback in JSON format:
{
  "overall_quality": 0.8,
  "should_continue": true,
  "issues": ["issue1", "issue2"],
  "suggestions": ["suggestion1", "suggestion2"],
  "confidence": 0.9,
  "next_actions": ["action1", "action2"]
}`

	return []*schema.Message{{
		Role:    schema.User,
		Content: prompt,
	}}
}

func processFeedbackResult(output *schema.Message, state *EnhancedState) error {
	// Parse feedback and update state
	// This is a simplified implementation
	feedback := map[string]any{
		"content":   output.Content,
		"timestamp": time.Now(),
	}
	state.FeedbackHistory = append(state.FeedbackHistory, feedback)
	return nil
}

func evaluateReflectionDecision(state *EnhancedState) (string, error) {
	// Simple decision logic - in practice, this would be more sophisticated
	if state.RoundNumber >= state.MaxRounds {
		return finishExecutionBranch, nil
	}

	// Check if we have good results
	if len(state.CollectedResults) > 0 {
		return finishExecutionBranch, nil
	}

	return continueExecutionBranch, nil
}

func buildPlanUpdatePrompt(state *EnhancedState) []*schema.Message {
	prompt := `Update the current plan based on feedback.

Current Plan:
`
	if state.CurrentPlan != nil {
		prompt += fmt.Sprintf("Name: %s\nDescription: %s\n", state.CurrentPlan.Name, state.CurrentPlan.Description)
	}

	prompt += `
Feedback:
`
	for _, feedback := range state.FeedbackHistory {
		if content, ok := feedback["content"].(string); ok {
			prompt += content + "\n"
		}
	}

	prompt += `
Provide an updated plan in the same JSON format as before.`

	return []*schema.Message{{
		Role:    schema.User,
		Content: prompt,
	}}
}

func processPlanUpdate(output *schema.Message, state *EnhancedState) error {
	// Parse and update the plan
	// This is a simplified implementation
	state.RoundNumber++
	return nil
}

func generateFinalAnswer(state *EnhancedState) *schema.Message {
	if state.FinalAnswer != nil {
		return state.FinalAnswer
	}

	// Generate final answer from collected results
	content := "Based on the analysis and execution:"
	for _, result := range state.CollectedResults {
		content += "\n" + result.Content
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: content,
	}
}
