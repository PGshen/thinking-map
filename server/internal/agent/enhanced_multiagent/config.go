package enhanced_multiagent

import (
	"context"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

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

	// 对话思考提示模板
	ConversationalThinkingPromptTemplate string

	// 对话上下文分析提示模板
	ContextAnalysisPromptTemplate string

	// 反思提示模板
	ReflectionPromptTemplate string

	// 流式工具调用检查器
	StreamToolCallChecker func(ctx context.Context, modelOutput *schema.StreamReader[*schema.Message]) (bool, error)

	// 回调处理器
	Callbacks []EnhancedMultiAgentCallback

	// 会话配置
	SessionConfig SessionConfig
}

// SessionConfig 会话配置
type SessionConfig struct {
	// 会话超时时间
	Timeout time.Duration

	// 上下文窗口大小
	ContextWindowSize int

	// 历史消息保留数量
	MaxHistoryMessages int

	// 是否启用上下文压缩
	EnableContextCompression bool

	// 压缩阈值
	CompressionThreshold int

	// 是否启用意图分析
	EnableIntentAnalysis bool

	// 是否启用情感分析
	EnableEmotionalAnalysis bool

	// 扩展配置
	Extensions map[string]interface{}
}

// EnhancedHost 增强版主控Agent配置
type EnhancedHost struct {
	// 基础模型配置
	Model model.BaseChatModel

	// 思考提示模板
	ThinkingPromptTemplate string

	// 规划提示模板
	PlanningPromptTemplate string

	// 反思提示模板
	ReflectionPromptTemplate string

	// 对话上下文分析提示模板
	ContextAnalysisPromptTemplate string

	// 最大思考轮次
	MaxThinkingRounds int

	// 思考超时时间
	ThinkingTimeout time.Duration

	// 是否启用结构化输出
	EnableStructuredOutput bool

	// 是否启用调试模式
	DebugMode bool

	// 自定义图选项
	GraphOptions []compose.GraphAddNodeOpt

	// 扩展配置
	Extensions map[string]interface{}
}

// EnhancedSpecialist 增强版专家Agent配置
type EnhancedSpecialist struct {
	// 专家名称
	Name string

	// 专家描述
	Description string

	// 专家能力标签
	Capabilities []string

	// 专家模型
	Model model.BaseChatModel

	// 专家工具配置
	ToolsConfig *compose.ToolsNodeConfig

	// 专家提示模板
	PromptTemplate string

	// 专家优先级
	Priority int

	// 最大执行时间
	MaxExecutionTime time.Duration

	// 最大重试次数
	MaxRetries int

	// 是否启用并行执行
	EnableParallelExecution bool

	// 质量阈值
	QualityThreshold float64

	// 相关性阈值
	RelevanceThreshold float64

	// 是否为动态注册的专家
	IsDynamic bool

	// 专家状态
	Status SpecialistStatus

	// 扩展配置
	Extensions map[string]interface{}
}

// SpecialistStatus 专家状态
type SpecialistStatus int

const (
	SpecialistStatusActive SpecialistStatus = iota
	SpecialistStatusInactive
	SpecialistStatusMaintenance
	SpecialistStatusError
)

func (ss SpecialistStatus) String() string {
	switch ss {
	case SpecialistStatusActive:
		return "active"
	case SpecialistStatusInactive:
		return "inactive"
	case SpecialistStatusMaintenance:
		return "maintenance"
	case SpecialistStatusError:
		return "error"
	default:
		return "unknown"
	}
}

// DefaultConfig 返回默认配置
func DefaultConfig() *EnhancedMultiAgentConfig {
	return &EnhancedMultiAgentConfig{
		Name:                                 "enhanced-multiagent",
		MaxRounds:                            5,
		ComplexityThreshold:                  0.7,
		PlanTemplate:                         DefaultPlanTemplate,
		ConversationalThinkingPromptTemplate: DefaultConversationalThinkingPrompt,
		ContextAnalysisPromptTemplate:        DefaultContextAnalysisPrompt,
		ReflectionPromptTemplate:             DefaultReflectionPrompt,
		SessionConfig: SessionConfig{
			Timeout:                  30 * time.Minute,
			ContextWindowSize:        4000,
			MaxHistoryMessages:       20,
			EnableContextCompression: true,
			CompressionThreshold:     10,
			EnableIntentAnalysis:     true,
			EnableEmotionalAnalysis:  false,
			Extensions:               make(map[string]interface{}),
		},
		Host: EnhancedHost{
			ThinkingPromptTemplate:        DefaultThinkingPrompt,
			PlanningPromptTemplate:        DefaultPlanningPrompt,
			ReflectionPromptTemplate:      DefaultReflectionPrompt,
			ContextAnalysisPromptTemplate: DefaultContextAnalysisPrompt,
			MaxThinkingRounds:             3,
			ThinkingTimeout:               5 * time.Minute,
			EnableStructuredOutput:        true,
			DebugMode:                     false,
			Extensions:                    make(map[string]interface{}),
		},
		Specialists: []*EnhancedSpecialist{},
		Callbacks:   []EnhancedMultiAgentCallback{},
	}
}

// 默认提示模板
const (
	DefaultConversationalThinkingPrompt = `你是一个智能助手，需要分析用户的对话历史并进行思考。

对话历史：
{{.ConversationHistory}}

对话上下文分析：
{{.ConversationContext}}

请基于以上信息进行思考，分析用户的意图和需求，评估任务复杂度，并决定下一步行动。

请按照以下格式回答：
思考：[你的思考过程]
复杂度：[simple/moderate/complex]
推理：[你的推理过程]
下一步行动：[direct_answer/create_plan/execute_step/reflect]
回答策略：[direct/planned/clarification]
`

	DefaultContextAnalysisPrompt = `请分析以下对话历史，提取关键信息：

对话历史：
{{.ConversationHistory}}

请分析：
1. 对话轮次和是否为首次对话
2. 是否为延续性问题
3. 用户意图和置信度
4. 情感色调
5. 关键实体
6. 复杂度提示
7. 相关历史上下文
8. 上下文摘要

请以JSON格式返回分析结果。
`

	DefaultThinkingPrompt = `你是一个智能助手，请基于以下信息进行思考：

用户消息：{{.UserMessage}}
对话上下文：{{.ConversationContext}}

请分析用户需求，评估任务复杂度，并决定下一步行动。
`

	DefaultPlanningPrompt = `请为以下任务创建详细的执行计划：

任务描述：{{.TaskDescription}}
对话上下文：{{.ConversationContext}}

请创建一个分步骤的执行计划，包括：
1. 步骤描述
2. 分配的专家
3. 预估时间
4. 依赖关系

请以Markdown格式返回计划。
`

	DefaultReflectionPrompt = `请基于以下执行结果进行反思：

执行结果：{{.ExecutionResults}}
当前计划：{{.CurrentPlan}}

请评估：
1. 执行结果的质量
2. 是否需要继续执行
3. 是否需要调整计划
4. 建议的下一步行动

请提供详细的反思和建议。
`

	DefaultPlanTemplate = `# 任务执行计划

## 任务描述
{{.TaskDescription}}

## 执行步骤
{{range .Steps}}
### 步骤 {{.ID}}: {{.Description}}
- 状态: {{.Status}}
- 分配给: {{.AssignedTo}}
- 优先级: {{.Priority}}
- 预估时间: {{.EstimatedDuration}}
{{if .Dependencies}}- 依赖: {{.Dependencies}}{{end}}
{{end}}

## 执行进度
- 当前步骤: {{.CurrentStep}}/{{.TotalSteps}}
- 已完成: {{.CompletedSteps}}
- 版本: {{.Version}}
`
)
