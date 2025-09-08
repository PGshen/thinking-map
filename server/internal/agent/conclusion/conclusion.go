package conclusion

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/PGshen/thinking-map/server/internal/agent/base/react"
	"github.com/PGshen/thinking-map/server/internal/agent/llmmodel"
	"github.com/PGshen/thinking-map/server/internal/agent/tool/node"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// 节点常量定义
const (
	// 节点键名
	initNodeKey           = "init_node" // 初始化节点
	analysisNodeKey       = "conclusion_analysis"
	generationNodeKey     = "conclusion_generation"
	refinementNodeKey     = "conclusion_refinement"
	toGenerationNodeKey   = "to_generation"
	toRefinementNodeKey   = "to_refinement"
	feedbackBranchNodeKey = "feedback_branch"
)

// 使用types.go中定义的UserMessage结构体

// ConclusionAgentState 结论生成Agent状态
type ConclusionAgentState struct {
	// 基本信息
	StartTime time.Time
	NodeID    string

	// 结论分析结果
	ConclusionType     string // 结论类型：分析型/创意型/决策型/学习型
	GenerationStrategy string // 生成策略

	// 结论内容
	InitialConclusion *schema.Message // 初步结论
	FinalConclusion   *schema.Message // 最终结论

	// 用户反馈
	UserFeedback *schema.Message // 用户对结论的反馈

	// 流程控制
	IsFirstRun  bool // 是否首次运行
	HasFeedback bool // 是否有用户反馈
}

// BuildConclusionAgent 创建结论生成Agent
func BuildConclusionAgent(ctx context.Context, option ...base.AgentOption) (compose.Runnable[*UserMessage, *schema.Message], error) {
	// 创建主模型
	cm, err := llmmodel.NewOpenAIModel(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 添加节点操作工具
	toolInfos, err := node.GetAllToolInfos(ctx)
	if err != nil {
		return nil, err
	}
	cm, _ = cm.WithTools(toolInfos)

	allTools := []tool.BaseTool{}
	nodeTools, err := node.GetAllNodeTools()
	if err != nil {
		return nil, err
	}
	allTools = append(allTools, nodeTools...)

	// 创建Graph，使用状态管理
	graph := compose.NewGraph[*UserMessage, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *ConclusionAgentState {
			return &ConclusionAgentState{
				StartTime:   time.Now(),
				IsFirstRun:  true,
				HasFeedback: false,
			}
		}),
	)

	// 创建结论分析节点的React Agent
	analysisReactAgent, err := react.NewAgent(ctx, react.ReactAgentConfig{
		ToolCallingModel: cm,
		ToolsConfig:      compose.ToolsNodeConfig{},
	}, option...)
	if err != nil {
		return nil, err
	}

	// 创建结论生成节点的React Agent
	generationReactAgent, err := react.NewAgent(ctx, react.ReactAgentConfig{
		ToolCallingModel: cm,
		ToolsConfig:      compose.ToolsNodeConfig{},
	}, option...)
	if err != nil {
		return nil, err
	}

	// 创建结论优化节点的React Agent
	refinementReactAgent, err := react.NewAgent(ctx, react.ReactAgentConfig{
		ToolCallingModel: cm,
		ToolsConfig:      compose.ToolsNodeConfig{},
	}, option...)
	if err != nil {
		return nil, err
	}

	// 创建节点处理函数
	analysisHandler := NewAnalysisNodeHandler(analysisReactAgent)
	generationHandler := NewGenerationNodeHandler(generationReactAgent)
	refinementHandler := NewRefinementNodeHandler(refinementReactAgent)

	// 添加结论分析节点
	err = graph.AddChatModelNode(analysisNodeKey, cm,
		compose.WithStatePreHandler(analysisHandler.PreHandler),
		compose.WithStatePostHandler(analysisHandler.PostHandler),
		compose.WithNodeName("结论分析节点"),
	)
	if err != nil {
		return nil, fmt.Errorf("添加结论分析节点失败: %w", err)
	}

	// 添加到生成节点的转换
	toGeneration := compose.ToList[*schema.Message]()
	graph.AddLambdaNode(toGenerationNodeKey, toGeneration, compose.WithNodeName("到结论生成节点"))

	// 添加结论生成节点
	err = graph.AddGraphNode(generationNodeKey, generationReactAgent.Graph,
		compose.WithStatePreHandler(generationHandler.PreHandler),
		compose.WithStatePostHandler(generationHandler.PostHandler),
		compose.WithNodeName("结论生成节点"),
	)
	if err != nil {
		return nil, fmt.Errorf("添加结论生成节点失败: %w", err)
	}

	// 注意：由于流程调整，不再需要到结论优化节点的转换
	// 现在反馈分支节点直接连接到结论优化节点或结论分析节点

	// 添加反馈分支
	feedbackBranch := compose.NewGraphBranch(func(ctx context.Context, input *UserMessage) (string, error) {
		var result string
		err = compose.ProcessState(ctx, func(ctx context.Context, state *ConclusionAgentState) error {
			// 检查状态中是否已有初步结论
			if state.InitialConclusion != nil {
				// 非首次运行，有用户反馈，进入结论优化节点
				state.HasFeedback = true
				result = refinementNodeKey
				return nil
			}

			// 首次运行或无结论内容，进入结论分析节点
			state.IsFirstRun = true
			result = analysisNodeKey
			return nil
		})
		return result, err
	}, map[string]bool{
		analysisNodeKey:   true,
		refinementNodeKey: true,
	})

	// 添加结论优化节点
	err = graph.AddGraphNode(refinementNodeKey, refinementReactAgent.Graph,
		compose.WithStatePreHandler(refinementHandler.PreHandler),
		compose.WithStatePostHandler(refinementHandler.PostHandler),
		compose.WithNodeName("结论优化节点"),
	)
	if err != nil {
		return nil, fmt.Errorf("添加结论优化节点失败: %w", err)
	}

	// 添加反馈分支节点
	graph.AddBranch(feedbackBranchNodeKey, feedbackBranch)

	// 添加初始化Lambda节点
	initNode := compose.InvokableLambda(func(ctx context.Context, input *UserMessage) (*UserMessage, error) {
		err = compose.ProcessState(ctx, func(ctx context.Context, state *ConclusionAgentState) error {
			// 记录用户消息中的信息到状态
			if input != nil {
				// 如果有结论内容，将其作为初步结论
				if input.Conclusion != "" {
					state.InitialConclusion = &schema.Message{
						Role:    schema.Assistant,
						Content: input.Conclusion,
					}
				}

				// 记录用户指令作为用户反馈
				if input.Instruction != "" {
					state.UserFeedback = &schema.Message{
						Role:    schema.User,
						Content: input.Instruction,
					}
				}
			}
			return nil
		})
		return input, err
	})

	// 添加初始化节点
	graph.AddLambdaNode(initNodeKey, initNode, compose.WithNodeName("init"))

	// 设置节点之间的连接关系
	// 1. 起始节点 -> 初始化节点
	graph.AddEdge(compose.START, initNodeKey)
	// 2. 初始化节点 -> 反馈分支节点（判断是首次生成还是优化已有结论）
	graph.AddEdge(initNodeKey, feedbackBranchNodeKey)

	// 首次生成路径
	// 2. 反馈分支节点 -> 结论分析节点（首次生成）
	graph.AddEdge(feedbackBranchNodeKey, analysisNodeKey)
	// 3. 结论分析节点 -> 到结论生成节点的转换
	graph.AddEdge(analysisNodeKey, toGenerationNodeKey)
	// 4. 到结论生成节点的转换 -> 结论生成节点
	graph.AddEdge(toGenerationNodeKey, generationNodeKey)
	// 5. 结论生成节点 -> 结束节点
	graph.AddEdge(generationNodeKey, compose.END)

	// 优化已有结论路径
	// 6. 反馈分支节点 -> 结论优化节点（有用户反馈时）
	graph.AddEdge(feedbackBranchNodeKey, refinementNodeKey)
	// 7. 结论优化节点 -> 结束节点
	graph.AddEdge(refinementNodeKey, compose.END)

	// 创建Agent
	agent, err := graph.Compile(ctx, compose.WithGraphName("conclusion"))
	if err != nil {
		return nil, fmt.Errorf("创建结论生成Agent失败: %w", err)
	}

	return agent, nil
}

// 结论分析节点处理函数
type AnalysisNodeHandler struct {
	agent *react.ReactAgent
}

// 创建结论分析节点处理函数
func NewAnalysisNodeHandler(agent *react.ReactAgent) *AnalysisNodeHandler {
	return &AnalysisNodeHandler{
		agent: agent,
	}
}

// 结论分析节点前置处理
func (h *AnalysisNodeHandler) PreHandler(ctx context.Context, input []*schema.Message, state *ConclusionAgentState) ([]*schema.Message, error) {
	// 构建分析提示词
	systemMsg := schema.SystemMessage(buildConclusionAnalysisPrompt())

	// 构建消息列表
	messages := []*schema.Message{systemMsg}

	// 如果有用户反馈，则使用用户反馈作为分析内容
	if state.UserFeedback != nil {
		messages = append(messages, state.UserFeedback)
	} else if len(input) > 0 {
		// 如果有输入消息，则使用最后一条输入消息
		messages = append(messages, input[len(input)-1])
	} else {
		// 如果没有用户反馈和输入消息，则使用默认提示
		defaultMsg := schema.UserMessage("请分析当前节点的上下文信息，确定结论类型和生成策略。")
		messages = append(messages, defaultMsg)
	}

	return messages, nil
}

// 结论分析节点后置处理
// 该函数负责从LLM输出中提取结论类型和生成策略，并将其保存到状态中
// 使用关键词匹配和语义分析来确定最适合的结论类型和生成策略
func (h *AnalysisNodeHandler) PostHandler(ctx context.Context, output *schema.Message, state *ConclusionAgentState) (*schema.Message, error) {
	// 从输出中提取结论类型和生成策略
	// 转换为小写以便不区分大小写进行匹配
	content := strings.ToLower(output.Content)

	// 增强的结论类型识别逻辑
	// 使用更多关键词匹配，提高识别准确性
	// 每种结论类型都有对应的关键词列表和策略列表
	type conclusionTypeMatch struct {
		type_      string   // 结论类型
		keywords   []string // 用于识别该类型的关键词
		strategies []string // 该类型可用的生成策略
	}

	// 定义各种结论类型的匹配器
	typeMatchers := []conclusionTypeMatch{
		{
			type_:    "分析型",
			keywords: []string{"分析型", "分析", "逻辑", "系统性", "结构化", "推理", "归纳", "演绎"},
			strategies: []string{"系统分析", "逻辑推理", "数据支持", "多角度思考"},
		},
		{
			type_:    "创意型",
			keywords: []string{"创意型", "创意", "创新", "创造", "发散", "想象", "灵感", "突破"},
			strategies: []string{"发散思维", "跨领域联想", "创新视角", "突破常规"},
		},
		{
			type_:    "决策型",
			keywords: []string{"决策型", "决策", "选择", "判断", "权衡", "评估", "方案", "行动"},
			strategies: []string{"多方案对比", "利弊权衡", "风险评估", "行动计划"},
		},
		{
			type_:    "学习型",
			keywords: []string{"学习型", "学习", "教育", "知识", "理解", "掌握", "记忆", "总结"},
			strategies: []string{"知识整合", "概念梳理", "要点提炼", "学习路径"},
		},
	}

	// 默认为分析型
	state.ConclusionType = "分析型"
	bestMatchScore := 0
	var bestMatchStrategies []string

	// 遍历所有类型匹配器，找出最佳匹配
	// 记录每种类型的匹配分数，用于调试和日志
	typeMatchScores := make(map[string]int)
	for _, matcher := range typeMatchers {
		score := 0
		matchedKeywords := []string{}

		// 计算每个关键词的匹配情况
		for _, keyword := range matcher.keywords {
			if strings.Contains(content, keyword) {
				score++
				matchedKeywords = append(matchedKeywords, keyword)
			}
		}

		typeMatchScores[matcher.type_] = score

		// 更新最佳匹配
		if score > bestMatchScore {
			bestMatchScore = score
			state.ConclusionType = matcher.type_
			bestMatchStrategies = matcher.strategies
		}
	}

	// 记录匹配分数，便于调试
	matchScoreLog := fmt.Sprintf("结论类型匹配分数: 分析型=%d, 创意型=%d, 决策型=%d, 学习型=%d, 最终选择=%s",
		typeMatchScores["分析型"], typeMatchScores["创意型"], typeMatchScores["决策型"], typeMatchScores["学习型"], state.ConclusionType)
	fmt.Println(matchScoreLog)

	// 从输出中提取或构建生成策略
	// 首先尝试从输出中直接提取策略信息
	strategyFound := false
	if strings.Contains(content, "策略") || strings.Contains(content, "方法") {
		// 简单提取策略描述的句子
		sentences := strings.Split(output.Content, "。")
		for _, sentence := range sentences {
			if strings.Contains(sentence, "策略") || strings.Contains(sentence, "方法") {
				state.GenerationStrategy = strings.TrimSpace(sentence)
				strategyFound = true
				fmt.Println("从输出中提取到策略:", state.GenerationStrategy)
				break
			}
		}
	}

	// 如果没有从输出中提取到策略，则基于结论类型构建策略
	if !strategyFound {
		// 从最佳匹配的策略列表中选择一个
		if len(bestMatchStrategies) > 0 {
			// 简单起见，这里选择第一个策略
			strategy := bestMatchStrategies[0]
			state.GenerationStrategy = fmt.Sprintf("基于%s的%s策略", state.ConclusionType, strategy)
			fmt.Println("基于结论类型构建策略:", state.GenerationStrategy)
		} else {
			// 兜底策略
			state.GenerationStrategy = "基于" + state.ConclusionType + "的综合思考策略"
			fmt.Println("使用兜底策略:", state.GenerationStrategy)
		}
	}

	// 确保策略不为空，提供默认策略作为兜底方案
	if state.GenerationStrategy == "" {
		state.GenerationStrategy = "基于" + state.ConclusionType + "的综合思考策略"
		fmt.Println("策略为空，使用默认策略:", state.GenerationStrategy)
	}

	return output, nil
}

// 结论生成节点处理函数
type GenerationNodeHandler struct {
	agent *react.ReactAgent
}

// 创建结论生成节点处理函数
func NewGenerationNodeHandler(agent *react.ReactAgent) *GenerationNodeHandler {
	return &GenerationNodeHandler{
		agent: agent,
	}
}

// 结论生成节点前置处理
func (h *GenerationNodeHandler) PreHandler(ctx context.Context, input []*schema.Message, state *ConclusionAgentState) ([]*schema.Message, error) {
	// 构建生成提示词
	systemMsg := schema.SystemMessage(buildConclusionGenerationPrompt())

	// 构建用户消息，包含结论类型和生成策略
	userContent := fmt.Sprintf("请基于节点上下文信息，生成一个%s结论。生成策略：%s",
		state.ConclusionType, state.GenerationStrategy)

	userMsg := schema.UserMessage(userContent)

	return []*schema.Message{systemMsg, userMsg}, nil
}

// 结论生成节点后置处理
func (h *GenerationNodeHandler) PostHandler(ctx context.Context, output *schema.Message, state *ConclusionAgentState) (*schema.Message, error) {
	// 保存初步结论
	state.InitialConclusion = output

	return output, nil
}

// 结论优化节点处理函数
type RefinementNodeHandler struct {
	agent *react.ReactAgent
}

// 创建结论优化节点处理函数
func NewRefinementNodeHandler(agent *react.ReactAgent) *RefinementNodeHandler {
	return &RefinementNodeHandler{
		agent: agent,
	}
}

// 结论优化节点前置处理
func (h *RefinementNodeHandler) PreHandler(ctx context.Context, input []*schema.Message, state *ConclusionAgentState) ([]*schema.Message, error) {
	// 构建优化提示词
	systemMsg := schema.SystemMessage(buildConclusionRefinementPrompt())

	// 构建消息列表，包含初步结论和用户反馈
	messages := []*schema.Message{systemMsg}

	// 添加初步结论
	if state.InitialConclusion != nil {
		// 将初步结论转换为助手消息
		initialConclusionMsg := schema.AssistantMessage(state.InitialConclusion.Content, nil)
		messages = append(messages, initialConclusionMsg)
	}

	// 添加用户反馈
	if state.UserFeedback != nil {
		messages = append(messages, state.UserFeedback)
	} else if len(input) > 0 {
		// 如果没有保存的用户反馈，但有输入，则使用最后一条输入作为用户反馈
		state.UserFeedback = input[len(input)-1]
		messages = append(messages, state.UserFeedback)
	} else {
		// 如果既没有保存的用户反馈，也没有输入，则使用默认提示
		userMsg := schema.UserMessage("请优化上述结论，使其更加清晰、准确和实用。")
		messages = append(messages, userMsg)
	}

	return messages, nil
}

// 结论优化节点后置处理
func (h *RefinementNodeHandler) PostHandler(ctx context.Context, output *schema.Message, state *ConclusionAgentState) (*schema.Message, error) {
	// 保存最终结论
	state.FinalConclusion = output

	return output, nil
}
