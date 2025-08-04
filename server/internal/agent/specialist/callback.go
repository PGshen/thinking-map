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
)

// PlanningMultiAgentCallback 规划多智能体回调接口
type PlanningMultiAgentCallback interface {
	// OnPlanGenerated 当生成执行计划时调用
	OnPlanGenerated(ctx context.Context, info *PlanGeneratedInfo)
	
	// OnStepStarted 当步骤开始执行时调用
	OnStepStarted(ctx context.Context, info *StepStartedInfo)
	
	// OnStepCompleted 当步骤完成时调用
	OnStepCompleted(ctx context.Context, info *StepCompletedInfo)
	
	// OnIterationCompleted 当迭代完成时调用
	OnIterationCompleted(ctx context.Context, info *IterationCompletedInfo)
	
	// OnTaskCompleted 当整个任务完成时调用
	OnTaskCompleted(ctx context.Context, info *TaskCompletedInfo)
	
	// OnError 当发生错误时调用
	OnError(ctx context.Context, info *ErrorInfo)
}

// PlanGeneratedInfo 计划生成信息
type PlanGeneratedInfo struct {
	// OriginalQuery 原始查询
	OriginalQuery string
	// Plan 生成的执行计划
	Plan *ExecutionPlan
	// IterationCount 当前迭代次数
	IterationCount int
}

// StepStartedInfo 步骤开始信息
type StepStartedInfo struct {
	// Step 开始执行的步骤
	Step *ExecutionStep
	// SpecialistName 负责的专家名称
	SpecialistName string
	// IterationCount 当前迭代次数
	IterationCount int
}

// StepCompletedInfo 步骤完成信息
type StepCompletedInfo struct {
	// Step 完成的步骤
	Step *ExecutionStep
	// SpecialistName 负责的专家名称
	SpecialistName string
	// Output 步骤输出
	Output string
	// Duration 执行时长（毫秒）
	Duration int64
	// IterationCount 当前迭代次数
	IterationCount int
}

// IterationCompletedInfo 迭代完成信息
type IterationCompletedInfo struct {
	// IterationCount 完成的迭代次数
	IterationCount int
	// CompletedSteps 本次迭代完成的步骤数
	CompletedSteps int
	// TotalSteps 总步骤数
	TotalSteps int
	// IsTaskCompleted 任务是否已完成
	IsTaskCompleted bool
}

// TaskCompletedInfo 任务完成信息
type TaskCompletedInfo struct {
	// OriginalQuery 原始查询
	OriginalQuery string
	// FinalResult 最终结果
	FinalResult string
	// TotalIterations 总迭代次数
	TotalIterations int
	// TotalSteps 总步骤数
	TotalSteps int
	// Duration 总执行时长（毫秒）
	Duration int64
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	// Error 错误对象
	Error error
	// Context 错误上下文
	Context string
	// Step 相关步骤（如果有）
	Step *ExecutionStep
	// IterationCount 当前迭代次数
	IterationCount int
}

// DefaultCallback 默认回调实现（空实现）
type DefaultCallback struct{}

// OnPlanGenerated 实现PlanningMultiAgentCallback接口
func (d *DefaultCallback) OnPlanGenerated(ctx context.Context, info *PlanGeneratedInfo) {
	// 默认空实现
}

// OnStepStarted 实现PlanningMultiAgentCallback接口
func (d *DefaultCallback) OnStepStarted(ctx context.Context, info *StepStartedInfo) {
	// 默认空实现
}

// OnStepCompleted 实现PlanningMultiAgentCallback接口
func (d *DefaultCallback) OnStepCompleted(ctx context.Context, info *StepCompletedInfo) {
	// 默认空实现
}

// OnIterationCompleted 实现PlanningMultiAgentCallback接口
func (d *DefaultCallback) OnIterationCompleted(ctx context.Context, info *IterationCompletedInfo) {
	// 默认空实现
}

// OnTaskCompleted 实现PlanningMultiAgentCallback接口
func (d *DefaultCallback) OnTaskCompleted(ctx context.Context, info *TaskCompletedInfo) {
	// 默认空实现
}

// OnError 实现PlanningMultiAgentCallback接口
func (d *DefaultCallback) OnError(ctx context.Context, info *ErrorInfo) {
	// 默认空实现
}

// LoggingCallback 日志回调实现
type LoggingCallback struct {
	// Logger 日志记录器（这里简化为打印到标准输出）
}

// OnPlanGenerated 实现PlanningMultiAgentCallback接口
func (l *LoggingCallback) OnPlanGenerated(ctx context.Context, info *PlanGeneratedInfo) {
	// TODO: 实现日志记录
}

// OnStepStarted 实现PlanningMultiAgentCallback接口
func (l *LoggingCallback) OnStepStarted(ctx context.Context, info *StepStartedInfo) {
	// TODO: 实现日志记录
}

// OnStepCompleted 实现PlanningMultiAgentCallback接口
func (l *LoggingCallback) OnStepCompleted(ctx context.Context, info *StepCompletedInfo) {
	// TODO: 实现日志记录
}

// OnIterationCompleted 实现PlanningMultiAgentCallback接口
func (l *LoggingCallback) OnIterationCompleted(ctx context.Context, info *IterationCompletedInfo) {
	// TODO: 实现日志记录
}

// OnTaskCompleted 实现PlanningMultiAgentCallback接口
func (l *LoggingCallback) OnTaskCompleted(ctx context.Context, info *TaskCompletedInfo) {
	// TODO: 实现日志记录
}

// OnError 实现PlanningMultiAgentCallback接口
func (l *LoggingCallback) OnError(ctx context.Context, info *ErrorInfo) {
	// TODO: 实现日志记录
}