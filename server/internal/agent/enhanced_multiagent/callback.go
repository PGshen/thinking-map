package enhanced_multiagent

import (
	"context"
	"time"

	"github.com/cloudwego/eino/schema"
)

// EnhancedMultiAgentCallback 增强版多智能体回调接口
type EnhancedMultiAgentCallback interface {
	// 思考开始
	OnThinkingStart(ctx context.Context, info *ThinkingStartInfo) context.Context

	// 思考完成
	OnThinkingComplete(ctx context.Context, info *ThinkingCompleteInfo) context.Context

	// 规划创建
	OnPlanCreated(ctx context.Context, info *PlanCreatedInfo) context.Context

	// 规划更新
	OnPlanUpdated(ctx context.Context, info *PlanUpdatedInfo) context.Context

	// 专家调用
	OnSpecialistInvoke(ctx context.Context, info *SpecialistInvokeInfo) context.Context

	// 专家完成
	OnSpecialistComplete(ctx context.Context, info *SpecialistCompleteInfo) context.Context

	// 反馈生成
	OnFeedbackGenerated(ctx context.Context, info *FeedbackGeneratedInfo) context.Context

	// 任务完成
	OnTaskComplete(ctx context.Context, info *TaskCompleteInfo) context.Context

	// 错误处理
	OnError(ctx context.Context, info *ErrorInfo) context.Context
}

// 回调信息结构体

// ThinkingStartInfo 思考开始信息
type ThinkingStartInfo struct {
	SessionID       string
	Messages        []*schema.Message
	ConversationCtx *ConversationContext
	Timestamp       time.Time
	Metadata        map[string]interface{}
}

// ThinkingCompleteInfo 思考完成信息
type ThinkingCompleteInfo struct {
	SessionID      string
	ThinkingResult *ThinkingResult
	Duration       time.Duration
	Timestamp      time.Time
	Metadata       map[string]interface{}
}

// PlanCreatedInfo 规划创建信息
type PlanCreatedInfo struct {
	SessionID      string
	Plan           *TaskPlan
	ThinkingResult *ThinkingResult
	Timestamp      time.Time
	Metadata       map[string]interface{}
}

// PlanUpdatedInfo 规划更新信息
type PlanUpdatedInfo struct {
	SessionID  string
	OldPlan    *TaskPlan
	NewPlan    *TaskPlan
	UpdateType PlanUpdateType
	Reason     string
	Timestamp  time.Time
	Metadata   map[string]interface{}
}

// SpecialistInvokeInfo 专家调用信息
type SpecialistInvokeInfo struct {
	SessionID      string
	SpecialistName string
	PlanStep       *PlanStep
	ExecutionCtx   *ExecutionContext
	Timestamp      time.Time
	Metadata       map[string]interface{}
}

// SpecialistCompleteInfo 专家完成信息
type SpecialistCompleteInfo struct {
	SessionID      string
	SpecialistName string
	Result         *SpecialistResult
	Duration       time.Duration
	Timestamp      time.Time
	Metadata       map[string]interface{}
}

// FeedbackGeneratedInfo 反馈生成信息
type FeedbackGeneratedInfo struct {
	SessionID        string
	FeedbackResult   *FeedbackResult
	CollectedResults *CollectedResults
	Round            int
	Timestamp        time.Time
	Metadata         map[string]interface{}
}

// TaskCompleteInfo 任务完成信息
type TaskCompleteInfo struct {
	SessionID        string
	FinalAnswer      *schema.Message
	ExecutionHistory []*ExecutionRecord
	TotalDuration    time.Duration
	TotalRounds      int
	IsSimpleTask     bool
	Timestamp        time.Time
	Metadata         map[string]interface{}
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	SessionID string
	Error     error
	ErrorType string
	Component string
	Context   map[string]interface{}
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// BaseCallback 基础回调实现，提供默认的空实现
type BaseCallback struct{}

func (bc *BaseCallback) OnThinkingStart(ctx context.Context, info *ThinkingStartInfo) context.Context {
	return ctx
}

func (bc *BaseCallback) OnThinkingComplete(ctx context.Context, info *ThinkingCompleteInfo) context.Context {
	return ctx
}

func (bc *BaseCallback) OnPlanCreated(ctx context.Context, info *PlanCreatedInfo) context.Context {
	return ctx
}

func (bc *BaseCallback) OnPlanUpdated(ctx context.Context, info *PlanUpdatedInfo) context.Context {
	return ctx
}

func (bc *BaseCallback) OnSpecialistInvoke(ctx context.Context, info *SpecialistInvokeInfo) context.Context {
	return ctx
}

func (bc *BaseCallback) OnSpecialistComplete(ctx context.Context, info *SpecialistCompleteInfo) context.Context {
	return ctx
}

func (bc *BaseCallback) OnFeedbackGenerated(ctx context.Context, info *FeedbackGeneratedInfo) context.Context {
	return ctx
}

func (bc *BaseCallback) OnTaskComplete(ctx context.Context, info *TaskCompleteInfo) context.Context {
	return ctx
}

func (bc *BaseCallback) OnError(ctx context.Context, info *ErrorInfo) context.Context {
	return ctx
}

// 确保BaseCallback实现了EnhancedMultiAgentCallback接口
var _ EnhancedMultiAgentCallback = (*BaseCallback)(nil)
