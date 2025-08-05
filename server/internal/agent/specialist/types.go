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

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// PlanningMultiAgent 是增强版的多智能体系统，具备规划和迭代执行能力
type PlanningMultiAgent struct {
	config   *PlanningMultiAgentConfig
	graph    *compose.Graph[[]*schema.Message, *schema.Message]
	runnable compose.Runnable[[]*schema.Message, *schema.Message]
}

// PlanningMultiAgentConfig 配置增强版多智能体系统
type PlanningMultiAgentConfig struct {
	// PlannerAgent 负责制定和更新执行计划
	PlannerAgent *PlannerAgent `json:"planner_agent"`
	// Specialists 专家智能体列表
	Specialists []*Specialist `json:"specialists"`
	// MaxIterations 最大迭代次数，防止无限循环
	MaxIterations int `json:"max_iterations"`
	// Summarizer 可选的结果汇总器
	Summarizer *Summarizer `json:"summarizer,omitempty"`
}

// PlannerAgent 规划智能体，负责制定计划和协调专家执行
type PlannerAgent struct {
	// AgentMeta 智能体元信息
	AgentMeta *AgentMeta `json:"agent_meta"`
	// ChatModel 用于规划和决策的聊天模型
	ChatModel model.BaseChatModel `json:"chat_model"`
	// PlanningPrompt 规划系统提示词
	PlanningPrompt string `json:"planning_prompt"`
	// UpdatePrompt 计划更新系统提示词
	UpdatePrompt string `json:"update_prompt"`
}

// Specialist 专家智能体
type Specialist struct {
	// AgentMeta 智能体元信息
	AgentMeta *AgentMeta `json:"agent_meta"`
	// ChatModel 专家的聊天模型
	ChatModel model.BaseChatModel `json:"chat_model,omitempty"`
	// SystemPrompt 专家的系统提示词
	SystemPrompt string `json:"system_prompt"`
	// Invokable 可调用的组件（与ChatModel二选一）
	Invokable compose.Invoke[[]*schema.Message, *schema.Message, Option] `json:"invokable,omitempty"`
	// Streamable 可流式调用的组件（与ChatModel二选一）
	Streamable compose.Stream[[]*schema.Message, *schema.Message, Option] `json:"streamable,omitempty"`
}

// AgentMeta 智能体元信息
type AgentMeta struct {
	// Name 智能体名称
	Name string `json:"name"`
	// IntendedUse 智能体的预期用途描述
	IntendedUse string `json:"intended_use"`
}

// Summarizer 结果汇总器
type Summarizer struct {
	// ChatModel 用于汇总的聊天模型
	ChatModel model.BaseChatModel `json:"chat_model"`
	// SummaryPrompt 汇总提示词
	SummaryPrompt string `json:"summary_prompt,omitempty"`
}

// ExecutionPlan 执行计划
type ExecutionPlan struct {
	// Steps 执行步骤列表
	Steps []*ExecutionStep `json:"steps"`
	// CurrentStepIndex 当前执行步骤索引
	CurrentStepIndex int `json:"current_step_index"`
	// IsCompleted 计划是否已完成
	IsCompleted bool `json:"is_completed"`
}

// ExecutionStep 执行步骤
type ExecutionStep struct {
	// ID 步骤唯一标识
	ID string `json:"id"`
	// Description 步骤描述
	Description string `json:"description"`
	// SpecialistName 负责执行的专家名称
	SpecialistName string `json:"specialist_name"`
	// Status 步骤状态
	Status StepStatus `json:"status"`
	// Input 步骤输入
	Input string `json:"input"`
	// Output 步骤输出
	Output string `json:"output,omitempty"`
	// Dependencies 依赖的步骤ID列表
	Dependencies []string `json:"dependencies,omitempty"`
}

// StepStatus 步骤状态
type StepStatus string

const (
	// StepStatusPending 待执行
	StepStatusPending StepStatus = "pending"
	// StepStatusInProgress 执行中
	StepStatusInProgress StepStatus = "in_progress"
	// StepStatusCompleted 已完成
	StepStatusCompleted StepStatus = "completed"
	// StepStatusFailed 执行失败
	StepStatusFailed StepStatus = "failed"
)

// PlanningState 规划状态，用于在图节点间传递状态
type PlanningState struct {
	// OriginalQuery 原始查询
	OriginalQuery string `json:"original_query"`
	// ExecutionPlan 当前执行计划
	ExecutionPlan *ExecutionPlan `json:"execution_plan"`
	// Messages 消息历史
	Messages []*schema.Message `json:"messages"`
	// IterationCount 当前迭代次数
	IterationCount int `json:"iteration_count"`
	// FinalResult 最终结果
	FinalResult string `json:"final_result,omitempty"`
	// IsCompleted 是否已完成
	IsCompleted bool `json:"is_completed"`
}

// Generate 生成响应
func (p *PlanningMultiAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no input messages provided")
	}

	// 获取最后一条用户消息作为查询
	var userQuery string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == schema.User {
			userQuery = messages[i].Content
			break
		}
	}

	if userQuery == "" {
		return nil, fmt.Errorf("no user query found in messages")
	}

	// 检查runnable是否已初始化
	if p.runnable == nil {
		return nil, fmt.Errorf("runnable not initialized")
	}

	// 运行图，传入消息数组
	result, err := p.runnable.Invoke(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to run graph: %w", err)
	}

	return result, nil
}

// Stream 流式生成响应
func (p *PlanningMultiAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no input messages provided")
	}

	// 检查runnable是否已初始化
	if p.runnable == nil {
		return nil, fmt.Errorf("runnable not initialized")
	}

	// 流式运行图，传入消息数组
	stream, err := p.runnable.Stream(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to stream graph: %w", err)
	}

	return stream, nil
}

// validate 验证配置
func (c *PlanningMultiAgentConfig) validate() error {
	if c.PlannerAgent == nil {
		return fmt.Errorf("planner agent is required")
	}

	if err := c.PlannerAgent.validate(); err != nil {
		return fmt.Errorf("invalid planner agent: %w", err)
	}

	if len(c.Specialists) == 0 {
		return fmt.Errorf("at least one specialist is required")
	}

	for i, specialist := range c.Specialists {
		if err := specialist.validate(); err != nil {
			return fmt.Errorf("invalid specialist at index %d: %w", i, err)
		}
	}

	if c.MaxIterations <= 0 {
		c.MaxIterations = 10 // 默认最大迭代次数
	}

	return nil
}

// validate 验证规划智能体配置
func (p *PlannerAgent) validate() error {
	if p.AgentMeta == nil {
		return fmt.Errorf("agent meta is required")
	}

	if err := p.AgentMeta.validate(); err != nil {
		return fmt.Errorf("invalid agent meta: %w", err)
	}

	if p.ChatModel == nil {
		return fmt.Errorf("chat model is required")
	}

	return nil
}

// validate 验证专家智能体配置
func (s *Specialist) validate() error {
	if s.AgentMeta == nil {
		return fmt.Errorf("agent meta is required")
	}

	if err := s.AgentMeta.validate(); err != nil {
		return fmt.Errorf("invalid agent meta: %w", err)
	}

	// 确保至少有一个执行组件
	componentCount := 0
	if s.ChatModel != nil {
		componentCount++
	}
	if s.Invokable != nil {
		componentCount++
	}
	if s.Streamable != nil {
		componentCount++
	}

	if componentCount == 0 {
		return fmt.Errorf("specialist must have at least one of: ChatModel, Invokable, or Streamable")
	}

	if componentCount > 1 {
		return fmt.Errorf("specialist can only have one of: ChatModel, Invokable, or Streamable")
	}

	return nil
}

// validate 验证智能体元信息
func (a *AgentMeta) validate() error {
	if a.Name == "" {
		return fmt.Errorf("agent name is required")
	}

	if a.IntendedUse == "" {
		return fmt.Errorf("agent intended use is required")
	}

	return nil
}

// GetNextStep 获取下一个待执行的步骤
func (p *ExecutionPlan) GetNextStep() *ExecutionStep {
	for _, step := range p.Steps {
		if step.Status == StepStatusPending {
			// 检查依赖是否都已完成
			allDependenciesCompleted := true
			for _, depID := range step.Dependencies {
				depStep := p.GetStepByID(depID)
				if depStep == nil || depStep.Status != StepStatusCompleted {
					allDependenciesCompleted = false
					break
				}
			}

			if allDependenciesCompleted {
				return step
			}
		}
	}
	return nil
}

// GetStepByID 根据ID获取步骤
func (p *ExecutionPlan) GetStepByID(id string) *ExecutionStep {
	for _, step := range p.Steps {
		if step.ID == id {
			return step
		}
	}
	return nil
}

// UpdateStepStatus 更新步骤状态
func (p *ExecutionPlan) UpdateStepStatus(stepID string, status StepStatus, output string) {
	step := p.GetStepByID(stepID)
	if step != nil {
		step.Status = status
		if output != "" {
			step.Output = output
		}
	}

	// 检查是否所有步骤都已完成
	allCompleted := true
	for _, s := range p.Steps {
		if s.Status != StepStatusCompleted {
			allCompleted = false
			break
		}
	}
	p.IsCompleted = allCompleted
}
