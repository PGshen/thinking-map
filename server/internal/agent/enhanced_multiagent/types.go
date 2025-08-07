package enhanced_multiagent

import (
	"time"

	"github.com/cloudwego/eino/schema"
)

// EnhancedState 增强版多智能体系统的全局状态
type EnhancedState struct {
	// 原始对话历史
	OriginalMessages []*schema.Message

	// 对话上下文分析结果
	ConversationContext *ConversationContext

	// 当前任务规划
	CurrentPlan *TaskPlan

	// 当前执行上下文
	CurrentExecution *ExecutionContext

	// 当前专家结果映射
	CurrentSpecialistResults map[string]*SpecialistResult

	// 当前收集的结果
	CurrentCollectedResults *CollectedResults

	// 当前反馈结果
	CurrentFeedbackResult *FeedbackResult

	// 当前思考结果
	CurrentThinkingResult *ThinkingResult

	// 执行历史（仅限当前调用）
	ExecutionHistory []*ExecutionRecord

	// 思考历史（仅限当前调用）
	ThinkingHistory []*ThinkingResult

	// 当前执行轮次
	CurrentRound int

	// 最大执行轮次
	MaxRounds int

	// 是否为简单任务
	IsSimpleTask bool

	// 任务完成状态
	IsCompleted bool

	// 最终答案
	FinalAnswer *schema.Message

	// 会话ID（用于日志追踪）
	SessionID string

	// 调用时间戳
	CallTimestamp time.Time
}

// ConversationContext 对话上下文分析结果
type ConversationContext struct {
	// 对话轮次数
	TurnCount int

	// 是否为首次对话
	IsFirstTurn bool

	// 是否为延续性问题
	IsContinuation bool

	// 最新用户消息
	LatestUserMessage *schema.Message

	// 最新助手回复
	LatestAssistantMessage *schema.Message

	// 对话主题
	ConversationTopic string

	// 用户意图分析
	UserIntent string

	// 意图置信度
	IntentConfidence float64

	// 情感色调分析
	EmotionalTone string

	// 关键实体提取
	KeyEntities []string

	// 复杂度提示
	ComplexityHint TaskComplexity

	// 相关历史上下文
	RelevantHistory []*schema.Message

	// 上下文摘要
	ContextSummary string

	// 分析时间戳
	AnalyzedAt time.Time

	// 扩展元数据
	Metadata map[string]interface{}
}

// ExecutionRecord 执行记录
type ExecutionRecord struct {
	Round       int
	PlanVersion int
	Results     *CollectedResults
	Feedback    *FeedbackResult
	Duration    time.Duration
	Timestamp   time.Time
	Status      ExecutionStatus
	Metadata    map[string]interface{}
}

// TaskPlan 任务规划结构
type TaskPlan struct {
	// 规划内容（Markdown格式）
	Content string

	// 当前步骤
	CurrentStep int

	// 总步骤数（动态更新）
	TotalSteps int

	// 已完成的步骤
	CompletedSteps []int

	// 步骤详情（支持动态添加、修改、删除）
	Steps []*PlanStep

	// 规划版本号（每次更新递增）
	Version int

	// 是否允许动态调整
	AllowDynamicUpdate bool

	// 创建时间
	CreatedAt time.Time

	// 最后更新时间
	UpdatedAt time.Time

	// 更新历史
	UpdateHistory []*PlanUpdate
}

// PlanUpdate 规划更新记录
type PlanUpdate struct {
	Version     int
	UpdateType  PlanUpdateType
	Description string
	Timestamp   time.Time
	Changes     map[string]interface{}
}

// PlanStep 规划步骤
type PlanStep struct {
	ID                int
	Description       string
	Status            StepStatus
	AssignedTo        string // 分配给哪个specialist
	Result            string // 执行结果
	Feedback          string // 反馈信息
	Priority          int    // 步骤优先级
	Dependencies      []int  // 依赖的步骤ID
	EstimatedDuration time.Duration
	ActualDuration    time.Duration
	RetryCount        int
	MaxRetries        int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ThinkingResult 思考结果
type ThinkingResult struct {
	// 思考内容
	Thought string

	// 任务复杂度评估
	Complexity TaskComplexity

	// 推理过程
	Reasoning string

	// 下一步行动
	NextAction ActionType

	// 原始消息
	OriginalMessages []*schema.Message

	// 对话上下文分析
	ConversationAnalysis *ConversationContext

	// 是否需要参考历史对话
	NeedsHistoryContext bool

	// 回答策略
	ResponseStrategy string

	// 关键洞察
	KeyInsights []string

	// 风险评估
	RiskAssessment string

	// 置信度评估
	Confidence float64

	// 思考耗时
	Duration time.Duration

	// 时间戳
	Timestamp time.Time

	// 扩展元数据
	Metadata map[string]interface{}
}

// ExecutionContext 执行上下文
type ExecutionContext struct {
	// 当前任务描述
	TaskDescription string

	// 相关的规划步骤
	PlanStep *PlanStep

	// 历史消息
	Messages []*schema.Message

	// 执行参数
	Parameters map[string]interface{}

	// 期望输出格式
	ExpectedFormat string
}

// SpecialistResult 专家执行结果
type SpecialistResult struct {
	// 专家名称
	SpecialistName string

	// 执行结果
	Result *schema.Message

	// 执行状态
	Status ExecutionStatus

	// 错误信息
	Error string

	// 执行时长
	Duration time.Duration

	// 置信度
	Confidence float64

	// 开始时间
	StartTime time.Time

	// 结束时间
	EndTime time.Time

	// 重试次数
	RetryCount int

	// 质量评分
	QualityScore float64

	// 相关性评分
	RelevanceScore float64

	// 扩展元数据
	Metadata map[string]interface{}
}

// CollectedResults 收集的结果
type CollectedResults struct {
	// 所有专家结果
	Results map[string]*SpecialistResult

	// 成功的结果
	SuccessfulResults []*SpecialistResult

	// 失败的结果
	FailedResults []*SpecialistResult

	// 部分成功的结果
	PartialResults []*SpecialistResult

	// 汇总信息
	Summary string

	// 整体质量评分
	OverallQuality float64

	// 结果一致性评分
	ConsistencyScore float64

	// 收集时间
	CollectedAt time.Time

	// 总执行时长
	TotalDuration time.Duration

	// 扩展元数据
	Metadata map[string]interface{}
}

// FeedbackResult 反馈结果
type FeedbackResult struct {
	// 反馈内容
	Feedback string

	// 是否需要继续执行
	ShouldContinue bool

	// 建议的下一步行动
	SuggestedAction ActionType

	// 规划更新建议
	PlanUpdateSuggestion string

	// 最终答案（如果任务完成）
	FinalAnswer *schema.Message

	// 收集的结果
	CollectedResults *CollectedResults

	// 对话连贯性评估
	ConversationCoherence float64

	// 用户满意度预测
	UserSatisfactionPrediction float64

	// 改进建议
	ImprovementSuggestions []string

	// 风险警告
	RiskWarnings []string

	// 反馈生成时间
	GeneratedAt time.Time

	// 扩展元数据
	Metadata map[string]interface{}
}

// 枚举类型定义
type TaskComplexity int

const (
	TaskComplexitySimple TaskComplexity = iota
	TaskComplexityModerate
	TaskComplexityComplex
)

func (tc TaskComplexity) String() string {
	switch tc {
	case TaskComplexitySimple:
		return "simple"
	case TaskComplexityModerate:
		return "moderate"
	case TaskComplexityComplex:
		return "complex"
	default:
		return "unknown"
	}
}

type ActionType int

const (
	ActionTypeDirectAnswer ActionType = iota
	ActionTypeCreatePlan
	ActionTypeExecuteStep
	ActionTypeReflect
	ActionTypeUpdatePlan
)

func (at ActionType) String() string {
	switch at {
	case ActionTypeDirectAnswer:
		return "direct_answer"
	case ActionTypeCreatePlan:
		return "create_plan"
	case ActionTypeExecuteStep:
		return "execute_step"
	case ActionTypeReflect:
		return "reflect"
	case ActionTypeUpdatePlan:
		return "update_plan"
	default:
		return "unknown"
	}
}

type StepStatus int

const (
	StepStatusPending StepStatus = iota
	StepStatusExecuting
	StepStatusCompleted
	StepStatusFailed
	StepStatusSkipped
	StepStatusBlocked
)

func (ss StepStatus) String() string {
	switch ss {
	case StepStatusPending:
		return "pending"
	case StepStatusExecuting:
		return "executing"
	case StepStatusCompleted:
		return "completed"
	case StepStatusFailed:
		return "failed"
	case StepStatusSkipped:
		return "skipped"
	case StepStatusBlocked:
		return "blocked"
	default:
		return "unknown"
	}
}

type ExecutionStatus int

const (
	ExecutionStatusRunning ExecutionStatus = iota
	ExecutionStatusCompleted
	ExecutionStatusFailed
	ExecutionStatusTimeout
	ExecutionStatusCancelled
	ExecutionStatusSuccess
	ExecutionStatusPartial
)

func (es ExecutionStatus) String() string {
	switch es {
	case ExecutionStatusRunning:
		return "running"
	case ExecutionStatusCompleted:
		return "completed"
	case ExecutionStatusFailed:
		return "failed"
	case ExecutionStatusTimeout:
		return "timeout"
	case ExecutionStatusCancelled:
		return "cancelled"
	case ExecutionStatusSuccess:
		return "success"
	case ExecutionStatusPartial:
		return "partial"
	default:
		return "unknown"
	}
}

type PlanUpdateType int

const (
	PlanUpdateTypeAddStep PlanUpdateType = iota
	PlanUpdateTypeModifyStep
	PlanUpdateTypeDeleteStep
	PlanUpdateTypeReorderSteps
	PlanUpdateTypeChangeStrategy
	PlanUpdateTypeRemoveStep
	PlanUpdateTypeModifyPlan
)

func (put PlanUpdateType) String() string {
	switch put {
	case PlanUpdateTypeAddStep:
		return "add_step"
	case PlanUpdateTypeModifyStep:
		return "modify_step"
	case PlanUpdateTypeDeleteStep:
		return "delete_step"
	case PlanUpdateTypeReorderSteps:
		return "reorder_steps"
	case PlanUpdateTypeChangeStrategy:
		return "change_strategy"
	case PlanUpdateTypeRemoveStep:
		return "remove_step"
	case PlanUpdateTypeModifyPlan:
		return "modify_plan"
	default:
		return "unknown"
	}
}
