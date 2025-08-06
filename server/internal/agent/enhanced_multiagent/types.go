package enhanced_multiagent

import (
	"context"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/components/model"
)

// 枚举类型定义

// TaskComplexity 任务复杂度
type TaskComplexity int

const (
	TaskComplexitySimple   TaskComplexity = iota // 简单任务
	TaskComplexityModerate                        // 中等复杂度
	TaskComplexityComplex                         // 复杂任务
)

// ActionType 行动类型
type ActionType int

const (
	ActionTypeDirectAnswer ActionType = iota // 直接回答
	ActionTypeCreatePlan                     // 创建规划
	ActionTypeExecuteStep                    // 执行步骤
	ActionTypeReflect                        // 反思
	ActionTypeUpdatePlan                     // 更新规划
)

// StepStatus 步骤状态
type StepStatus int

const (
	StepStatusPending   StepStatus = iota // 待执行
	StepStatusExecuting                   // 执行中
	StepStatusCompleted                   // 已完成
	StepStatusFailed                      // 失败
	StepStatusSkipped                     // 跳过
)

// ExecutionStatus 执行状态
type ExecutionStatus int

const (
	ExecutionStatusSuccess ExecutionStatus = iota // 成功
	ExecutionStatusFailed                         // 失败
	ExecutionStatusPartial                        // 部分成功
)

// PlanUpdateType 规划更新类型
type PlanUpdateType int

const (
	PlanUpdateTypeAddStep     PlanUpdateType = iota // 添加步骤
	PlanUpdateTypeModifyStep                        // 修改步骤
	PlanUpdateTypeRemoveStep                        // 删除步骤
	PlanUpdateTypeReorderSteps                      // 重排步骤
	PlanUpdateTypeModifyPlan                        // 修改规划
)

// 核心状态类型定义

// EnhancedState 增强版多智能体系统的全局状态
type EnhancedState struct {
	// 原始输入消息
	OriginalMessages []*schema.Message

	// 当前任务规划（Markdown格式）
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

	// 执行历史
	ExecutionHistory []*ExecutionRecord

	// 思考历史
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
}

// ExecutionRecord 执行记录
type ExecutionRecord struct {
	Round     int
	Results   *CollectedResults
	Timestamp time.Time
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
	UpdateType  PlanUpdateType // add_step, modify_step, remove_step, reorder_steps
	Description string
	Timestamp   time.Time
	Changes     map[string]interface{}
}

// PlanStep 规划步骤
type PlanStep struct {
	ID                int
	Description       string
	Status            StepStatus        // pending, executing, completed, failed, skipped
	AssignedTo        string            // 分配给哪个specialist
	Result            string            // 执行结果
	Feedback          string            // 反馈信息
	Priority          int               // 步骤优先级
	Dependencies      []int             // 依赖的步骤ID
	EstimatedDuration time.Duration     // 预估执行时间
	ActualDuration    time.Duration     // 实际执行时间
	RetryCount        int               // 重试次数
	MaxRetries        int               // 最大重试次数
	CreatedAt         time.Time         // 步骤创建时间
	UpdatedAt         time.Time         // 步骤更新时间
}

// ThinkingResult 思考结果
type ThinkingResult struct {
	// 思考内容
	Thought string

	// 任务复杂度评估
	Complexity TaskComplexity // simple, moderate, complex

	// 推理过程
	Reasoning string

	// 下一步行动
	NextAction ActionType // direct_answer, create_plan, execute_step, reflect

	// 原始消息
	OriginalMessages []*schema.Message

	// 时间戳
	Timestamp time.Time
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
	Status ExecutionStatus // success, failed, partial

	// 错误信息
	Error string

	// 执行时长
	Duration time.Duration

	// 置信度
	Confidence float64
}

// CollectedResults 收集的结果
type CollectedResults struct {
	// 所有专家结果
	Results map[string]*SpecialistResult

	// 成功的结果
	SuccessfulResults []*SpecialistResult

	// 失败的结果
	FailedResults []*SpecialistResult

	// 汇总信息
	Summary string
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
}

// 配置类型定义

// EnhancedMultiAgentConfig 增强版多智能体配置
type EnhancedMultiAgentConfig struct {
	// 主控Agent配置
	Host EnhancedHost

	// 专家Agent列表
	Specialists []*EnhancedSpecialist

	// 系统名称
	Name string

	// 最大执行轮次
	MaxRounds int

	// 复杂度判断阈值
	ComplexityThreshold float64

	// 规划模板
	PlanTemplate string

	// 思考提示模板
	ThinkingPromptTemplate string

	// 反思提示模板
	ReflectionPromptTemplate string

	// 流式工具调用检查器
	StreamToolCallChecker func(ctx context.Context, modelOutput *schema.StreamReader[*schema.Message]) (bool, error)

	// 回调处理器
	Callbacks []EnhancedMultiAgentCallback
}

// EnhancedHost 增强版主控Agent
type EnhancedHost struct {
	// 工具调用模型
	ToolCallingModel model.ToolCallingChatModel

	// 思考模型（可以与工具调用模型相同）
	ThinkingModel model.BaseChatModel

	// 系统提示
	SystemPrompt string

	// 思考提示模板
	ThinkingPrompt string

	// 规划提示模板
	PlanningPrompt string

	// 反思提示模板
	ReflectionPrompt string
}

// AgentMeta Agent元信息
type AgentMeta struct {
	Name        string
	Description string
	Version     string
}

// EnhancedSpecialist 增强版专家Agent
type EnhancedSpecialist struct {
	// 基础元信息
	AgentMeta

	// 聊天模型
	ChatModel model.BaseChatModel

	// 系统提示
	SystemPrompt string

	// 可调用组件
	Invokable compose.Invoke[[]*schema.Message, *schema.Message, any]

	// 流式组件
	Streamable compose.Stream[[]*schema.Message, *schema.Message, any]

	// 专家能力描述
	Capabilities []string

	// 输入预处理器
	InputProcessor func(ctx context.Context, input *ExecutionContext) ([]*schema.Message, error)

	// 输出后处理器
	OutputProcessor func(ctx context.Context, output *schema.Message) (*SpecialistResult, error)
}

// 回调接口定义

// EnhancedMultiAgentCallback 增强版多智能体回调接口
type EnhancedMultiAgentCallback interface {
	// 思考阶段回调
	OnThinking(ctx context.Context, state *EnhancedState, thinking *ThinkingResult) error

	// 规划阶段回调
	OnPlanning(ctx context.Context, state *EnhancedState, plan *TaskPlan) error

	// 专家调用回调
	OnSpecialistCall(ctx context.Context, state *EnhancedState, specialist string, result *SpecialistResult) error

	// 反馈阶段回调
	OnFeedback(ctx context.Context, state *EnhancedState, feedback *FeedbackResult) error

	// 任务完成回调
	OnTaskComplete(ctx context.Context, state *EnhancedState, finalAnswer *schema.Message) error
}

// StateCheckpoint 状态检查点
type StateCheckpoint struct {
	Timestamp time.Time
	Round     int
	State     *EnhancedState
}

// StateSerializer 状态序列化接口
type StateSerializer interface {
	Serialize(state *EnhancedState) ([]byte, error)
	Deserialize(data []byte) (*EnhancedState, error)
}