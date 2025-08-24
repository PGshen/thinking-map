# Eino Enhanced MultiAgent 核心类型定义

## 核心状态类型

### EnhancedState - 全局状态

```go
// EnhancedState 增强版多智能体系统的全局状态
type EnhancedState struct {
    // 基础信息
    SessionID           string                 `json:"session_id"`           // 会话唯一标识
    StartTime          time.Time              `json:"start_time"`           // 开始时间
    LastUpdateTime     time.Time              `json:"last_update_time"`    // 最后更新时间
    
    // 对话上下文 (新增)
    ConversationContext *ConversationContext   `json:"conversation_context"` // 对话上下文分析结果
    
    // 原始输入
    OriginalMessages   []*schema.Message      `json:"original_messages"`   // 原始对话历史
    
    // 思考阶段
    CurrentThinkingResult *ThinkingResult     `json:"current_thinking_result"` // 当前思考结果
    ThinkingHistory      []*ThinkingResult    `json:"thinking_history"`       // 思考历史
    
    // 任务规划
    CurrentPlan         *TaskPlan             `json:"current_plan"`         // 当前任务规划
    PlanHistory         []*TaskPlan           `json:"plan_history"`         // 规划历史
    
    // 执行状态
    CurrentExecution    *ExecutionContext     `json:"current_execution"`    // 当前执行上下文
    ExecutionHistory    []*ExecutionRecord    `json:"execution_history"`    // 执行历史记录
    
    // 专家结果
    CurrentSpecialistResults map[string]*SpecialistResult `json:"current_specialist_results"` // 当前专家执行结果
    
    // 结果收集
    CurrentCollectedResults *CollectedResults  `json:"current_collected_results"` // 当前收集的结果
    
    // 反馈处理
    CurrentFeedbackResult *FeedbackResult     `json:"current_feedback_result"` // 当前反馈结果
    
    // 执行控制
    CurrentRound        int                   `json:"current_round"`        // 当前执行轮次
    MaxRounds          int                   `json:"max_rounds"`          // 最大执行轮次
    IsCompleted        bool                  `json:"is_completed"`        // 是否已完成
    IsSimpleTask       bool                  `json:"is_simple_task"`      // 是否为简单任务
    
    // 最终结果
    FinalAnswer        *schema.Message       `json:"final_answer"`        // 最终答案
    
    // 元数据
    Metadata           map[string]interface{} `json:"metadata"`            // 扩展元数据
    
    // 并发控制
    mu                 sync.RWMutex          `json:"-"`                   // 读写锁
}
```

### ConversationContext - 对话上下文

```go
// ConversationContext 对话上下文分析结果
type ConversationContext struct {
    // 对话基本信息
    TotalTurns       int                    `json:"total_turns"`        // 总对话轮次
    CurrentTurn      int                    `json:"current_turn"`       // 当前轮次
    IsFirstTurn      bool                   `json:"is_first_turn"`      // 是否首次对话
    IsContinuation   bool                   `json:"is_continuation"`    // 是否延续对话
    
    // 用户意图分析
    UserIntent       string                 `json:"userIntent"`        // 用户意图
    IntentConfidence float64               `json:"intent_confidence"`  // 意图置信度
    IntentCategory   string                 `json:"intent_category"`    // 意图分类
    
    // 上下文关联
    RelevantHistory  []*schema.Message      `json:"relevantHistory"`   // 相关历史消息
    ContextSummary   string                 `json:"contextSummary"`    // 上下文摘要
    KeyTopics        []string               `json:"keyTopics"`         // 关键话题
    
    // 对话状态
    RequiresClarification bool              `json:"requires_clarification"` // 是否需要澄清
    HasUnresolvedIssues  bool               `json:"has_unresolved_issues"` // 是否有未解决问题
    
    // 元数据
    AnalyzedAt       time.Time              `json:"analyzed_at"`        // 分析时间
    Metadata         map[string]interface{} `json:"metadata"`           // 扩展信息
}
```

### ExecutionRecord - 执行记录

```go
// ExecutionRecord 执行记录
type ExecutionRecord struct {
    ID              string                 `json:"id"`                // 记录唯一标识
    Round           int                    `json:"round"`             // 执行轮次
    PlanVersion     int                    `json:"planVersion"`      // 规划版本
    StepID          string                 `json:"stepID"`           // 步骤ID
    StepName        string                 `json:"step_name"`         // 步骤名称
    AssignedTo      string                 `json:"assigned_to"`       // 分配给的专家
    Status          ExecutionStatus        `json:"status"`            // 执行状态
    StartTime       time.Time              `json:"start_time"`        // 开始时间
    EndTime         time.Time              `json:"end_time"`          // 结束时间
    Duration        time.Duration          `json:"duration"`          // 执行时长
    Result          *SpecialistResult      `json:"result"`            // 执行结果
    Error           string                 `json:"error,omitempty"`   // 错误信息
    Metadata        map[string]interface{} `json:"metadata"`          // 扩展信息
}
```

## 枚举类型定义

### TaskComplexity - 任务复杂度

```go
// TaskComplexity 任务复杂度枚举
type TaskComplexity int

const (
    TaskComplexityUnknown TaskComplexity = iota // 未知复杂度
    TaskComplexityLow                           // 低复杂度 - 简单问答
    TaskComplexityMedium                        // 中等复杂度 - 需要简单推理
    TaskComplexityHigh                          // 高复杂度 - 需要多步骤处理
    TaskComplexityVeryHigh                      // 极高复杂度 - 需要复杂协作
)

// String 返回复杂度的字符串表示
func (tc TaskComplexity) String() string {
    switch tc {
    case TaskComplexityLow:
        return "low"
    case TaskComplexityMedium:
        return "medium"
    case TaskComplexityHigh:
        return "high"
    case TaskComplexityVeryHigh:
        return "very_high"
    default:
        return "unknown"
    }
}

// IsComplex 判断是否为复杂任务
func (tc TaskComplexity) IsComplex() bool {
    return tc >= TaskComplexityHigh
}
```

### ActionType - 动作类型

```go
// ActionType 动作类型枚举
type ActionType int

const (
    ActionTypeUnknown ActionType = iota // 未知动作
    ActionTypeThink                     // 思考
    ActionTypePlan                      // 规划
    ActionTypeExecute                   // 执行
    ActionTypeReflect                   // 反思
    ActionTypeAnswer                    // 回答
    ActionTypeAddStep                   // 添加步骤
    ActionTypeModifyStep                // 修改步骤
    ActionTypeRemoveStep                // 移除步骤
    ActionTypeReorderSteps              // 重排步骤
    ActionTypeContinue                  // 继续执行
    ActionTypeComplete                  // 完成任务
)

// String 返回动作类型的字符串表示
func (at ActionType) String() string {
    switch at {
    case ActionTypeThink:
        return "think"
    case ActionTypePlan:
        return "plan"
    case ActionTypeExecute:
        return "execute"
    case ActionTypeReflect:
        return "reflect"
    case ActionTypeAnswer:
        return "answer"
    case ActionTypeAddStep:
        return "add_step"
    case ActionTypeModifyStep:
        return "modify_step"
    case ActionTypeRemoveStep:
        return "remove_step"
    case ActionTypeReorderSteps:
        return "reorder_steps"
    case ActionTypeContinue:
        return "continue"
    case ActionTypeComplete:
        return "complete"
    default:
        return "unknown"
    }
}
```

### StepStatus - 步骤状态

```go
// StepStatus 步骤状态枚举
type StepStatus int

const (
    StepStatusPending StepStatus = iota // 待执行
    StepStatusExecuting                 // 执行中
    StepStatusCompleted                 // 已完成
    StepStatusFailed                    // 执行失败
    StepStatusSkipped                   // 已跳过
    StepStatusBlocked                   // 被阻塞
)

// String 返回步骤状态的字符串表示
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

// IsTerminal 判断是否为终止状态
func (ss StepStatus) IsTerminal() bool {
    return ss == StepStatusCompleted || ss == StepStatusFailed || ss == StepStatusSkipped
}

// CanExecute 判断是否可以执行
func (ss StepStatus) CanExecute() bool {
    return ss == StepStatusPending
}
```

### ExecutionStatus - 执行状态

```go
// ExecutionStatus 执行状态枚举
type ExecutionStatus int

const (
    ExecutionStatusUnknown ExecutionStatus = iota // 未知状态
    ExecutionStatusStarted                        // 已开始
    ExecutionStatusRunning                        // 运行中
    ExecutionStatusSuccess                        // 成功完成
    ExecutionStatusFailed                         // 执行失败
    ExecutionStatusTimeout                        // 执行超时
    ExecutionStatusCancelled                      // 已取消
)

// String 返回执行状态的字符串表示
func (es ExecutionStatus) String() string {
    switch es {
    case ExecutionStatusStarted:
        return "started"
    case ExecutionStatusRunning:
        return "running"
    case ExecutionStatusSuccess:
        return "success"
    case ExecutionStatusFailed:
        return "failed"
    case ExecutionStatusTimeout:
        return "timeout"
    case ExecutionStatusCancelled:
        return "cancelled"
    default:
        return "unknown"
    }
}

// IsTerminal 判断是否为终止状态
func (es ExecutionStatus) IsTerminal() bool {
    return es == ExecutionStatusSuccess || es == ExecutionStatusFailed || 
           es == ExecutionStatusTimeout || es == ExecutionStatusCancelled
}

// IsSuccessful 判断是否成功
func (es ExecutionStatus) IsSuccessful() bool {
    return es == ExecutionStatusSuccess
}
```

### PlanUpdateType - 规划更新类型

```go
// PlanUpdateType 规划更新类型枚举
type PlanUpdateType int

const (
    PlanUpdateTypeUnknown PlanUpdateType = iota // 未知更新
    PlanUpdateTypeStepAdd                       // 添加步骤
    PlanUpdateTypeStepModify                    // 修改步骤
    PlanUpdateTypeStepRemove                    // 移除步骤
    PlanUpdateTypeStepReorder                   // 重排步骤
    PlanUpdateTypePriorityChange                // 优先级变更
    PlanUpdateTypeDependencyChange              // 依赖关系变更
    PlanUpdateTypeResourceReallocation          // 资源重新分配
    PlanUpdateTypeStrategyChange                // 策略变更
)

// String 返回更新类型的字符串表示
func (put PlanUpdateType) String() string {
    switch put {
    case PlanUpdateTypeStepAdd:
        return "step_add"
    case PlanUpdateTypeStepModify:
        return "step_modify"
    case PlanUpdateTypeStepRemove:
        return "step_remove"
    case PlanUpdateTypeStepReorder:
        return "step_reorder"
    case PlanUpdateTypePriorityChange:
        return "priority_change"
    case PlanUpdateTypeDependencyChange:
        return "dependency_change"
    case PlanUpdateTypeResourceReallocation:
        return "resource_reallocation"
    case PlanUpdateTypeStrategyChange:
        return "strategy_change"
    default:
        return "unknown"
    }
}

// IsStructuralChange 判断是否为结构性变更
func (put PlanUpdateType) IsStructuralChange() bool {
    return put == PlanUpdateTypeStepAdd || put == PlanUpdateTypeStepRemove || 
           put == PlanUpdateTypeStepReorder || put == PlanUpdateTypeDependencyChange
}
```

## 核心结构体类型

### TaskPlan - 任务规划

```go
// TaskPlan 任务规划结构
type TaskPlan struct {
    // 基本信息
    ID              string                 `json:"id"`                // 规划唯一标识
    Version         int                    `json:"version"`           // 规划版本
    Name            string                 `json:"name"`              // 规划名称
    Description     string                 `json:"description"`       // 规划描述
    
    // 规划状态
    Status          PlanStatus             `json:"status"`            // 规划状态
    CreatedAt       time.Time              `json:"created_at"`        // 创建时间
    UpdatedAt       time.Time              `json:"updated_at"`        // 更新时间
    
    // 步骤信息
    Steps           []*PlanStep            `json:"steps"`             // 规划步骤列表
    TotalSteps      int                    `json:"total_steps"`       // 总步骤数
    CurrentStep     int                    `json:"current_step"`      // 当前步骤索引
    CompletedSteps  int                    `json:"completed_steps"`   // 已完成步骤数
    
    // 依赖关系
    Dependencies    map[string][]string    `json:"dependencies"`      // 步骤依赖关系
    
    // 资源分配
    ResourceAllocation map[string]string   `json:"resource_allocation"` // 资源分配映射
    
    // 更新历史
    UpdateHistory   []*PlanUpdate          `json:"updateHistory"`    // 更新历史记录
    
    // 元数据
    Metadata        map[string]interface{} `json:"metadata"`          // 扩展信息
}

// PlanStatus 规划状态
type PlanStatus int

const (
    PlanStatusDraft PlanStatus = iota // 草稿状态
    PlanStatusActive                  // 活跃状态
    PlanStatusPaused                  // 暂停状态
    PlanStatusCompleted               // 已完成
    PlanStatusFailed                  // 执行失败
    PlanStatusCancelled               // 已取消
)

func (ps PlanStatus) String() string {
    switch ps {
    case PlanStatusDraft:
        return "draft"
    case PlanStatusActive:
        return "active"
    case PlanStatusPaused:
        return "paused"
    case PlanStatusCompleted:
        return "completed"
    case PlanStatusFailed:
        return "failed"
    case PlanStatusCancelled:
        return "cancelled"
    default:
        return "unknown"
    }
}
```

### PlanUpdate - 规划更新记录

```go
// PlanUpdate 规划更新记录
type PlanUpdate struct {
    ID              string                 `json:"id"`                // 更新记录ID
    Version         int                    `json:"version"`           // 目标版本
    UpdateType      PlanUpdateType         `json:"updateType"`       // 更新类型
    Description     string                 `json:"description"`       // 更新描述
    Reason          string                 `json:"reason"`            // 更新原因
    
    // 变更详情
    Changes         []*PlanChange          `json:"changes"`           // 具体变更列表
    
    // 时间信息
    Timestamp       time.Time              `json:"timestamp"`         // 更新时间
    
    // 影响评估
    ImpactAssessment *ImpactAssessment     `json:"impact_assessment"` // 影响评估
    
    // 元数据
    Metadata        map[string]interface{} `json:"metadata"`          // 扩展信息
}

// PlanChange 规划变更详情
type PlanChange struct {
    ChangeType      string                 `json:"change_type"`       // 变更类型
    TargetID        string                 `json:"target_id"`         // 目标对象ID
    Field           string                 `json:"field"`             // 变更字段
    OldValue        interface{}            `json:"old_value"`         // 原值
    NewValue        interface{}            `json:"new_value"`         // 新值
    Metadata        map[string]interface{} `json:"metadata"`          // 扩展信息
}

// ImpactAssessment 影响评估
type ImpactAssessment struct {
    AffectedSteps   []string               `json:"affected_steps"`    // 受影响的步骤
    TimeImpact      time.Duration          `json:"time_impact"`       // 时间影响
    ResourceImpact  map[string]float64     `json:"resource_impact"`   // 资源影响
    RiskLevel       string                 `json:"risk_level"`        // 风险等级
    Mitigation      []string               `json:"mitigation"`        // 缓解措施
}
```

### PlanStep - 规划步骤

```go
// PlanStep 规划步骤
type PlanStep struct {
    // 基本信息
    ID              string                 `json:"id"`                // 步骤唯一标识
    Name            string                 `json:"name"`              // 步骤名称
    Description     string                 `json:"description"`       // 步骤描述
    
    // 执行信息
    AssignedTo      string                 `json:"assigned_to"`       // 分配给的专家
    Priority        int                    `json:"priority"`          // 优先级 (1-10)
    EstimatedTime   time.Duration          `json:"estimated_time"`    // 预估执行时间
    
    // 状态信息
    Status          StepStatus             `json:"status"`            // 步骤状态
    StartTime       *time.Time             `json:"start_time"`        // 开始时间
    EndTime         *time.Time             `json:"end_time"`          // 结束时间
    ActualTime      time.Duration          `json:"actual_time"`       // 实际执行时间
    
    // 依赖关系
    Dependencies    []string               `json:"dependencies"`      // 前置依赖步骤ID
    Dependents      []string               `json:"dependents"`        // 后续依赖步骤ID
    
    // 执行参数
    Parameters      map[string]interface{} `json:"parameters"`        // 执行参数
    
    // 结果信息
    Result          *StepResult            `json:"result"`            // 执行结果
    
    // 元数据
    Metadata        map[string]interface{} `json:"metadata"`          // 扩展信息
}

// StepResult 步骤执行结果
type StepResult struct {
    Success         bool                   `json:"success"`           // 是否成功
    Output          interface{}            `json:"output"`            // 输出结果
    Error           string                 `json:"error,omitempty"`   // 错误信息
    Confidence      float64                `json:"confidence"`        // 结果置信度
    QualityScore    float64                `json:"qualityScore"`     // 质量评分
    Metadata        map[string]interface{} `json:"metadata"`          // 扩展信息
}
```