package conclusion

import (
	"context"
	"encoding/json"
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
	AnalysisResult     *ConclusionAnalysisResult // 结构化的分析结果
	ConclusionType     string                    // 结论类型：分析型/创意型/决策型/学习型
	GenerationStrategy string                    // 生成策略

	// 结论内容
	InitialConclusion *schema.Message // 初步结论
	FinalConclusion   *schema.Message // 最终结论

	// 用户反馈和局部优化
	UserFeedback         *schema.Message           // 用户对结论的反馈
	LocalOptimizationReq *LocalOptimizationRequest // 局部优化请求
	ReferencedContent    string                    // 用户引用的具体内容

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

				// 设置引用内容
				if input.Reference != "" {
					state.ReferencedContent = input.Reference
				}

				// 创建局部优化请求
				if input.Instruction != "" || input.Reference != "" {
					state.LocalOptimizationReq = &LocalOptimizationRequest{
						TargetContent: input.Reference,
						Instruction:   input.Instruction,
						Context:       input.Conclusion, // 可以根据需要从其他地方获取上下文
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
// 该函数负责从LLM输出中解析JSON格式的分析结果，并将其保存到状态中
func (h *AnalysisNodeHandler) PostHandler(ctx context.Context, output *schema.Message, state *ConclusionAgentState) (*schema.Message, error) {
	// 从输出中提取JSON内容
	content := output.Content
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")

	if start == -1 || end == -1 || start >= end {
		// 如果没有找到有效的JSON，使用默认值
		fmt.Println("警告：未找到有效的JSON格式，使用默认分析结果")
		state.AnalysisResult = &ConclusionAnalysisResult{
			ConclusionType:     "分析型",
			GenerationStrategy: "基于分析型的综合思考策略",
			Reasoning:          "无法解析分析结果，使用默认配置",
			Confidence:         0.5,
		}
		state.ConclusionType = state.AnalysisResult.ConclusionType
		state.GenerationStrategy = state.AnalysisResult.GenerationStrategy
		return output, nil
	}

	jsonStr := content[start : end+1]

	// 解析JSON结果
	var analysisResult ConclusionAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &analysisResult); err != nil {
		// JSON解析失败，使用默认值
		fmt.Printf("警告：JSON解析失败: %v，使用默认分析结果\n", err)
		state.AnalysisResult = &ConclusionAnalysisResult{
			ConclusionType:     "分析型",
			GenerationStrategy: "基于分析型的综合思考策略",
			Reasoning:          "JSON解析失败，使用默认配置",
			Confidence:         0.5,
		}
	} else {
		// 验证结论类型是否有效
		validTypes := map[string]bool{
			"分析型": true,
			"创意型": true,
			"决策型": true,
			"学习型": true,
		}

		if !validTypes[analysisResult.ConclusionType] {
			fmt.Printf("警告：无效的结论类型 '%s'，使用默认类型\n", analysisResult.ConclusionType)
			analysisResult.ConclusionType = "分析型"
		}

		// 确保必要字段不为空
		if analysisResult.GenerationStrategy == "" {
			analysisResult.GenerationStrategy = "基于" + analysisResult.ConclusionType + "的综合思考策略"
		}

		if analysisResult.Confidence <= 0 || analysisResult.Confidence > 1 {
			analysisResult.Confidence = 0.8
		}

		state.AnalysisResult = &analysisResult
	}

	// 更新状态中的结论类型和生成策略
	state.ConclusionType = state.AnalysisResult.ConclusionType
	state.GenerationStrategy = state.AnalysisResult.GenerationStrategy

	fmt.Printf("结论分析结果: 类型=%s, 策略=%s, 置信度=%.2f\n",
		state.ConclusionType, state.GenerationStrategy, state.AnalysisResult.Confidence)

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

	// 构建消息列表
	messages := []*schema.Message{systemMsg}

	// 构建局部优化的用户消息
	var userContent strings.Builder

	// 添加现有结论
	if state.InitialConclusion != nil {
		userContent.WriteString("现有结论：\n")
		userContent.WriteString(state.InitialConclusion.Content)
		userContent.WriteString("\n\n")
	}

	// 添加优化指令
	var optimizationInstruction string
	if state.LocalOptimizationReq != nil && state.LocalOptimizationReq.Instruction != "" {
		optimizationInstruction = state.LocalOptimizationReq.Instruction
	} else if state.UserFeedback != nil {
		optimizationInstruction = state.UserFeedback.Content
	} else if len(input) > 0 {
		optimizationInstruction = input[len(input)-1].Content
		state.UserFeedback = input[len(input)-1]
	} else {
		optimizationInstruction = "请对结论进行优化，使其更加清晰、准确和实用。"
	}

	userContent.WriteString("优化指令：\n")
	userContent.WriteString(optimizationInstruction)
	userContent.WriteString("\n\n")

	// 添加引用内容
	var referencedContent string
	if state.LocalOptimizationReq != nil && state.LocalOptimizationReq.TargetContent != "" {
		referencedContent = state.LocalOptimizationReq.TargetContent
	} else if state.ReferencedContent != "" {
		referencedContent = state.ReferencedContent
	}

	if referencedContent != "" {
		userContent.WriteString("引用内容：\n")
		userContent.WriteString(referencedContent)
		userContent.WriteString("\n\n")
	}

	// 添加上下文信息
	if state.LocalOptimizationReq != nil && state.LocalOptimizationReq.Context != "" {
		userContent.WriteString("上下文信息：\n")
		userContent.WriteString(state.LocalOptimizationReq.Context)
		userContent.WriteString("\n\n")
	}

	userContent.WriteString("请根据以上信息进行精准的局部优化，输出完整的优化后结论。")

	userMsg := schema.UserMessage(userContent.String())
	messages = append(messages, userMsg)

	return messages, nil
}

// 结论优化节点后置处理
func (h *RefinementNodeHandler) PostHandler(ctx context.Context, output *schema.Message, state *ConclusionAgentState) (*schema.Message, error) {
	// 保存最终结论
	state.FinalConclusion = output

	return output, nil
}
