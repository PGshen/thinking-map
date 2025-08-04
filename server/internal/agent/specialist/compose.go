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
	
	// 创建图
	graph := compose.NewGraph[*schema.Message, *schema.Message]()
	
	// 构建多智能体图
	if err := buildPlanningMultiAgentGraph(graph, config, options); err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}
	
	return &PlanningMultiAgent{
		config: config,
		graph:  graph,
	}, nil
}

// buildPlanningMultiAgentGraph 构建规划多智能体图
func buildPlanningMultiAgentGraph(graph *compose.Graph[*schema.Message, *schema.Message], config *PlanningMultiAgentConfig, options *options) error {
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
func addPlannerNode(graph *compose.Graph[*schema.Message, *schema.Message], planner *PlannerAgent, options *options) error {
	// 创建规划Lambda节点
	plannerLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		return executePlannerLogic(ctx, input, planner)
	})
	
	return graph.AddLambdaNode(plannerNodeKey, plannerLambda)
}

// addSpecialistNode 添加专家节点
func addSpecialistNode(graph *compose.Graph[*schema.Message, *schema.Message], specialist *Specialist, options *options) error {
	nodeKey := specialistNodeKeyPrefix + specialist.AgentMeta.Name
	
	// 根据专家类型创建不同的节点
	if specialist.ChatModel != nil {
		// 使用ChatModel的专家
		specialistLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
			return executeSpecialistWithChatModel(ctx, input, specialist)
		})
		return graph.AddLambdaNode(nodeKey, specialistLambda)
	} else if specialist.Invokable != nil {
		// 使用Invokable的专家
		specialistLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
			return executeSpecialistWithInvokable(ctx, input, specialist)
		})
		return graph.AddLambdaNode(nodeKey, specialistLambda)
	} else if specialist.Streamable != nil {
		// 使用Streamable的专家
		specialistLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
			return executeSpecialistWithStreamable(ctx, input, specialist)
		})
		return graph.AddLambdaNode(nodeKey, specialistLambda)
	}
	
	return fmt.Errorf("specialist %s has no valid execution component", specialist.AgentMeta.Name)
}

// addIteratorNode 添加迭代控制节点
func addIteratorNode(graph *compose.Graph[*schema.Message, *schema.Message], config *PlanningMultiAgentConfig, options *options) error {
	iteratorLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		return executeIteratorLogic(ctx, input, config)
	})
	
	return graph.AddLambdaNode(iteratorNodeKey, iteratorLambda)
}

// addCompletionCheckerNode 添加完成检查节点
func addCompletionCheckerNode(graph *compose.Graph[*schema.Message, *schema.Message], options *options) error {
	completionLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		return executeCompletionChecker(ctx, input)
	})
	
	return graph.AddLambdaNode(completionCheckerNodeKey, completionLambda)
}

// addSummarizerNode 添加汇总节点
func addSummarizerNode(graph *compose.Graph[*schema.Message, *schema.Message], summarizer *Summarizer, options *options) error {
	summarizerLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		return executeSummarizerLogic(ctx, input, summarizer)
	})
	
	return graph.AddLambdaNode(summarizerNodeKey, summarizerLambda)
}

// buildGraphConnections 构建图的连接关系
func buildGraphConnections(graph *compose.Graph[*schema.Message, *schema.Message], config *PlanningMultiAgentConfig) error {
	// START -> Planner
	if err := graph.AddEdge(compose.START, plannerNodeKey); err != nil {
		return err
	}
	
	// Planner -> Iterator
	if err := graph.AddEdge(plannerNodeKey, iteratorNodeKey); err != nil {
		return err
	}
	
	// Iterator -> Specialists (通过分支)
	// TODO: 实现分支逻辑，这里先简化为直接连接到完成检查节点
	// iteratorBranch := compose.NewBranch(...)
	// if err := graph.AddBranch(iteratorNodeKey, iteratorBranch); err != nil {
	//	return err
	// }
	
	// 暂时直接连接到完成检查节点
	if err := graph.AddEdge(iteratorNodeKey, completionCheckerNodeKey); err != nil {
		return err
	}
	
	// Specialists -> Iterator (形成循环)
	for _, specialist := range config.Specialists {
		nodeKey := specialistNodeKeyPrefix + specialist.AgentMeta.Name
		if err := graph.AddEdge(nodeKey, iteratorNodeKey); err != nil {
			return err
		}
	}
	
	// CompletionChecker -> END or Summarizer
	if config.Summarizer != nil {
		if err := graph.AddEdge(completionCheckerNodeKey, summarizerNodeKey); err != nil {
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

// executePlannerLogic 执行规划逻辑
func executePlannerLogic(ctx context.Context, input *schema.Message, planner *PlannerAgent) (*schema.Message, error) {
	// 解析输入状态
	state, err := parseStateFromMessage(input)
	if err != nil {
		// 如果是第一次调用，创建初始状态
		state = &PlanningState{
			OriginalQuery: input.Content,
			Messages:      []*schema.Message{input},
			IterationCount: 0,
		}
	}
	
	// 构建规划提示
	prompt := buildPlanningPrompt(planner.PlanningPrompt, state.ExecutionPlan, state.IterationCount)
	
	// 调用规划模型
	planningMessages := []*schema.Message{schema.SystemMessage(prompt)}
	response, err := planner.ChatModel.Generate(ctx, planningMessages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}
	
	// 解析执行计划
	plan, err := parsePlanFromResponse(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse plan: %w", err)
	}
	
	// 更新状态
	state.ExecutionPlan = plan
	state.Messages = append(state.Messages, response)
	
	return encodeStateToMessage(state), nil
}



// executeSpecialistWithChatModel 使用ChatModel执行专家逻辑
func executeSpecialistWithChatModel(ctx context.Context, input *schema.Message, specialist *Specialist) (*schema.Message, error) {
	state, err := parseStateFromMessage(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}
	
	// 获取当前步骤
	nextStep := state.ExecutionPlan.GetNextStep()
	if nextStep == nil {
		return nil, fmt.Errorf("no step to execute")
	}
	
	// 构建专家提示
	prompt := buildSpecialistPrompt(nextStep, specialist.SystemPrompt, state.OriginalQuery)
	
	// 调用专家模型
	specialistMessages := []*schema.Message{schema.SystemMessage(prompt)}
	response, err := specialist.ChatModel.Generate(ctx, specialistMessages)
	if err != nil {
		return nil, fmt.Errorf("failed to execute specialist: %w", err)
	}
	
	// 更新步骤状态
	state.ExecutionPlan.UpdateStepStatus(nextStep.ID, StepStatusCompleted, response.Content)
	state.Messages = append(state.Messages, response)
	
	return encodeStateToMessage(state), nil
}



// executeSpecialistWithInvokable 使用Invokable执行专家逻辑
func executeSpecialistWithInvokable(ctx context.Context, input *schema.Message, specialist *Specialist) (*schema.Message, error) {
	// TODO: 实现Invokable专家逻辑
	return nil, fmt.Errorf("invokable specialist not implemented yet")
}

// executeSpecialistWithStreamable 使用Streamable执行专家逻辑
func executeSpecialistWithStreamable(ctx context.Context, input *schema.Message, specialist *Specialist) (*schema.Message, error) {
	// TODO: 实现Streamable专家逻辑
	return nil, fmt.Errorf("streamable specialist not implemented yet")
}

// executeIteratorLogic 执行迭代控制逻辑
func executeIteratorLogic(ctx context.Context, input *schema.Message, config *PlanningMultiAgentConfig) (*schema.Message, error) {
	state, err := parseStateFromMessage(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}
	
	// 增加迭代计数
	state.IterationCount++
	
	// 检查最大迭代次数
	if state.IterationCount > config.MaxIterations {
		state.IsCompleted = true
		state.FinalResult = "达到最大迭代次数，任务终止"
	}
	
	return encodeStateToMessage(state), nil
}

// executeCompletionChecker 执行完成检查逻辑
func executeCompletionChecker(ctx context.Context, input *schema.Message) (*schema.Message, error) {
	state, err := parseStateFromMessage(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}
	
	// 检查是否所有步骤都已完成
	if state.ExecutionPlan.IsCompleted {
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
	
	return encodeStateToMessage(state), nil
}

// executeSummarizerLogic 执行汇总逻辑
func executeSummarizerLogic(ctx context.Context, input *schema.Message, summarizer *Summarizer) (*schema.Message, error) {
	state, err := parseStateFromMessage(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}
	
	// 构建汇总提示
		prompt := buildSummaryPrompt(summarizer.SummaryPrompt, state.ExecutionPlan)
	
	// 调用汇总模型
	summaryMessages := []*schema.Message{schema.SystemMessage(prompt)}
	response, err := summarizer.ChatModel.Generate(ctx, summaryMessages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}
	
	// 更新最终结果
	state.FinalResult = response.Content
	state.Messages = append(state.Messages, response)
	
	return encodeStateToMessage(state), nil
}