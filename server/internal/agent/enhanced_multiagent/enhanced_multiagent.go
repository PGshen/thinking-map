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

package enhanced_multiagent

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

var registerStateOnce sync.Once

// EnhancedMultiAgentState 增强版多智能体状态
type EnhancedMultiAgentState struct {
	// 主控React Agent的消息历史
	HostMessages []*schema.Message
	// 专家代理的执行结果
	SpecialistResults map[string]*schema.Message
	// 当前执行计划
	ExecutionPlan string
	// 任务复杂度评估
	TaskComplexity string
	// 是否需要专家协助
	NeedsSpecialistHelp bool
	// 反思次数
	ReflectionCount int
	// 最大反思次数
	MaxReflections int
	// 当前处理阶段
	CurrentPhase string
	// 专家调用历史
	SpecialistCallHistory []string
}

// SpecialistConfig 专家代理配置
type SpecialistConfig struct {
	// 专家名称
	Name string
	// 专家描述
	Description string
	// 专家的聊天模型
	ChatModel model.ChatModel
	// 专家的工具调用模型
	ToolCallingModel model.ToolCallingChatModel
	// 专家的工具配置
	ToolsConfig *compose.ToolsNodeConfig
	// 专家的系统提示
	SystemPrompt string
	// 是否支持流式输出
	Streamable bool
}

// EnhancedMultiAgentConfig 增强版多智能体系统配置
type EnhancedMultiAgentConfig struct {
	// 主控React Agent配置
	HostReactConfig *react.AgentConfig
	// 专家代理配置列表
	Specialists []*SpecialistConfig
	// 任务规划模型
	PlanningModel model.ChatModel
	// 复杂度评估模型
	ComplexityEvaluationModel model.ChatModel
	// 反思模型
	ReflectionModel model.ChatModel
	// 最大反思次数
	MaxReflections int
	// 最大执行步数
	MaxSteps int
	// 图名称
	GraphName string
	// 回调处理器
	Callback EnhancedMultiAgentCallback
}

// EnhancedMultiAgentCallback 增强版多智能体回调接口
type EnhancedMultiAgentCallback interface {
	// 任务规划回调
	OnTaskPlanning(ctx context.Context, task string, plan string) error
	// 复杂度评估回调
	OnComplexityEvaluation(ctx context.Context, task string, complexity string) error
	// 专家调用回调
	OnSpecialistCall(ctx context.Context, specialistName string, input *schema.Message) error
	// 专家响应回调
	OnSpecialistResponse(ctx context.Context, specialistName string, response *schema.Message) error
	// 反思回调
	OnReflection(ctx context.Context, reflectionInput string, reflectionOutput string) error
	// 阶段切换回调
	OnPhaseTransition(ctx context.Context, fromPhase string, toPhase string) error
}

// EnhancedMultiAgent 增强版多智能体系统
type EnhancedMultiAgent struct {
	runnable compose.Runnable[[]*schema.Message, *schema.Message]
	graph    *compose.Graph[[]*schema.Message, *schema.Message]
	config   *EnhancedMultiAgentConfig
}

// 节点键常量
const (
	// 主要节点
	nodeKeyComplexityEvaluation = "complexity_evaluation"
	nodeKeyTaskPlanning         = "task_planning"
	nodeKeyHostReact           = "host_react"
	nodeKeyReflection          = "reflection"
	nodeKeySpecialistRouter    = "specialist_router"
	nodeKeyResultIntegration   = "result_integration"
	
	// 专家节点前缀
	nodeKeySpecialistPrefix = "specialist_"
	
	// 阶段常量
	phaseComplexityEvaluation = "complexity_evaluation"
	phaseTaskPlanning        = "task_planning"
	phaseHostProcessing      = "host_processing"
	phaseSpecialistProcessing = "specialist_processing"
	phaseReflection          = "reflection"
	phaseResultIntegration   = "result_integration"
	phaseCompleted           = "completed"
)

// 默认提示模板
const (
	defaultComplexityEvaluationPrompt = `请评估以下任务的复杂度：

任务：{{.Task}}

请从以下选项中选择：
- SIMPLE: 简单任务，可以直接回答
- MODERATE: 中等复杂度，需要一定的推理和规划
- COMPLEX: 复杂任务，需要多步骤规划和专家协助

只返回复杂度级别（SIMPLE/MODERATE/COMPLEX）：`

	defaultTaskPlanningPrompt = `基于以下信息制定详细的执行计划：

任务：{{.Task}}
复杂度：{{.Complexity}}
可用专家：{{.Specialists}}

请制定一个详细的执行计划，包括：
1. 任务分解
2. 执行步骤
3. 需要调用的专家（如果有）
4. 预期结果

执行计划：`

	defaultReflectionPrompt = `请基于当前上下文进行反思和分析：

当前任务：{{.Task}}
执行历史：{{.History}}
专家结果：{{.SpecialistResults}}
当前状态：{{.CurrentState}}

请分析：
1. 当前进展如何？
2. 是否需要调整策略？
3. 下一步应该怎么做？
4. 是否需要更多专家协助？

反思结果：`

	defaultSpecialistRouterPrompt = `基于当前上下文，决定是否需要调用专家：

当前任务：{{.Task}}
执行计划：{{.Plan}}
当前消息：{{.CurrentMessage}}
可用专家：{{.Specialists}}

如果需要专家协助，请返回专家名称，否则返回 "NONE"。
如果需要多个专家，请用逗号分隔。

专家选择：`
)

// NewEnhancedMultiAgent 创建增强版多智能体系统
func NewEnhancedMultiAgent(ctx context.Context, config *EnhancedMultiAgentConfig) (*EnhancedMultiAgent, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	
	if config.HostReactConfig == nil {
		return nil, fmt.Errorf("host react config cannot be nil")
	}
	
	if config.MaxReflections <= 0 {
		config.MaxReflections = 3
	}
	
	if config.MaxSteps <= 0 {
		config.MaxSteps = 20
	}
	
	if config.GraphName == "" {
		config.GraphName = "EnhancedMultiAgent"
	}
	
	// 注册状态类型（只注册一次）
	var registerErr error
	registerStateOnce.Do(func() {
		registerErr = compose.RegisterSerializableType[EnhancedMultiAgentState]("_eino_enhanced_multiagent_state")
	})
	if registerErr != nil {
		return nil, fmt.Errorf("failed to register state type: %w", registerErr)
	}
	
	// 创建图
	graph := compose.NewGraph[[]*schema.Message, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *EnhancedMultiAgentState {
			return &EnhancedMultiAgentState{
				HostMessages:          make([]*schema.Message, 0),
				SpecialistResults:     make(map[string]*schema.Message),
				ExecutionPlan:         "",
				TaskComplexity:        "",
				NeedsSpecialistHelp:   false,
				ReflectionCount:       0,
				MaxReflections:        config.MaxReflections,
				CurrentPhase:          phaseComplexityEvaluation,
				SpecialistCallHistory: make([]string, 0),
			}
		}),
	)
	
	// 构建图结构
	enhancedAgent := &EnhancedMultiAgent{
		graph:  graph,
		config: config,
	}
	
	if err := enhancedAgent.buildGraph(ctx); err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}
	
	// 编译图
	compileOpts := []compose.GraphCompileOption{
		compose.WithMaxRunSteps(config.MaxSteps),
		compose.WithNodeTriggerMode(compose.AnyPredecessor),
		compose.WithGraphName(config.GraphName),
	}
	
	runnable, err := graph.Compile(ctx, compileOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile graph: %w", err)
	}
	
	enhancedAgent.runnable = runnable
	return enhancedAgent, nil
}

// Generate 生成响应
func (e *EnhancedMultiAgent) Generate(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
	return e.runnable.Invoke(ctx, input)
}

// Stream 流式生成响应
func (e *EnhancedMultiAgent) Stream(ctx context.Context, input []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return e.runnable.Stream(ctx, input)
}

// ExportGraph 导出图结构
func (e *EnhancedMultiAgent) ExportGraph() (compose.AnyGraph, []compose.GraphAddNodeOpt) {
	return e.graph, []compose.GraphAddNodeOpt{}
}

// buildGraph 构建图结构
func (e *EnhancedMultiAgent) buildGraph(ctx context.Context) error {
	// 添加复杂度评估节点
	if err := e.addComplexityEvaluationNode(); err != nil {
		return fmt.Errorf("failed to add complexity evaluation node: %w", err)
	}
	
	// 添加任务规划节点
	if err := e.addTaskPlanningNode(); err != nil {
		return fmt.Errorf("failed to add task planning node: %w", err)
	}
	
	// 添加主控React Agent节点
	if err := e.addHostReactNode(ctx); err != nil {
		return fmt.Errorf("failed to add host react node: %w", err)
	}
	
	// 添加专家路由节点
	if err := e.addSpecialistRouterNode(); err != nil {
		return fmt.Errorf("failed to add specialist router node: %w", err)
	}
	
	// 添加专家节点
	if err := e.addSpecialistNodes(ctx); err != nil {
		return fmt.Errorf("failed to add specialist nodes: %w", err)
	}
	
	// 添加反思节点
	if err := e.addReflectionNode(); err != nil {
		return fmt.Errorf("failed to add reflection node: %w", err)
	}
	
	// 添加结果集成节点
	if err := e.addResultIntegrationNode(); err != nil {
		return fmt.Errorf("failed to add result integration node: %w", err)
	}
	
	// 添加边和分支
	if err := e.addEdgesAndBranches(); err != nil {
		return fmt.Errorf("failed to add edges and branches: %w", err)
	}
	
	return nil
}