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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// parseStateFromMessage 从消息中解析规划状态
func parseStateFromMessage(msg *schema.Message) (*PlanningState, error) {
	if msg.Content == "" {
		return nil, fmt.Errorf("empty message content")
	}
	
	// 尝试从消息内容中解析JSON状态
	var state PlanningState
	if err := json.Unmarshal([]byte(msg.Content), &state); err != nil {
		// 如果解析失败，可能是普通文本消息，返回错误让调用者处理
		return nil, fmt.Errorf("failed to parse state from message: %w", err)
	}
	
	return &state, nil
}

// encodeStateToMessage 将规划状态编码为消息
func encodeStateToMessage(state *PlanningState) *schema.Message {
	stateJSON, err := json.Marshal(state)
	if err != nil {
		// 如果编码失败，返回错误消息
		return schema.AssistantMessage(fmt.Sprintf("Error encoding state: %v", err), nil)
	}
	
	return schema.AssistantMessage(string(stateJSON), nil)
}

// buildPlanningPrompt 构建规划提示词
func buildPlanningPrompt(query string, plan *ExecutionPlan, iterationCount int) string {
	// 构建基础提示词
	prompt := fmt.Sprintf("请为以下查询制定执行计划：%s\n\n", query)
	
	// 如果有现有计划，添加迭代信息
	if plan != nil && iterationCount > 0 {
		prompt += fmt.Sprintf("当前是第 %d 次迭代，现有计划：\n", iterationCount)
		for _, step := range plan.Steps {
			prompt += fmt.Sprintf("- %s: %s (状态: %s)\n", step.ID, step.Description, step.Status)
		}
		prompt += "\n请优化或调整计划：\n"
	}
	
	prompt += "请以JSON格式返回执行计划，包含steps数组，每个步骤包含id、description、specialist_name、input、dependencies字段。"
	
	return prompt
}

// buildPlanningPromptWithTemplate 使用模板构建规划提示词
func buildPlanningPromptWithTemplate(template, query string) string {
	// 简单的模板替换
	prompt := strings.ReplaceAll(template, "{{query}}", query)
	
	// TODO: 添加专家信息替换
	specialistsInfo := "专家信息将在后续版本中添加"
	prompt = strings.ReplaceAll(prompt, "{{specialists_info}}", specialistsInfo)
	
	return prompt
}

// buildSpecialistPrompt 构建专家提示词
func buildSpecialistPrompt(step *ExecutionStep, originalQuery, specialistDescription string) string {
	prompt := fmt.Sprintf("%s\n\n原始查询：%s\n\n当前任务：%s\n\n输入要求：%s", 
		specialistDescription, originalQuery, step.Description, step.Input)
	return prompt
}

// buildSpecialistPromptWithSystem 使用系统提示词构建专家提示词
func buildSpecialistPromptWithSystem(systemPrompt string, step *ExecutionStep) string {
	prompt := fmt.Sprintf("%s\n\n当前任务：%s\n\n输入要求：%s", 
		systemPrompt, step.Description, step.Input)
	return prompt
}

// buildSummaryPrompt 构建汇总提示词
func buildSummaryPrompt(originalQuery string, plan *ExecutionPlan) string {
	executionResults := formatExecutionResults(plan)
	prompt := fmt.Sprintf("原始查询：%s\n\n执行结果：\n%s\n\n请对以上执行结果进行汇总和总结。", 
		originalQuery, executionResults)
	return prompt
}

// buildSummaryPromptWithTemplate 使用模板构建汇总提示词
func buildSummaryPromptWithTemplate(template, executionResults string) string {
	return strings.ReplaceAll(template, "{{execution_results}}", executionResults)
}

// parsePlanFromResponse 从响应中解析执行计划
func parsePlanFromResponse(content string) (*ExecutionPlan, error) {
	// 尝试从响应中提取JSON
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	
	if jsonStart == -1 || jsonEnd == -1 || jsonStart >= jsonEnd {
		return nil, fmt.Errorf("no valid JSON found in response")
	}
	
	jsonContent := content[jsonStart : jsonEnd+1]
	
	// 解析JSON为临时结构
	var planData struct {
		Steps []struct {
			ID             string   `json:"id"`
			Description    string   `json:"description"`
			SpecialistName string   `json:"specialist_name"`
			Input          string   `json:"input"`
			Dependencies   []string `json:"dependencies"`
		} `json:"steps"`
	}
	
	if err := json.Unmarshal([]byte(jsonContent), &planData); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
	}
	
	// 转换为ExecutionPlan
	plan := &ExecutionPlan{
		Steps:            make([]*ExecutionStep, len(planData.Steps)),
		CurrentStepIndex: 0,
		IsCompleted:      false,
	}
	
	for i, stepData := range planData.Steps {
		plan.Steps[i] = &ExecutionStep{
			ID:             stepData.ID,
			Description:    stepData.Description,
			SpecialistName: stepData.SpecialistName,
			Status:         StepStatusPending,
			Input:          stepData.Input,
			Dependencies:   stepData.Dependencies,
		}
	}
	
	return plan, nil
}

// generateStepID 生成步骤ID
func generateStepID(index int) string {
	return fmt.Sprintf("step_%d", index+1)
}

// validatePlan 验证执行计划
func validatePlan(plan *ExecutionPlan) error {
	if len(plan.Steps) == 0 {
		return fmt.Errorf("plan must have at least one step")
	}
	
	// 检查步骤ID唯一性
	stepIDs := make(map[string]bool)
	for _, step := range plan.Steps {
		if step.ID == "" {
			return fmt.Errorf("step ID cannot be empty")
		}
		if stepIDs[step.ID] {
			return fmt.Errorf("duplicate step ID: %s", step.ID)
		}
		stepIDs[step.ID] = true
	}
	
	// 检查依赖关系
	for _, step := range plan.Steps {
		for _, depID := range step.Dependencies {
			if !stepIDs[depID] {
				return fmt.Errorf("step %s depends on non-existent step %s", step.ID, depID)
			}
		}
	}
	
	return nil
}

// formatExecutionResults 格式化执行结果
func formatExecutionResults(plan *ExecutionPlan) string {
	var results []string
	
	for _, step := range plan.Steps {
		if step.Status == StepStatusCompleted && step.Output != "" {
			result := fmt.Sprintf("步骤 %s (%s):\n输入: %s\n输出: %s", 
				step.ID, step.Description, step.Input, step.Output)
			results = append(results, result)
		}
	}
	
	return strings.Join(results, "\n\n---\n\n")
}

// isAllStepsCompleted 检查是否所有步骤都已完成
func isAllStepsCompleted(plan *ExecutionPlan) bool {
	for _, step := range plan.Steps {
		if step.Status != StepStatusCompleted {
			return false
		}
	}
	return true
}

// getCompletedStepsCount 获取已完成步骤数量
func getCompletedStepsCount(plan *ExecutionPlan) int {
	count := 0
	for _, step := range plan.Steps {
		if step.Status == StepStatusCompleted {
			count++
		}
	}
	return count
}

// getPendingStepsCount 获取待执行步骤数量
func getPendingStepsCount(plan *ExecutionPlan) int {
	count := 0
	for _, step := range plan.Steps {
		if step.Status == StepStatusPending {
			count++
		}
	}
	return count
}

// isSpecialistSuitable 检查专家是否适合执行步骤
func isSpecialistSuitable(specialist *Specialist, step *ExecutionStep) bool {
	// 简单的名称匹配检查
	return specialist.AgentMeta.Name == step.SpecialistName
}