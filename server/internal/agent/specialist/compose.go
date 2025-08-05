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
	"strings"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// 节点键名常量
const (
	plannerNodeKey           = "planner"
	specialistNodeKeyPrefix  = "specialist_"
	iteratorNodeKey          = "iterator"
	summarizerNodeKey        = "summarizer"
	completionCheckerNodeKey = "completion_checker"
)

// 默认提示词
const (
	defaultPlanningPrompt = `你是一个智能任务规划助手。请根据用户的查询，制定一个详细的执行计划。

可用的专家智能体：
{{specialists_info}}

请将任务分解为具体的步骤，每个步骤应该：
1. 有清晰的描述
2. 指定负责的专家
3. 明确输入要求
4. 定义依赖关系（如果有）

请以JSON格式返回执行计划：
{
  "steps": [
    {
      "id": "step_1",
      "description": "步骤描述",
      "specialist_name": "专家名称",
      "input": "输入要求",
      "dependencies": []
    }
  ]
}

用户查询：{{query}}`

	defaultUpdatePrompt = `你是一个智能任务规划助手。请根据当前的执行情况，更新执行计划。

当前执行计划：
{{current_plan}}

已完成的步骤：
{{completed_steps}}

请评估是否需要调整计划，如果需要，请提供更新后的计划。如果不需要调整，请返回原计划。

请以JSON格式返回更新后的执行计划。`

	defaultSummaryPrompt = `请根据以下执行结果，生成一个综合性的总结：

{{execution_results}}

请提供一个清晰、简洁的总结，包含主要发现和结论。`
)

// NewPlanningMultiAgent 创建增强版多智能体系统
func NewPlanningMultiAgent(config *PlanningMultiAgentConfig, opts ...Option) (*PlanningMultiAgent, error) {
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 应用选项
	options := applyOptions(opts)

	// 设置默认提示词
	if config.PlannerAgent.PlanningPrompt == "" {
		config.PlannerAgent.PlanningPrompt = defaultPlanningPrompt
	}
	if config.PlannerAgent.UpdatePrompt == "" {
		config.PlannerAgent.UpdatePrompt = defaultUpdatePrompt
	}
	if config.Summarizer != nil && config.Summarizer.SummaryPrompt == "" {
		config.Summarizer.SummaryPrompt = defaultSummaryPrompt
	}

	// 创建全局状态生成器函数
	stateGenerator := func(ctx context.Context) *PlanningState {
		return &PlanningState{
			Messages:       make([]*schema.Message, 0),
			IterationCount: 0,
			IsCompleted:    false,
			FinalResult:    "",
		}
	}

	// 创建带全局状态的图
	graph := compose.NewGraph[[]*schema.Message, *schema.Message](compose.WithGenLocalState(stateGenerator))

	// 构建多智能体图
	if err := buildPlanningMultiAgentGraph(graph, config, options); err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}

	// 编译图
	ctx := context.Background()
	compileOpts := []compose.GraphCompileOption{
		compose.WithMaxRunSteps(config.MaxIterations),
		compose.WithNodeTriggerMode(compose.AnyPredecessor),
		compose.WithGraphName("PlanningMultiAgentGraph"),
	}
	runnable, err := graph.Compile(ctx, compileOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile graph: %w", err)
	}

	return &PlanningMultiAgent{
		config:   config,
		graph:    graph,
		runnable: runnable,
	}, nil
}

// buildPlanningMultiAgentGraph 构建规划多智能体图
func buildPlanningMultiAgentGraph(graph *compose.Graph[[]*schema.Message, *schema.Message], config *PlanningMultiAgentConfig, options *options) error {
	// 1. 添加规划节点
	if err := addPlannerNode(graph, config.PlannerAgent, options); err != nil {
		return fmt.Errorf("failed to add planner node: %w", err)
	}

	// 2. 添加专家节点
	for _, specialist := range config.Specialists {
		if err := addSpecialistNode(graph, specialist, options); err != nil {
			return fmt.Errorf("failed to add specialist node %s: %w", specialist.AgentMeta.Name, err)
		}
	}

	// 3. 添加迭代控制节点
	if err := addIteratorNode(graph, config, options); err != nil {
		return fmt.Errorf("failed to add iterator node: %w", err)
	}

	// 4. 添加完成检查节点
	if err := addCompletionCheckerNode(graph, options); err != nil {
		return fmt.Errorf("failed to add completion checker node: %w", err)
	}

	// 5. 添加汇总节点（如果配置了）
	if config.Summarizer != nil {
		if err := addSummarizerNode(graph, config.Summarizer, options); err != nil {
			return fmt.Errorf("failed to add summarizer node: %w", err)
		}
	}

	// 6. 构建图的连接关系
	if err := buildGraphConnections(graph, config); err != nil {
		return fmt.Errorf("failed to build graph connections: %w", err)
	}

	return nil
}

// addPlannerNode 添加规划节点
func addPlannerNode(graph *compose.Graph[[]*schema.Message, *schema.Message], planner *PlannerAgent, options *options) error {
	// 使用ChatModel节点而不是Lambda节点
	return graph.AddChatModelNode(plannerNodeKey, planner.ChatModel,
		compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *PlanningState) ([]*schema.Message, error) {
			return plannerNodePreHandler(ctx, input, state, planner)
		}),
		compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *PlanningState) (*schema.Message, error) {
			return plannerNodePostHandler(ctx, output, state)
		}),
	)
}

// addSpecialistNode 添加专家节点
func addSpecialistNode(graph *compose.Graph[[]*schema.Message, *schema.Message], specialist *Specialist, options *options) error {
	nodeKey := specialistNodeKeyPrefix + specialist.AgentMeta.Name

	// 根据专家类型创建不同的节点
	if specialist.ChatModel != nil {
		// 使用ChatModel的专家
		return graph.AddChatModelNode(nodeKey, specialist.ChatModel,
			compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *PlanningState) ([]*schema.Message, error) {
				return specialistNodePreHandler(ctx, input, state, specialist)
			}),
			compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *PlanningState) (*schema.Message, error) {
				return specialistNodePostHandler(ctx, output, state)
			}),
			compose.WithNodeName(specialist.AgentMeta.Name),
		)
	} else if specialist.Invokable != nil || specialist.Streamable != nil {
		lambda, err := compose.AnyLambda(specialist.Invokable, specialist.Streamable, nil, nil, compose.WithLambdaType("Specialist"))
		if err != nil {
			return err
		}
		if err := graph.AddLambdaNode(nodeKey, lambda, compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *PlanningState) ([]*schema.Message, error) {
			return specialistNodePreHandler(ctx, input, state, specialist)
		}),
			compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *PlanningState) (*schema.Message, error) {
				return specialistNodePostHandler(ctx, output, state)
			}),
			compose.WithNodeName(specialist.AgentMeta.Name)); err != nil {
			return err
		}
	}

	return fmt.Errorf("specialist %s has no valid execution component", specialist.AgentMeta.Name)
}

// addIteratorNode 添加迭代控制节点
func addIteratorNode(graph *compose.Graph[[]*schema.Message, *schema.Message], config *PlanningMultiAgentConfig, options *options) error {
	// 迭代器节点使用Lambda节点，因为它不需要ChatModel
	iteratorLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		return input, nil
	})

	return graph.AddLambdaNode(iteratorNodeKey, iteratorLambda,
		compose.WithStatePreHandler(func(ctx context.Context, input *schema.Message, state *PlanningState) (*schema.Message, error) {
			return iteratorNodePreHandler(ctx, input, state, config)
		}),
	)
}

// addCompletionCheckerNode 添加完成检查节点
func addCompletionCheckerNode(graph *compose.Graph[[]*schema.Message, *schema.Message], options *options) error {
	// 完成检查器节点使用Lambda节点，因为它不需要ChatModel
	completionLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		return input, nil
	})

	return graph.AddLambdaNode(completionCheckerNodeKey, completionLambda,
		compose.WithStatePreHandler(func(ctx context.Context, input *schema.Message, state *PlanningState) (*schema.Message, error) {
			return completionCheckerNodePreHandler(ctx, input, state)
		}),
	)
}

// addSummarizerNode 添加汇总节点
func addSummarizerNode(graph *compose.Graph[[]*schema.Message, *schema.Message], summarizer *Summarizer, options *options) error {
	// 使用ChatModel节点而不是Lambda节点
	return graph.AddChatModelNode(summarizerNodeKey, summarizer.ChatModel,
		compose.WithStatePreHandler(func(ctx context.Context, input []*schema.Message, state *PlanningState) ([]*schema.Message, error) {
			return summarizerNodePreHandler(ctx, input, state, summarizer)
		}),
		compose.WithStatePostHandler(func(ctx context.Context, output *schema.Message, state *PlanningState) (*schema.Message, error) {
			return summarizerNodePostHandler(ctx, output, state)
		}),
	)
}

// buildGraphConnections 构建图的连接关系
func buildGraphConnections(graph *compose.Graph[[]*schema.Message, *schema.Message], config *PlanningMultiAgentConfig) error {
	// START -> Planner
	if err := graph.AddEdge(compose.START, plannerNodeKey); err != nil {
		return err
	}

	// 添加消息转换节点，将单个消息转换为消息数组
	msg2ListLambda := compose.ToList[*schema.Message]()
	if err := graph.AddLambdaNode("msg2list", msg2ListLambda); err != nil {
		return err
	}

	// Planner -> Iterator
	if err := graph.AddEdge(plannerNodeKey, iteratorNodeKey); err != nil {
		return err
	}

	// Iterator -> Specialists (通过分支)
	iteratorBranch := compose.NewStreamGraphBranch(
		func(ctx context.Context, msgStream *schema.StreamReader[*schema.Message]) (endNode string, err error) {
			msgStream.Close()
			
			// 获取全局状态
			var nextSpecialistNode string
			err = compose.ProcessState(ctx, func(_ context.Context, state *PlanningState) error {
				if state.ExecutionPlan == nil {
					return fmt.Errorf("no execution plan available")
				}
				
				// 获取下一个待执行的步骤
				nextStep := state.ExecutionPlan.GetNextStep()
				if nextStep == nil {
					// 没有待执行步骤，转到完成检查节点
					nextSpecialistNode = completionCheckerNodeKey
					return nil
				}
				
				// 根据步骤的specialist_name路由到对应的specialist节点
				nextSpecialistNode = specialistNodeKeyPrefix + nextStep.SpecialistName
				return nil
			})
			
			if err != nil {
				return "", err
			}
			
			return nextSpecialistNode, nil
		},
		// 定义可能的目标节点
		func() map[string]bool {
			targets := make(map[string]bool)
			targets[completionCheckerNodeKey] = true
			// 添加所有specialist节点作为可能的目标
			for _, specialist := range config.Specialists {
				targets[specialistNodeKeyPrefix+specialist.AgentMeta.Name] = true
			}
			return targets
		}(),
	)
	
	if err := graph.AddBranch(iteratorNodeKey, iteratorBranch); err != nil {
		return err
	}

	// Iterator -> CompletionChecker 的连接现在通过分支处理

	// 为specialist节点添加转换器
	for _, specialist := range config.Specialists {
		nodeKey := specialistNodeKeyPrefix + specialist.AgentMeta.Name
		// Specialists -> Iterator (形成循环)
		if err := graph.AddEdge(nodeKey, iteratorNodeKey); err != nil {
			return err
		}
	}

	// CompletionChecker -> END or Summarizer
	if config.Summarizer != nil {
		// 添加completion checker转换器，将*schema.Message转换为[]*schema.Message
		completionConverterLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) ([]*schema.Message, error) {
			return []*schema.Message{input}, nil
		})
		if err := graph.AddLambdaNode("completion_converter", completionConverterLambda); err != nil {
			return err
		}

		if err := graph.AddEdge(completionCheckerNodeKey, "completion_converter"); err != nil {
			return err
		}
		if err := graph.AddEdge("completion_converter", summarizerNodeKey); err != nil {
			return err
		}
		if err := graph.AddEdge(summarizerNodeKey, compose.END); err != nil {
			return err
		}
	} else {
		if err := graph.AddEdge(completionCheckerNodeKey, compose.END); err != nil {
			return err
		}
	}

	return nil
}

// 执行逻辑函数

// plannerNodePreHandler 规划节点前置处理器
func plannerNodePreHandler(ctx context.Context, input []*schema.Message, state *PlanningState, planner *PlannerAgent) ([]*schema.Message, error) {
	// 初始化状态（如果是第一次调用）
	if len(input) > 0 && state.OriginalQuery == "" {
		lastMessage := input[len(input)-1]
		state.OriginalQuery = lastMessage.Content
		state.Messages = make([]*schema.Message, len(input))
		copy(state.Messages, input)
	}

	// 构建规划提示
	prompt := buildPlanningPrompt(planner.PlanningPrompt, state.ExecutionPlan, state.IterationCount)

	// 创建规划消息
	planningMessages := []*schema.Message{schema.SystemMessage(prompt)}
	if len(state.Messages) > 0 {
		planningMessages = append(planningMessages, state.Messages...)
	}

	return planningMessages, nil
}

// plannerNodePostHandler 规划节点后置处理器
func plannerNodePostHandler(ctx context.Context, output *schema.Message, state *PlanningState) (*schema.Message, error) {
	if output == nil {
		return nil, fmt.Errorf("no planning response generated")
	}

	// 解析执行计划
	plan, err := parsePlanFromResponse(output.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse plan: %w", err)
	}

	// 更新状态
	state.ExecutionPlan = plan
	state.Messages = append(state.Messages, output)

	return output, nil
}

// specialistNodePreHandler 专家节点前置处理器
func specialistNodePreHandler(ctx context.Context, input []*schema.Message, state *PlanningState, specialist *Specialist) ([]*schema.Message, error) {
	// 获取当前步骤
	nextStep := state.ExecutionPlan.GetNextStep()
	if nextStep == nil {
		return nil, fmt.Errorf("no step to execute")
	}

	// 构建专家提示
	prompt := buildSpecialistPrompt(nextStep, specialist.SystemPrompt, state.OriginalQuery)

	// 创建专家消息
	specialistMessages := []*schema.Message{schema.SystemMessage(prompt)}
	specialistMessages = append(specialistMessages, input...)

	return specialistMessages, nil
}

// specialistNodePostHandler 专家节点后置处理器
func specialistNodePostHandler(ctx context.Context, output *schema.Message, state *PlanningState) (*schema.Message, error) {
	if output == nil {
		return nil, fmt.Errorf("no specialist response generated")
	}

	// 获取当前步骤
	nextStep := state.ExecutionPlan.GetNextStep()
	if nextStep != nil {
		// 更新步骤状态
		state.ExecutionPlan.UpdateStepStatus(nextStep.ID, StepStatusCompleted, output.Content)
	}

	// 更新消息历史
	state.Messages = append(state.Messages, output)

	return output, nil
}

// iteratorNodePreHandler 迭代节点前置处理器
func iteratorNodePreHandler(ctx context.Context, input *schema.Message, state *PlanningState, config *PlanningMultiAgentConfig) (*schema.Message, error) {
	// 增加迭代计数
	state.IterationCount++

	// 检查最大迭代次数
	if state.IterationCount > config.MaxIterations {
		state.IsCompleted = true
		state.FinalResult = "达到最大迭代次数，任务终止"
	}

	return input, nil
}

// completionCheckerNodePreHandler 完成检查节点前置处理器
func completionCheckerNodePreHandler(ctx context.Context, input *schema.Message, state *PlanningState) (*schema.Message, error) {
	// 检查是否所有步骤都已完成
	if state.ExecutionPlan != nil && state.ExecutionPlan.IsCompleted {
		state.IsCompleted = true
		// 收集所有步骤的输出作为最终结果
		var results []string
		for _, step := range state.ExecutionPlan.Steps {
			if step.Output != "" {
				results = append(results, fmt.Sprintf("%s: %s", step.Description, step.Output))
			}
		}
		state.FinalResult = strings.Join(results, "\n\n")
	}

	return input, nil
}

// summarizerNodePreHandler 汇总节点前置处理器
func summarizerNodePreHandler(ctx context.Context, input []*schema.Message, state *PlanningState, summarizer *Summarizer) ([]*schema.Message, error) {
	// 构建汇总提示
	prompt := buildSummaryPrompt(summarizer.SummaryPrompt, state.ExecutionPlan)

	// 创建汇总消息
	summaryMessages := []*schema.Message{schema.SystemMessage(prompt)}
	if len(state.Messages) > 0 {
		summaryMessages = append(summaryMessages, state.Messages...)
	}

	return summaryMessages, nil
}

// summarizerNodePostHandler 汇总节点后置处理器
func summarizerNodePostHandler(ctx context.Context, output *schema.Message, state *PlanningState) (*schema.Message, error) {
	if output == nil {
		return nil, fmt.Errorf("no summarizer response generated")
	}

	// 更新最终结果
	state.FinalResult = output.Content
	state.Messages = append(state.Messages, output)

	return output, nil
}
