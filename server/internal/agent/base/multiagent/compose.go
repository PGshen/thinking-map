package multiagent

import (
	"context"
	"fmt"
	"time"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/PGshen/thinking-map/server/internal/agent/base/react"
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
	generalSpecialistNodeKey    = "general_specialist"
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
func NewMultiAgent(ctx context.Context, config *MultiAgentConfig, agentOptions ...base.AgentOption) (*MultiAgent, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	err := addDefaultSpecialist(ctx, config, agentOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to add default specialist: %w", err)
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
	err = graph.AddChatModelNode(conversationAnalyzerNodeKey, config.Host.Model,
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
	resultCollectorHandler := NewResultCollectorHandler(config)
	err = graph.AddLambdaNode(resultCollectorNodeKey,
		compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
			var result *schema.Message
			err = compose.ProcessState(ctx, func(ctx context.Context, state *MultiAgentState) error {
				// Convert single message to slice for ResultCollector
				messages := []*schema.Message{input}
				result, err = resultCollectorHandler.ResultCollector(ctx, messages, state)
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
	feedbackProcessorHandler := NewFeedbackProcessorHandler(config)
	err = graph.AddChatModelNode(feedbackProcessorNodeKey, config.Host.Model,
		compose.WithStatePreHandler(feedbackProcessorHandler.PreHandler),
		compose.WithStatePostHandler(feedbackProcessorHandler.PostHandler),
		compose.WithNodeName("feedback_processor"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add feedback processor node: %w", err)
	}

	// Add reflection branch node
	reflectionBranchHandler := NewReflectionBranchHandler(config)
	reflectionBranch := compose.NewGraphBranch(func(ctx context.Context, input *schema.Message) (string, error) {
		var result string
		err = compose.ProcessState(ctx, func(ctx context.Context, state *MultiAgentState) error {
			result = reflectionBranchHandler.evaluateReflectionDecision(state)
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
	planUpdateHandler := NewPlanUpdateHandler(config)
	err = graph.AddChatModelNode(planUpdateNodeKey, config.Host.Model,
		compose.WithStatePreHandler(planUpdateHandler.PreHandler),
		compose.WithStatePostHandler(planUpdateHandler.PostHandler),
		compose.WithNodeName("plan_update"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add plan update node: %w", err)
	}

	toFinalAnswer := compose.ToList[*schema.Message]()
	graph.AddLambdaNode(toFinalAnswerNodeKey, toFinalAnswer, compose.WithNodeName("to_final_answer"))

	// Add final answer node
	finalAnswerHandler := NewFinalAnswerHandler(config)
	err = graph.AddChatModelNode(finalAnswerNodeKey, config.Host.Model,
		compose.WithStatePreHandler(finalAnswerHandler.PreHandler),
		compose.WithStatePostHandler(finalAnswerHandler.PostHandler),
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
		Runnable:         runnable,
		Graph:            graph,
		GraphAddNodeOpts: []compose.GraphAddNodeOpt{},
		AgentOptions:     agentOptions,
		Config:           config,
	}, nil
}

// 增加一个通用的specialist, 用于处理通用任务
func addDefaultSpecialist(ctx context.Context, config *MultiAgentConfig, agentOptions ...base.AgentOption) error {
	reactAgent, err := react.NewAgent(ctx, react.ReactAgentConfig{
		ToolCallingModel: config.Host.Model,
		// todo 添加搜索工具？
	}, agentOptions...)
	if err != nil {
		return fmt.Errorf("failed to create react agent: %w", err)
	}
	config.Specialists = append(config.Specialists, &Specialist{
		Name:         generalSpecialistNodeKey,
		IntendedUse:  "General tasks",
		ReactAgent:   reactAgent,
		SystemPrompt: "You are a general specialist, you can handle any tasks.",
	})
	return nil
}

func addSpecialist(graph *compose.Graph[[]*schema.Message, *schema.Message], specialist *Specialist) error {
	specialistHandler := NewSpecialistHandler(specialist)
	if specialist.ReactAgent != nil {
		if err := graph.AddGraphNode(specialist.Name, specialist.ReactAgent.Graph,
			compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *MultiAgentState) ([]*schema.Message, error) {
				return specialistHandler.PreHandler(ctx, input, state)
			}),
			compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *MultiAgentState) (*schema.Message, error) {
				return specialistHandler.PostHandler(ctx, output, state)
			}),
			compose.WithNodeName(specialist.Name)); err != nil {
			return err
		}
	} else if specialist.Invokable != nil || specialist.Streamable != nil {
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
