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
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// addComplexityEvaluationNode 添加复杂度评估节点
func (e *EnhancedMultiAgent) addComplexityEvaluationNode() error {
	if e.config.ComplexityEvaluationModel == nil {
		return fmt.Errorf("complexity evaluation model is required")
	}
	
	complexityEvaluationLambda := compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
		var taskContent string
		if len(input) > 0 && input[len(input)-1].Content != "" {
			taskContent = input[len(input)-1].Content
		} else {
			taskContent = "未指定任务"
		}
		
		// 构建复杂度评估提示
		prompt := strings.ReplaceAll(defaultComplexityEvaluationPrompt, "{{.Task}}", taskContent)
		messages := []*schema.Message{
			{Role: schema.System, Content: prompt},
		}
		
		// 调用复杂度评估模型
		response, err := e.config.ComplexityEvaluationModel.Generate(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("complexity evaluation failed: %w", err)
		}
		
		// 更新状态
		err = compose.ProcessState[*EnhancedMultiAgentState](ctx, func(ctx context.Context, state *EnhancedMultiAgentState) error {
			state.TaskComplexity = strings.TrimSpace(response.Content)
			state.CurrentPhase = phaseTaskPlanning
			state.HostMessages = append(state.HostMessages, input...)
			
			// 触发回调
			if e.config.Callback != nil {
				if err := e.config.Callback.OnComplexityEvaluation(ctx, taskContent, state.TaskComplexity); err != nil {
					return fmt.Errorf("complexity evaluation callback failed: %w", err)
				}
			}
			
			return nil
		})
		
		if err != nil {
			return nil, err
		}
		
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("任务复杂度评估完成：%s", response.Content),
		}, nil
	})
	
	return e.graph.AddLambdaNode(nodeKeyComplexityEvaluation, complexityEvaluationLambda)
}

// addTaskPlanningNode 添加任务规划节点
func (e *EnhancedMultiAgent) addTaskPlanningNode() error {
	if e.config.PlanningModel == nil {
		return fmt.Errorf("planning model is required")
	}
	
	taskPlanningLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		var taskContent string
		var complexity string
		var specialists []string
		
		// 从状态中获取信息
		err := compose.ProcessState[*EnhancedMultiAgentState](ctx, func(ctx context.Context, state *EnhancedMultiAgentState) error {
			if len(state.HostMessages) > 0 {
				taskContent = state.HostMessages[len(state.HostMessages)-1].Content
			}
			complexity = state.TaskComplexity
			
			// 收集专家信息
			for _, spec := range e.config.Specialists {
				specialists = append(specialists, fmt.Sprintf("%s: %s", spec.Name, spec.Description))
			}
			
			return nil
		})
		
		if err != nil {
			return nil, err
		}
		
		// 构建任务规划提示
		prompt := defaultTaskPlanningPrompt
		prompt = strings.ReplaceAll(prompt, "{{.Task}}", taskContent)
		prompt = strings.ReplaceAll(prompt, "{{.Complexity}}", complexity)
		prompt = strings.ReplaceAll(prompt, "{{.Specialists}}", strings.Join(specialists, "\n"))
		
		messages := []*schema.Message{
			{Role: schema.System, Content: prompt},
		}
		
		// 调用任务规划模型
		response, err := e.config.PlanningModel.Generate(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("task planning failed: %w", err)
		}
		
		// 更新状态
		err = compose.ProcessState[*EnhancedMultiAgentState](ctx, func(ctx context.Context, state *EnhancedMultiAgentState) error {
			state.ExecutionPlan = response.Content
			state.CurrentPhase = phaseHostProcessing
			
			// 触发回调
			if e.config.Callback != nil {
				if err := e.config.Callback.OnTaskPlanning(ctx, taskContent, state.ExecutionPlan); err != nil {
					return fmt.Errorf("task planning callback failed: %w", err)
				}
			}
			
			return nil
		})
		
		if err != nil {
			return nil, err
		}
		
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("任务规划完成：%s", response.Content),
		}, nil
	})
	
	return e.graph.AddLambdaNode(nodeKeyTaskPlanning, taskPlanningLambda)
}

// addHostReactNode 添加主控React Agent节点
func (e *EnhancedMultiAgent) addHostReactNode(ctx context.Context) error {
	// 创建React Agent
	reactAgent, err := react.NewAgent(ctx, e.config.HostReactConfig)
	if err != nil {
		return fmt.Errorf("failed to create react agent: %w", err)
	}
	
	// 包装React Agent为Lambda
	hostReactLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		var messages []*schema.Message
		
		// 从状态中获取消息历史
		err := compose.ProcessState[*EnhancedMultiAgentState](ctx, func(ctx context.Context, state *EnhancedMultiAgentState) error {
			// 构建完整的消息历史
			messages = make([]*schema.Message, len(state.HostMessages))
			copy(messages, state.HostMessages)
			
			// 添加执行计划作为系统消息
			if state.ExecutionPlan != "" {
				planMessage := &schema.Message{
					Role:    schema.System,
					Content: fmt.Sprintf("执行计划：%s", state.ExecutionPlan),
				}
				messages = append([]*schema.Message{planMessage}, messages...)
			}
			
			// 添加专家结果作为上下文
			if len(state.SpecialistResults) > 0 {
				var specialistContext []string
				for name, result := range state.SpecialistResults {
					specialistContext = append(specialistContext, fmt.Sprintf("%s: %s", name, result.Content))
				}
				contextMessage := &schema.Message{
					Role:    schema.System,
					Content: fmt.Sprintf("专家结果：\n%s", strings.Join(specialistContext, "\n")),
				}
				messages = append([]*schema.Message{contextMessage}, messages...)
			}
			
			return nil
		})
		
		if err != nil {
			return nil, err
		}
		
		// 调用React Agent
		response, err := reactAgent.Generate(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("react agent failed: %w", err)
		}
		
		// 更新状态
		err = compose.ProcessState[*EnhancedMultiAgentState](ctx, func(ctx context.Context, state *EnhancedMultiAgentState) error {
			state.HostMessages = append(state.HostMessages, response)
			return nil
		})
		
		if err != nil {
			return nil, err
		}
		
		return response, nil
	})
	
	return e.graph.AddLambdaNode(nodeKeyHostReact, hostReactLambda)
}

// addSpecialistRouterNode 添加专家路由节点
func (e *EnhancedMultiAgent) addSpecialistRouterNode() error {
	specialistRouterLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		var taskContent string
		var executionPlan string
		var specialists []string
		
		// 从状态中获取信息
		err := compose.ProcessState[*EnhancedMultiAgentState](ctx, func(ctx context.Context, state *EnhancedMultiAgentState) error {
			if len(state.HostMessages) > 0 {
				taskContent = state.HostMessages[0].Content // 原始任务
			}
			executionPlan = state.ExecutionPlan
			
			// 收集专家信息
			for _, spec := range e.config.Specialists {
				specialists = append(specialists, fmt.Sprintf("%s: %s", spec.Name, spec.Description))
			}
			
			return nil
		})
		
		if err != nil {
			return nil, err
		}
		
		// 构建专家路由提示
		prompt := defaultSpecialistRouterPrompt
		prompt = strings.ReplaceAll(prompt, "{{.Task}}", taskContent)
		prompt = strings.ReplaceAll(prompt, "{{.Plan}}", executionPlan)
		prompt = strings.ReplaceAll(prompt, "{{.CurrentMessage}}", input.Content)
		prompt = strings.ReplaceAll(prompt, "{{.Specialists}}", strings.Join(specialists, "\n"))
		
		messages := []*schema.Message{
			{Role: schema.System, Content: prompt},
		}
		
		// 使用主控模型进行路由决策
		var routingModel model.ChatModel
		if e.config.HostReactConfig.Model != nil {
			routingModel = e.config.HostReactConfig.Model
		} else {
			return nil, fmt.Errorf("no available model for routing")
		}
		
		response, err := routingModel.Generate(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("specialist routing failed: %w", err)
		}
		
		// 解析路由结果
		selectedSpecialists := strings.TrimSpace(response.Content)
		
		// 更新状态
		err = compose.ProcessState[*EnhancedMultiAgentState](ctx, func(ctx context.Context, state *EnhancedMultiAgentState) error {
			if selectedSpecialists != "NONE" {
				state.NeedsSpecialistHelp = true
				state.CurrentPhase = phaseSpecialistProcessing
				// 记录专家调用历史，只保留有效的专家名称
				specialistNames := strings.Split(selectedSpecialists, ",")
				for _, name := range specialistNames {
					name = strings.TrimSpace(name)
					if name != "" {
						// 验证专家名称是否存在于配置中
						for _, specialist := range e.config.Specialists {
							if specialist.Name == name {
								state.SpecialistCallHistory = append(state.SpecialistCallHistory, name)
								break
							}
						}
					}
				}
			} else {
				state.NeedsSpecialistHelp = false
				state.CurrentPhase = phaseReflection
			}
			return nil
		})
		
		if err != nil {
			return nil, err
		}
		
		return &schema.Message{
			Role:    schema.Assistant,
			Content: fmt.Sprintf("专家路由决策：%s", selectedSpecialists),
		}, nil
	})
	
	return e.graph.AddLambdaNode(nodeKeySpecialistRouter, specialistRouterLambda)
}

// addSpecialistNodes 添加专家节点
func (e *EnhancedMultiAgent) addSpecialistNodes(ctx context.Context) error {
	for _, specialist := range e.config.Specialists {
		if err := e.addSpecialistNode(ctx, specialist); err != nil {
			return fmt.Errorf("failed to add specialist node %s: %w", specialist.Name, err)
		}
	}
	return nil
}

// addSpecialistNode 添加单个专家节点
func (e *EnhancedMultiAgent) addSpecialistNode(ctx context.Context, specialist *SpecialistConfig) error {
	nodeKey := nodeKeySpecialistPrefix + specialist.Name
	
	specialistLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		// 构建专家输入消息
		messages := []*schema.Message{
			{Role: schema.System, Content: specialist.SystemPrompt},
			input,
		}
		
		// 触发专家调用回调
		if e.config.Callback != nil {
			if err := e.config.Callback.OnSpecialistCall(ctx, specialist.Name, input); err != nil {
				return nil, fmt.Errorf("specialist call callback failed: %w", err)
			}
		}
		
		// 选择合适的模型
		var response *schema.Message
		var err error
		
		if specialist.ToolCallingModel != nil {
			response, err = specialist.ToolCallingModel.Generate(ctx, messages)
		} else if specialist.ChatModel != nil {
			response, err = specialist.ChatModel.Generate(ctx, messages)
		} else {
			return nil, fmt.Errorf("specialist %s has no available model", specialist.Name)
		}
		
		if err != nil {
			return nil, fmt.Errorf("specialist %s generation failed: %w", specialist.Name, err)
		}
		
		// 触发专家响应回调
		if e.config.Callback != nil {
			if err := e.config.Callback.OnSpecialistResponse(ctx, specialist.Name, response); err != nil {
				return nil, fmt.Errorf("specialist response callback failed: %w", err)
			}
		}
		
		// 更新状态
		err = compose.ProcessState[*EnhancedMultiAgentState](ctx, func(ctx context.Context, state *EnhancedMultiAgentState) error {
			state.SpecialistResults[specialist.Name] = response
			return nil
		})
		
		if err != nil {
			return nil, err
		}
		
		return response, nil
	})
	
	return e.graph.AddLambdaNode(nodeKey, specialistLambda)
}

// addReflectionNode 添加反思节点
func (e *EnhancedMultiAgent) addReflectionNode() error {
	if e.config.ReflectionModel == nil {
		return fmt.Errorf("reflection model is required")
	}
	
	reflectionLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		var taskContent string
		var history []string
		var specialistResults []string
		var currentState string
		
		// 从状态中获取信息
		err := compose.ProcessState[*EnhancedMultiAgentState](ctx, func(ctx context.Context, state *EnhancedMultiAgentState) error {
			if len(state.HostMessages) > 0 {
				taskContent = state.HostMessages[0].Content
			}
			
			// 构建历史记录
			for i, msg := range state.HostMessages {
				history = append(history, fmt.Sprintf("Step %d: %s", i+1, msg.Content))
			}
			
			// 构建专家结果
			for name, result := range state.SpecialistResults {
				specialistResults = append(specialistResults, fmt.Sprintf("%s: %s", name, result.Content))
			}
			
			currentState = fmt.Sprintf("Phase: %s, Reflections: %d/%d", state.CurrentPhase, state.ReflectionCount, state.MaxReflections)
			
			return nil
		})
		
		if err != nil {
			return nil, err
		}
		
		// 构建反思提示
		prompt := defaultReflectionPrompt
		prompt = strings.ReplaceAll(prompt, "{{.Task}}", taskContent)
		prompt = strings.ReplaceAll(prompt, "{{.History}}", strings.Join(history, "\n"))
		prompt = strings.ReplaceAll(prompt, "{{.SpecialistResults}}", strings.Join(specialistResults, "\n"))
		prompt = strings.ReplaceAll(prompt, "{{.CurrentState}}", currentState)
		
		messages := []*schema.Message{
			{Role: schema.System, Content: prompt},
		}
		
		// 调用反思模型
		response, err := e.config.ReflectionModel.Generate(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("reflection failed: %w", err)
		}
		
		// 更新状态
		err = compose.ProcessState[*EnhancedMultiAgentState](ctx, func(ctx context.Context, state *EnhancedMultiAgentState) error {
			state.ReflectionCount++
			
			// 触发反思回调
			if e.config.Callback != nil {
				if err := e.config.Callback.OnReflection(ctx, prompt, response.Content); err != nil {
					return fmt.Errorf("reflection callback failed: %w", err)
				}
			}
			
			// 决定下一阶段
			if state.ReflectionCount >= state.MaxReflections {
				state.CurrentPhase = phaseResultIntegration
			} else {
				// 可以继续反思或回到主控处理
				state.CurrentPhase = phaseHostProcessing
			}
			
			return nil
		})
		
		if err != nil {
			return nil, err
		}
		
		return response, nil
	})
	
	return e.graph.AddLambdaNode(nodeKeyReflection, reflectionLambda)
}

// addResultIntegrationNode 添加结果集成节点
func (e *EnhancedMultiAgent) addResultIntegrationNode() error {
	resultIntegrationLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		var finalResult strings.Builder
		
		// 从状态中获取所有结果
		err := compose.ProcessState[*EnhancedMultiAgentState](ctx, func(ctx context.Context, state *EnhancedMultiAgentState) error {
			finalResult.WriteString("=== 增强版多智能体系统执行结果 ===\n\n")
			
			// 添加任务信息
			if len(state.HostMessages) > 0 {
				finalResult.WriteString(fmt.Sprintf("原始任务：%s\n", state.HostMessages[0].Content))
			}
			finalResult.WriteString(fmt.Sprintf("任务复杂度：%s\n", state.TaskComplexity))
			finalResult.WriteString(fmt.Sprintf("执行计划：%s\n\n", state.ExecutionPlan))
			
			// 添加主控处理结果
			if len(state.HostMessages) > 1 {
				finalResult.WriteString("=== 主控Agent处理结果 ===\n")
				for i := 1; i < len(state.HostMessages); i++ {
					finalResult.WriteString(fmt.Sprintf("Step %d: %s\n", i, state.HostMessages[i].Content))
				}
				finalResult.WriteString("\n")
			}
			
			// 添加专家结果
			if len(state.SpecialistResults) > 0 {
				finalResult.WriteString("=== 专家处理结果 ===\n")
				for name, result := range state.SpecialistResults {
					finalResult.WriteString(fmt.Sprintf("%s: %s\n", name, result.Content))
				}
				finalResult.WriteString("\n")
			}
			
			// 添加执行统计
			finalResult.WriteString("=== 执行统计 ===\n")
			finalResult.WriteString(fmt.Sprintf("反思次数：%d/%d\n", state.ReflectionCount, state.MaxReflections))
			finalResult.WriteString(fmt.Sprintf("专家调用：%s\n", strings.Join(state.SpecialistCallHistory, ", ")))
			finalResult.WriteString(fmt.Sprintf("最终阶段：%s\n", state.CurrentPhase))
			
			state.CurrentPhase = phaseCompleted
			return nil
		})
		
		if err != nil {
			return nil, err
		}
		
		return &schema.Message{
			Role:    schema.Assistant,
			Content: finalResult.String(),
		}, nil
	})
	
	return e.graph.AddLambdaNode(nodeKeyResultIntegration, resultIntegrationLambda)
}

// addEdgesAndBranches 添加边和分支逻辑
func (e *EnhancedMultiAgent) addEdgesAndBranches() error {
	// 首先添加专家处理节点
	if err := e.addSpecialistProcessingNode(); err != nil {
		return fmt.Errorf("failed to add specialist processing node: %w", err)
	}
	
	// 起始边：START -> 复杂度评估
	if err := e.graph.AddEdge(compose.START, nodeKeyComplexityEvaluation); err != nil {
		return fmt.Errorf("failed to add start edge: %w", err)
	}
	
	// 复杂度评估 -> 任务规划
	if err := e.graph.AddEdge(nodeKeyComplexityEvaluation, nodeKeyTaskPlanning); err != nil {
		return fmt.Errorf("failed to add complexity to planning edge: %w", err)
	}
	
	// 任务规划 -> 主控React Agent
	if err := e.graph.AddEdge(nodeKeyTaskPlanning, nodeKeyHostReact); err != nil {
		return fmt.Errorf("failed to add planning to host edge: %w", err)
	}
	
	// 主控React Agent -> 专家路由
	if err := e.graph.AddEdge(nodeKeyHostReact, nodeKeySpecialistRouter); err != nil {
		return fmt.Errorf("failed to add host to router edge: %w", err)
	}
	
	// 专家路由分支
	specialistBranchCondition := func(ctx context.Context, input *schema.Message) (string, error) {
		var needsSpecialist bool
		err := compose.ProcessState[*EnhancedMultiAgentState](ctx, func(ctx context.Context, state *EnhancedMultiAgentState) error {
			needsSpecialist = state.NeedsSpecialistHelp
			return nil
		})
		if err != nil {
			return "", err
		}
		
		if needsSpecialist {
			return "specialist_processing", nil
		}
		return nodeKeyReflection, nil
	}
	
	specialistBranch := compose.NewGraphBranch(specialistBranchCondition, map[string]bool{
		"specialist_processing": true,
		nodeKeyReflection:       true,
	})
	
	if err := e.graph.AddBranch(nodeKeySpecialistRouter, specialistBranch); err != nil {
		return fmt.Errorf("failed to add specialist branch: %w", err)
	}
	
	// 添加专家处理的多分支
	if err := e.addSpecialistMultiBranch(); err != nil {
		return fmt.Errorf("failed to add specialist multi branch: %w", err)
	}
	
	// 反思分支
	reflectionBranchCondition := func(ctx context.Context, input *schema.Message) (string, error) {
		var shouldContinue bool
		err := compose.ProcessState[*EnhancedMultiAgentState](ctx, func(ctx context.Context, state *EnhancedMultiAgentState) error {
			// 如果还没达到最大反思次数，且任务还没完成，可以继续
			shouldContinue = state.ReflectionCount < state.MaxReflections && state.CurrentPhase != phaseCompleted
			return nil
		})
		if err != nil {
			return "", err
		}
		
		if shouldContinue {
			return nodeKeyHostReact, nil // 回到主控继续处理
		}
		return nodeKeyResultIntegration, nil // 进入结果集成
	}
	
	reflectionBranch := compose.NewGraphBranch(reflectionBranchCondition, map[string]bool{
		nodeKeyHostReact:         true,
		nodeKeyResultIntegration: true,
	})
	
	if err := e.graph.AddBranch(nodeKeyReflection, reflectionBranch); err != nil {
		return fmt.Errorf("failed to add reflection branch: %w", err)
	}
	
	// 结果集成 -> END
	if err := e.graph.AddEdge(nodeKeyResultIntegration, compose.END); err != nil {
		return fmt.Errorf("failed to add result to end edge: %w", err)
	}
	
	return nil
}

// addSpecialistMultiBranch 添加专家处理的多分支逻辑
// addSpecialistProcessingNode 添加专家处理节点
func (e *EnhancedMultiAgent) addSpecialistProcessingNode() error {
	// 创建一个虚拟的专家处理节点
	specialistProcessingLambda := compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		// 这个节点主要用于触发专家多分支
		return input, nil
	})
	
	return e.graph.AddLambdaNode("specialist_processing", specialistProcessingLambda)
}

func (e *EnhancedMultiAgent) addSpecialistMultiBranch() error {
	
	// 创建专家多分支条件
	specialistMultiBranchCondition := func(ctx context.Context, input *schema.Message) (map[string]bool, error) {
		selectedSpecialists := make(map[string]bool)
		
		err := compose.ProcessState[*EnhancedMultiAgentState](ctx, func(ctx context.Context, state *EnhancedMultiAgentState) error {
			// 根据专家调用历史决定调用哪些专家
			for _, specialistName := range state.SpecialistCallHistory {
				// 验证专家名称是否存在于配置中
				for _, specialist := range e.config.Specialists {
					if specialist.Name == specialistName {
						nodeKey := nodeKeySpecialistPrefix + specialistName
						selectedSpecialists[nodeKey] = true
						break
					}
				}
			}
			return nil
		})
		
		if err != nil {
			return nil, err
		}
		
		// 如果没有选择专家，直接进入反思
		if len(selectedSpecialists) == 0 {
			selectedSpecialists[nodeKeyReflection] = true
		}
		
		return selectedSpecialists, nil
	}
	
	// 构建所有可能的终点节点
	endNodes := make(map[string]bool)
	for _, specialist := range e.config.Specialists {
		nodeKey := nodeKeySpecialistPrefix + specialist.Name
		endNodes[nodeKey] = true
	}
	endNodes[nodeKeyReflection] = true
	
	specialistMultiBranch := compose.NewGraphMultiBranch(specialistMultiBranchCondition, endNodes)
	
	if err := e.graph.AddBranch("specialist_processing", specialistMultiBranch); err != nil {
		return fmt.Errorf("failed to add specialist multi branch: %w", err)
	}
	
	// 所有专家节点都连接到反思节点
	for _, specialist := range e.config.Specialists {
		nodeKey := nodeKeySpecialistPrefix + specialist.Name
		if err := e.graph.AddEdge(nodeKey, nodeKeyReflection); err != nil {
			return fmt.Errorf("failed to add specialist %s to reflection edge: %w", specialist.Name, err)
		}
	}
	
	return nil
}