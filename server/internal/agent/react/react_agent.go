package react

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/PGshen/thinking-map/server/internal/pkg/logger"
	"github.com/PGshen/thinking-map/server/internal/pkg/utils"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

// AgentConfig is the configuration for the ReAct agent
type AgentConfig struct {
	// Model for reasoning and tool calling
	ToolCallingModel model.ToolCallingChatModel

	// ReasoningPrompt is the prompt template for reasoning
	ReasoningPrompt string

	// IsSupportStructuredOutput is whether the model supports structured output
	IsSupportStructuredOutput bool

	// Tools available to the agent
	ToolsConfig compose.ToolsNodeConfig

	// Message modifier for preprocessing
	MessageModifier MessageModifier

	// Maximum number of reasoning iterations
	MaxStep int

	// Enable debug mode
	DebugMode bool

	// Custom graph options
	GraphOptions []compose.GraphAddNodeOpt

	// Tools that should return directly without further reasoning
	// Map of tool name to whether it should return directly
	ToolReturnDirectly map[string]bool
}

// Agent represents the improved ReAct agent with proper reasoning-first architecture
type Agent struct {
	runnable         compose.Runnable[[]*schema.Message, *schema.Message]
	graph            *compose.Graph[[]*schema.Message, *schema.Message]
	graphAddNodeOpts []compose.GraphAddNodeOpt
	agentOptions     []AgentOption
	config           *AgentConfig
	monitor          *Monitor // 监控调试模块
}

type ReasoningOutput struct {
	Thought     string  `json:"thought"`
	Action      string  `json:"action"`
	FinalAnswer string  `json:"final_answer"`
	Confidence  float64 `json:"confidence"`
}

// Node keys for the improved ReAct graph
const (
	nodeKeyInit         = "init"
	nodeKeyReasoning    = "reasoning"
	nodeKeyTools        = "tools"
	nodeKeyToolsChecker = "tools_checker"
	nodeKeyToReasoning  = "to_reasoning"
	nodeKeyComplete     = "complete"
)

// NewAgent creates a new improved ReAct agent following the document design
// 实现完整的ReAct循环：思考→推理→决策→(可选)工具调用→观察→继续循环
func NewAgent(ctx context.Context, config *AgentConfig, opts ...AgentOption) (*Agent, error) {
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Initialize core components
	monitor := NewMonitor(config.DebugMode, LogLevelInfo, log.Default())
	agent := &Agent{
		config:  config,
		monitor: monitor,
	}

	// Setup chat model with tools
	chatModel, err := agent.setupChatModel(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup chat model: %w", err)
	}

	// Build the graph
	graph, err := agent.buildGraph(ctx, chatModel)
	if err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}

	// Compile the graph
	compileOpts := []compose.GraphCompileOption{compose.WithMaxRunSteps(config.MaxStep), compose.WithNodeTriggerMode(compose.AnyPredecessor), compose.WithGraphName("ReactGraph")}
	runnable, err := graph.Compile(ctx, compileOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile graph: %w", err)
	}

	agent.runnable = runnable
	agent.graph = graph
	agent.graphAddNodeOpts = []compose.GraphAddNodeOpt{compose.WithGraphCompileOptions(compileOpts...)}
	agent.agentOptions = opts

	return agent, nil
}

// setupChatModel initializes and configures the chat model with tools
func (a *Agent) setupChatModel(ctx context.Context) (model.BaseChatModel, error) {
	a.monitor.Debug("Setup", "Setting up chat model with tools", nil)

	// Generate tool infos for model binding
	toolInfos, err := genToolInfos(ctx, a.config.ToolsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tool infos: %w", err)
	}

	// Bind tools to chat model (required for tool calling capability)
	chatModel, err := a.config.ToolCallingModel.WithTools(toolInfos)
	if err != nil {
		return nil, fmt.Errorf("failed to bind tools: %w", err)
	}

	a.monitor.Info("Setup", "Chat model setup completed", map[string]interface{}{
		"tool_count": len(toolInfos),
		"tool_infos": toolInfos,
	})

	return chatModel, nil
}

// buildGraph constructs the complete ReAct graph with all nodes and edges
func (a *Agent) buildGraph(ctx context.Context, chatModel model.BaseChatModel) (*compose.Graph[[]*schema.Message, *schema.Message], error) {
	a.monitor.Debug("Graph", "Building ReAct graph", nil)

	// Create global state generator function
	stateGenerator := func(ctx context.Context) *AgentState {
		return &AgentState{
			Messages:         make([]*schema.Message, 0),
			ReasoningHistory: make([]ReasoningDecision, 0),
			Iteration:        0,
			MaxIterations:    a.config.MaxStep,
			Completed:        false,
			FinalAnswer:      "",
		}
	}

	// Build the improved ReAct graph with global state support
	graph := compose.NewGraph[[]*schema.Message, *schema.Message](compose.WithGenLocalState(stateGenerator))

	// Add all nodes
	if err := a.addInitNode(graph); err != nil {
		return nil, fmt.Errorf("failed to add init node: %w", err)
	}

	if err := a.addReasoningNode(graph, chatModel); err != nil {
		return nil, fmt.Errorf("failed to add reasoning node: %w", err)
	}

	if err := a.addToReasoningNode(graph); err != nil {
		return nil, fmt.Errorf("failed to add to-reasoning node: %w", err)
	}

	if err := a.addToolsNode(ctx, graph); err != nil {
		return nil, fmt.Errorf("failed to add tools node: %w", err)
	}

	if err := a.addToolsCheckerNode(graph); err != nil {
		return nil, fmt.Errorf("failed to add tools checker node: %w", err)
	}

	if err := a.addCompleteNode(graph); err != nil {
		return nil, fmt.Errorf("failed to add complete node: %w", err)
	}

	// Add branches
	if err := a.addDecisionBranch(graph); err != nil {
		return nil, fmt.Errorf("failed to add decision branch: %w", err)
	}

	if err := a.addToolsCheckerBranch(graph); err != nil {
		return nil, fmt.Errorf("failed to add tools checker branch: %w", err)
	}

	// Add edges
	if err := a.addGraphEdges(graph); err != nil {
		return nil, fmt.Errorf("failed to add graph edges: %w", err)
	}

	a.monitor.Info("Graph", "ReAct graph built successfully", nil)
	return graph, nil
}

// addInitNode adds the initialization node to the graph
func (a *Agent) addInitNode(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	return graph.AddLambdaNode(nodeKeyInit, compose.InvokableLambda(
		func(ctx context.Context, input []*schema.Message) ([]*schema.Message, error) {
			return input, nil // Pass through input to next node
		}),
		compose.WithStatePreHandler(a.initNodePreHandler),
	)
}

// initNodePreHandler handles the pre-processing for the init node
func (a *Agent) initNodePreHandler(ctx context.Context, input []*schema.Message, state *AgentState) ([]*schema.Message, error) {
	a.monitor.Info("Init", "进入初始化节点", map[string]interface{}{
		"input_message_count": len(input),
		"input_messages":      input,
	})

	// Initialize state with input messages
	state.Messages = make([]*schema.Message, len(input))
	copy(state.Messages, input)

	a.monitor.Debug("Init", "状态初始化完成", map[string]interface{}{
		"messages_copied": len(state.Messages),
		"max_iterations":  state.MaxIterations,
	})

	a.monitor.Info("Init", "离开初始化节点", nil)
	return input, nil
}

// addReasoningNode adds the reasoning node to the graph
func (a *Agent) addReasoningNode(graph *compose.Graph[[]*schema.Message, *schema.Message], chatModel model.BaseChatModel) error {
	return graph.AddChatModelNode(nodeKeyReasoning, chatModel,
		compose.WithStatePreHandler(a.reasoningNodePreHandler),
		compose.WithStatePostHandler(a.reasoningNodePostHandler),
		// compose.WithStreamStatePostHandler(a.reasoningNodeStreamPostHandler),
	)
}

// reasoningNodePreHandler handles the pre-processing for the reasoning node
func (a *Agent) reasoningNodePreHandler(ctx context.Context, input []*schema.Message, state *AgentState) ([]*schema.Message, error) {
	a.monitor.Info("Reasoning", "进入推理节点", map[string]interface{}{
		"iteration":     state.Iteration,
		"message_count": len(state.Messages),
		"messages":      state.Messages,
	})

	// Build reasoning prompt with detailed tool information
	if a.config.ReasoningPrompt == "" {
		a.config.ReasoningPrompt = buildReasoningSystemPrompt()
	}
	reasoningPrompt := a.config.ReasoningPrompt

	// Create messages for reasoning
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: reasoningPrompt,
		},
	}

	// Add conversation history
	messages = append(messages, state.Messages...)

	a.monitor.Debug("Reasoning", "推理消息准备完成", map[string]interface{}{
		"total_messages":       len(messages),
		"system_prompt_length": len(reasoningPrompt),
	})

	return messages, nil
}

// reasoningNodePostHandler handles the post-processing for the reasoning node
func (a *Agent) reasoningNodePostHandler(ctx context.Context, output *schema.Message, state *AgentState) (*schema.Message, error) {
	if output == nil {
		a.monitor.Error("Reasoning", "推理响应为空", nil)
		return nil, NewReasoningError("no reasoning response generated", nil)
	}

	// Parse reasoning response
	reasoning, err := parseReasoningResponse(output)
	if err != nil {
		a.monitor.Error("Reasoning", "解析推理响应失败", err)
		return nil, NewReasoningError("failed to parse reasoning response", err)
	}

	// Update state with reasoning result
	state.ReasoningHistory = append(state.ReasoningHistory, *reasoning)
	state.Iteration++

	// Record model response in state messages
	state.Messages = append(state.Messages, output)

	a.monitor.Info("Reasoning", "离开推理节点", map[string]interface{}{
		"iteration": state.Iteration,
		"action":    reasoning.Action,
		"thought":   reasoning.Thought,
		"output":    output,
	})

	return output, nil
}

func (a *Agent) reasoningNodeStreamPostHandler(ctx context.Context, sr *schema.StreamReader[*schema.Message], state *AgentState) (*schema.StreamReader[*schema.Message], error) {
	r, w := schema.Pipe[*schema.Message](10240)
	// 使用流式JSON解析器解析ReasoningOutput
	matcher := utils.NewSimplePathMatcher()
	// 使用增量模式避免重复内容
	parser := utils.NewStreamingJsonParser(matcher, true, true)

	var thought, action, finalAnswer strings.Builder

	// 注册路径匹配器来提取thought和final_answer字段
	matcher.On("thought", func(value interface{}, path []interface{}) {
		// fmt.Print("thought:", value)
		if str, ok := value.(string); ok {
			w.Send(&schema.Message{
				Role:    schema.Assistant,
				Content: str,
			}, nil)
			thought.WriteString(str)
		}
	})

	matcher.On("action", func(value interface{}, path []interface{}) {
		if str, ok := value.(string); ok {
			action.WriteString(str)
		}
	})

	matcher.On("final_answer", func(value interface{}, path []interface{}) {
		if str, ok := value.(string); ok {
			w.Send(&schema.Message{
				Role:    schema.Assistant,
				Content: str,
			}, nil)
			finalAnswer.WriteString(str)
		}
	})
	ss := sr.Copy(2) // 复制流，一个用于组装解析工具和推理、一个用于实时解析
	go func() {
		fullMsgs := make([]*schema.Message, 0)
		defer func() {
			ss[1].Close()
			fullMsg, err2 := schema.ConcatMessages(fullMsgs)
			if err2 != nil {
				logger.Warn("concat message failed", zap.Error(err2))
				return
			}
			// Parse reasoning response
			reasoning, err := parseReasoningResponse(fullMsg)
			if err != nil {
				a.monitor.Error("Reasoning", "解析推理响应失败", err)
				return
			}

			// Update state with reasoning result
			state.ReasoningHistory = append(state.ReasoningHistory, *reasoning)
			state.Iteration++

			// Record model response in state messages
			state.Messages = append(state.Messages, fullMsg)

			a.monitor.Info("Reasoning", "离开推理节点", map[string]interface{}{
				"iteration": state.Iteration,
				"action":    reasoning.Action,
				"thought":   reasoning.Thought,
				"output":    fullMsg,
			})
		}()
	outer:
		for {
			select {
			case <-ctx.Done():
				fmt.Println("context done", ctx.Err())
				return
			default:
				chunk, err2 := sr.Recv()
				if err2 != nil {
					if errors.Is(err2, io.EOF) {
						fmt.Println()
						break outer
					}
				}
				// fmt.Printf("%s", chunk.Content)
				fullMsgs = append(fullMsgs, chunk)
			}
		}
	}()

	// 实施解析输入流
	for {
		msg, err := ss[0].Recv()
		if err != nil {
			fmt.Println(err)
			if err.Error() == "EOF" {
				break
			}
			// 解析失败时回退到原始内容
			return sr, nil
		}

		// fmt.Print(msg.Content)
		// 尝试解析JSON
		if err := parser.Write(msg.Content); err != nil {
			fmt.Println(err)
			a.monitor.Warn("Reasoning", "解析推理响应失败", map[string]interface{}{
				"original_content": msg.Content,
			})
		}
	}
	w.Close()

	// 创建新的StreamReader
	return r, nil
}

// addToReasoningNode adds the to-reasoning conversion node
func (a *Agent) addToReasoningNode(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	return graph.AddLambdaNode(nodeKeyToReasoning, compose.ToList[*schema.Message]())
}

// addToolsNode adds the tools execution node to the graph
func (a *Agent) addToolsNode(ctx context.Context, graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	toolsNodeConfig := &a.config.ToolsConfig
	toolsNode, err := compose.NewToolNode(ctx, toolsNodeConfig)
	if err != nil {
		return fmt.Errorf("failed to create tools node: %w", err)
	}

	return graph.AddToolsNode(nodeKeyTools, toolsNode,
		compose.WithStatePreHandler(a.toolsNodePreHandler),
		compose.WithStatePostHandler(a.toolsNodePostHandler),
	)
}

// toolsNodePreHandler handles the pre-processing for the tools node
func (a *Agent) toolsNodePreHandler(ctx context.Context, input *schema.Message, state *AgentState) (*schema.Message, error) {
	if len(state.ReasoningHistory) == 0 {
		a.monitor.Warn("Tools", "进入工具节点但无推理历史", nil)
		return input, nil // Skip if no reasoning available
	}

	reasoning := &state.ReasoningHistory[len(state.ReasoningHistory)-1]

	a.monitor.Info("Tools", "进入工具执行节点", map[string]interface{}{
		"iteration":        state.Iteration,
		"tool_calls_count": len(reasoning.ToolCalls),
		"tool_calls":       reasoning.ToolCalls,
	})

	for i, toolCall := range reasoning.ToolCalls {
		a.monitor.Debug("Tools", "准备执行工具", map[string]interface{}{
			"tool_index": i,
			"tool_name":  toolCall.Function.Name,
			"tool_id":    toolCall.ID,
		})
	}

	return input, nil
}

// toolsNodePostHandler handles the post-processing for the tools node
func (a *Agent) toolsNodePostHandler(ctx context.Context, output []*schema.Message, state *AgentState) ([]*schema.Message, error) {
	a.monitor.Info("Tools", "工具执行完成", map[string]interface{}{
		"output_count":      len(output),
		"tool_calls_output": output,
	})

	// Update state with tool result
	state.Messages = append(state.Messages, output...)

	// Check if any tool should return directly
	if a.config.ToolReturnDirectly != nil {
		for _, msg := range output {
			if msg.ToolCallID != "" {
				// Find the corresponding tool call to get tool name
				for _, reasoning := range state.ReasoningHistory {
					for _, toolCall := range reasoning.ToolCalls {
						if toolCall.ID == msg.ToolCallID {
							if shouldReturn, exists := a.config.ToolReturnDirectly[toolCall.Function.Name]; exists && shouldReturn {
								state.ReturnDirectlyToolCallID = msg.ToolCallID
								a.monitor.Info("Tools", "工具标记为直接返回", map[string]interface{}{
									"tool_name":    toolCall.Function.Name,
									"tool_call_id": msg.ToolCallID,
								})
								break
							}
						}
					}
				}
			}
		}
	}

	a.monitor.Info("Tools", "离开工具执行节点", nil)
	return output, nil
}

// addToolsCheckerNode adds the tools checker node to the graph
func (a *Agent) addToolsCheckerNode(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	return graph.AddLambdaNode(nodeKeyToolsChecker, compose.InvokableLambda(
		func(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
			return a.toolsCheckerNodeHandler(ctx, input)
		}),
	)
}

// toolsCheckerNodeHandler handles the tools checker node logic
func (a *Agent) toolsCheckerNodeHandler(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
	a.monitor.Info("ToolsChecker", "进入工具检查节点", map[string]interface{}{
		"input_count":    len(input),
		"input_messages": input,
	})

	// Find the message that should return directly
	var directReturnMsg *schema.Message
	err := compose.ProcessState(ctx, func(_ context.Context, state *AgentState) error {
		if state.ReturnDirectlyToolCallID != "" {
			a.monitor.Debug("ToolsChecker", "检查直接返回工具调用", map[string]interface{}{
				"tool_call_id": state.ReturnDirectlyToolCallID,
			})
			// Find the message with the matching tool call ID
			for _, msg := range input {
				if msg.ToolCallID == state.ReturnDirectlyToolCallID {
					directReturnMsg = msg
					state.FinalAnswer = msg.Content
					state.Completed = true
					a.monitor.Info("ToolsChecker", "找到直接返回消息", map[string]interface{}{
						"tool_call_id":   msg.ToolCallID,
						"content_length": len(msg.Content),
					})
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		a.monitor.Error("ToolsChecker", "处理状态时出错", err)
		return nil, err
	}

	if directReturnMsg != nil {
		a.monitor.Info("ToolsChecker", "离开工具检查节点 - 直接返回", nil)
		return directReturnMsg, nil
	}

	// If no direct return, return the first message for further processing
	if len(input) > 0 {
		a.monitor.Info("ToolsChecker", "离开工具检查节点 - 继续处理", nil)
		return input[0], nil
	}

	a.monitor.Info("ToolsChecker", "离开工具检查节点 - 空消息", nil)
	return &schema.Message{Role: schema.Assistant, Content: ""}, nil
}

// addCompleteNode adds the completion node to the graph
func (a *Agent) addCompleteNode(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	return graph.AddLambdaNode(nodeKeyComplete, compose.InvokableLambda(
		func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
			return input, nil
		}),
		compose.WithStatePreHandler(a.completeNodePreHandler),
		compose.WithStatePostHandler(a.completeNodePostHandler),
	)
}

// completeNodePreHandler handles the pre-processing for the complete node
func (a *Agent) completeNodePreHandler(ctx context.Context, input *schema.Message, state *AgentState) (*schema.Message, error) {
	a.monitor.Info("Complete", "进入完成节点", map[string]interface{}{
		"completed":  state.Completed,
		"iterations": state.Iteration,
	})

	// Only proceed if this is actually an end condition
	if !state.Completed {
		a.monitor.Debug("Complete", "跳过完成节点，未完成", map[string]interface{}{
			"completed": state.Completed,
		})
		return input, nil
	}

	a.monitor.Info("Complete", "Agent执行已完成", map[string]interface{}{
		"iterations":          state.Iteration,
		"final_answer_length": len(state.FinalAnswer),
	})

	return input, nil
}

// completeNodePostHandler handles the post-processing for the complete node
func (a *Agent) completeNodePostHandler(ctx context.Context, output *schema.Message, state *AgentState) (*schema.Message, error) {
	// Create final message based on state
	var finalMessage *schema.Message

	if state.FinalAnswer != "" {
		finalMessage = &schema.Message{
			Role:    schema.Assistant,
			Content: state.FinalAnswer,
		}
		a.monitor.Debug("Complete", "使用最终答案创建响应", map[string]interface{}{
			"final_answer_length": len(state.FinalAnswer),
		})
	} else if len(state.Messages) > 0 {
		finalMessage = state.Messages[len(state.Messages)-1]
		a.monitor.Debug("Complete", "使用最后消息创建响应", nil)
	} else {
		finalMessage = &schema.Message{
			Role:    schema.Assistant,
			Content: "I apologize, but I was unable to provide a response.",
		}
		a.monitor.Warn("Complete", "使用默认错误消息", nil)
	}

	a.monitor.Info("Complete", "离开完成节点", map[string]interface{}{
		"final_message_length": len(finalMessage.Content),
	})

	return finalMessage, nil
}

// addDecisionBranch adds the decision branch to the graph
func (a *Agent) addDecisionBranch(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	return graph.AddBranch(nodeKeyReasoning, compose.NewStreamGraphBranch(
		a.decisionBranchHandler,
		map[string]bool{
			nodeKeyTools:       true,
			nodeKeyComplete:    true,
			nodeKeyToReasoning: true,
		},
	))
}

// decisionBranchHandler handles the decision branch logic
func (a *Agent) decisionBranchHandler(ctx context.Context, msgsStream *schema.StreamReader[*schema.Message]) (endNode string, err error) {
	msgsStream.Close()

	a.monitor.Info("Decision", "进入决策分支", nil)

	err = compose.ProcessState(ctx, func(_ context.Context, state *AgentState) error {
		if len(state.ReasoningHistory) == 0 {
			a.monitor.Error("Decision", "无推理结果可用", nil)
			return NewDecisionError("no reasoning result available", nil)
		}

		reasoning := &state.ReasoningHistory[len(state.ReasoningHistory)-1]
		a.monitor.Debug("Decision", "分析推理结果", map[string]interface{}{
			"iteration": state.Iteration,
			"action":    reasoning.Action,
		})

		// Check if max iterations reached
		if state.Iteration > state.MaxIterations {
			a.monitor.Info("Decision", "达到最大迭代次数，结束执行", map[string]interface{}{
				"iteration":      state.Iteration,
				"max_iterations": state.MaxIterations,
			})
			// Force final answer
			state.FinalAnswer = "Maximum iterations reached. Unable to complete the task."
			state.Completed = true
			endNode = nodeKeyComplete
			return nil
		}

		// Make decision based on reasoning result
		switch reasoning.Action {
		case "tool_call":
			// Validate tool calls exist
			if len(reasoning.ToolCalls) == 0 {
				a.monitor.Error("Decision", "未找到工具调用", fmt.Errorf("tool_call action but no ToolCalls"))
				return NewDecisionError("no tool calls found", nil)
			}
			// Route to tools node
			a.monitor.Info("Decision", "路由到工具节点", map[string]interface{}{
				"tool_calls_count": len(reasoning.ToolCalls),
				"tool_calls":       reasoning.ToolCalls,
			})
			endNode = nodeKeyTools
		case "final_answer":
			// Set final answer and mark as completed
			state.FinalAnswer = reasoning.FinalAnswer
			state.Completed = true
			a.monitor.Info("Decision", "路由到完成节点", map[string]interface{}{
				"final_answer_length": len(reasoning.FinalAnswer),
			})
			endNode = nodeKeyComplete
		default:
			a.monitor.Info("Decision", "未知动作，继续推理", map[string]interface{}{
				"action": reasoning.Action,
			})
			// Route back to reasoning for retry
			endNode = nodeKeyToReasoning
		}

		return nil
	})
	if err != nil {
		a.monitor.Error("Decision", "决策处理失败", err)
		return "", err
	}

	a.monitor.Info("Decision", "离开决策分支", map[string]interface{}{
		"next_node": endNode,
	})
	return endNode, nil
}

// addToolsCheckerBranch adds the tools checker branch to the graph
func (a *Agent) addToolsCheckerBranch(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	return graph.AddBranch(nodeKeyToolsChecker, compose.NewStreamGraphBranch(
		a.toolsCheckerBranchHandler,
		map[string]bool{
			nodeKeyComplete:    true,
			nodeKeyToReasoning: true,
		},
	))
}

// toolsCheckerBranchHandler handles the tools checker branch logic
func (a *Agent) toolsCheckerBranchHandler(ctx context.Context, msgStream *schema.StreamReader[*schema.Message]) (endNode string, err error) {
	msgStream.Close()

	a.monitor.Info("ToolsCheckerBranch", "进入工具检查分支", nil)

	err = compose.ProcessState(ctx, func(_ context.Context, state *AgentState) error {
		if state.ReturnDirectlyToolCallID != "" && state.Completed {
			a.monitor.Info("ToolsCheckerBranch", "工具结果将直接返回", map[string]interface{}{
				"tool_call_id": state.ReturnDirectlyToolCallID,
			})
			endNode = nodeKeyComplete
		} else {
			a.monitor.Debug("ToolsCheckerBranch", "继续到推理节点", map[string]interface{}{
				"completed":          state.Completed,
				"return_directly_id": state.ReturnDirectlyToolCallID,
			})
			endNode = nodeKeyToReasoning
		}
		return nil
	})
	if err != nil {
		a.monitor.Error("ToolsCheckerBranch", "工具检查分支处理失败", err)
		return "", err
	}

	a.monitor.Info("ToolsCheckerBranch", "离开工具检查分支", map[string]interface{}{
		"next_node": endNode,
	})
	return endNode, nil
}

// addGraphEdges adds all edges to connect the graph nodes
func (a *Agent) addGraphEdges(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	a.monitor.Debug("Graph", "添加图边连接", nil)

	// Start -> Init (entry point)
	if err := graph.AddEdge(compose.START, nodeKeyInit); err != nil {
		return fmt.Errorf("failed to add start edge: %w", err)
	}

	// Init -> Reasoning (always start with reasoning)
	if err := graph.AddEdge(nodeKeyInit, nodeKeyReasoning); err != nil {
		return fmt.Errorf("failed to add init to reasoning edge: %w", err)
	}

	// Tools -> ToolsChecker (check if tool should return directly)
	if err := graph.AddEdge(nodeKeyTools, nodeKeyToolsChecker); err != nil {
		return fmt.Errorf("failed to add tools to tools checker edge: %w", err)
	}

	// ToList -> Reasoning (continue reasoning after conversion)
	if err := graph.AddEdge(nodeKeyToReasoning, nodeKeyReasoning); err != nil {
		return fmt.Errorf("failed to add tolist to reasoning edge: %w", err)
	}

	// End -> END (final output)
	if err := graph.AddEdge(nodeKeyComplete, compose.END); err != nil {
		return fmt.Errorf("failed to add complete to end edge: %w", err)
	}

	a.monitor.Debug("Graph", "图边连接完成", nil)
	return nil
}

// Generate executes the agent with comprehensive error handling and monitoring
func (a *Agent) Generate(ctx context.Context, messages []*schema.Message, opts ...AgentOption) (*schema.Message, error) {
	// Validate input
	if len(messages) == 0 {
		return nil, fmt.Errorf("input messages cannot be empty")
	}

	a.monitor.Info("Agent", "开始Agent生成", map[string]interface{}{
		"input_message_count": len(messages),
		"input_messages":      messages,
		"max_steps":           a.config.MaxStep,
	})

	startTime := time.Now()
	options := GetComposeOptions(opts...)
	options = append(options, GetComposeOptions(a.agentOptions...)...) //合并option
	result, err := a.runnable.Invoke(ctx, messages, options...)
	if err != nil {
		a.monitor.Error("Agent", "Agent执行失败", err)
		return nil, fmt.Errorf("agent execution failed: %w", err)
	}

	a.monitor.Info("Agent", "Agent生成成功完成", map[string]interface{}{
		"execution_time": time.Since(startTime),
		"result_length":  len(result.Content),
	})

	return result, nil
}

// Stream executes the agent with streaming support and comprehensive monitoring
func (a *Agent) Stream(ctx context.Context, messages []*schema.Message, opts ...AgentOption) (*schema.StreamReader[*schema.Message], error) {
	// Validate input
	if len(messages) == 0 {
		return nil, fmt.Errorf("input messages cannot be empty")
	}

	a.monitor.Info("Agent", "开始Agent流式处理", map[string]interface{}{
		"input_message_count": len(messages),
		"input_messages":      messages,
		"max_steps":           a.config.MaxStep,
	})

	startTime := time.Now()

	// Execute streaming with error handling
	options := GetComposeOptions(opts...)
	options = append(options, GetComposeOptions(a.agentOptions...)...) //合并option
	stream, err := a.runnable.Stream(ctx, messages, options...)
	if err != nil {
		a.monitor.Error("Agent", "启动流式处理失败", err)
		return nil, fmt.Errorf("failed to start streaming: %w", err)
	}

	a.monitor.Info("Agent", "流式处理启动成功", map[string]interface{}{
		"startup_time": time.Since(startTime),
	})

	return stream, nil
}

// validateConfig validates the agent configuration
func validateConfig(config *AgentConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if config.ToolCallingModel == nil {
		return fmt.Errorf("ToolCallingModel cannot be nil")
	}
	if config.MaxStep <= 0 {
		config.MaxStep = 10 // default value
	}
	return nil
}

// genToolInfos generates tool information from tools config
func genToolInfos(ctx context.Context, config compose.ToolsNodeConfig) ([]*schema.ToolInfo, error) {
	toolInfos := make([]*schema.ToolInfo, 0, len(config.Tools))
	for _, tool := range config.Tools {
		info, err := tool.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get tool info: %w", err)
		}
		toolInfos = append(toolInfos, info)
	}
	return toolInfos, nil
}

// buildReasoningSystemPrompt builds the system prompt for reasoning
func buildReasoningSystemPrompt() string {
	return `你是一个智能AI助手，遵循结构化推理过程来解决问题。

## 推理框架
你必须严格按照以下格式进行推理：

1. **分析问题**：仔细理解用户的问题或需求
2. **制定策略**：思考解决问题的步骤和方法
3. **选择行动**：决定下一步要采取的行动
4. **执行决策**：根据选择执行相应的操作

## action行动选项,仅限以下取值
- **continue**：需要继续思考或分析，还没有足够信息做决定
- **tool_call**：需要调用工具来获取信息或执行操作
- **final_answer**：已经有足够信息，可以提供最终答案

## 响应格式
你必须严格按照以下JSON格式回复：

{
  "thought": "详细的推理过程，包括问题分析、策略制定等",
  "action": "continue|tool_call|final_answer",
  "final_answer": "仅在action为final_answer时提供",
  "confidence": 0.8
}

## 推理示例

**用户问题**："帮我查找今天的天气情况"

**正确的推理过程**：
{
  "thought": "用户想要了解今天的天气情况。为了提供准确的天气信息，我需要：1. 确定用户的地理位置（如果没有提供）2. 调用天气查询工具获取当前天气数据 3. 整理并呈现天气信息。由于我没有用户的具体位置信息，也没有实时天气数据，我需要调用天气查询工具。",
  "action": "tool_call",
  "final_answer": "",
  "confidence": 0.9
}

**如果有工具调用结果后**：
{
  "thought": "已经通过天气工具获取到了今天的天气数据，包括温度、湿度、风速等信息。现在可以为用户提供完整的天气报告。",
  "action": "final_answer",
  "final_answer": "根据最新数据，今天天气晴朗，温度22-28°C，湿度65%，东南风3级，适合外出活动。",
  "confidence": 0.95
}

## 重要原则
- 始终先思考再行动，确保推理过程清晰完整
- 如果信息不足，优先选择continue或tool_call获取更多信息
- 只有在确信能够提供准确、完整答案时才选择final_answer
- 保持推理过程的逻辑性和连贯性
- 所有回复都使用中文
- 必须严格按照JSON格式回复，不要添加任何额外的文本`
}

// parseReasoningResponse parses the reasoning response from the model
func parseReasoningResponse(message *schema.Message) (*ReasoningDecision, error) {
	reasoning := &ReasoningDecision{
		Confidence: 0.8, // Default confidence
	}

	content := message.Content

	// Check if there are tool calls in the message
	if len(message.ToolCalls) > 0 {
		// Parse from structured tool calls
		reasoning.Action = "tool_call"
		reasoning.ToolCalls = message.ToolCalls
		// Set thought from content if available
		if strings.TrimSpace(content) != "" {
			reasoning.Thought = strings.TrimSpace(content)
		}
		return reasoning, nil
	}

	// Try to parse as JSON first
	var jsonResponse ReasoningOutput

	// Clean the content - remove markdown code blocks if present
	cleanContent := strings.TrimSpace(content)
	if strings.HasPrefix(cleanContent, "```json") {
		cleanContent = strings.TrimPrefix(cleanContent, "```json")
		cleanContent = strings.TrimSuffix(cleanContent, "```")
		cleanContent = strings.TrimSpace(cleanContent)
	} else if strings.HasPrefix(cleanContent, "```") {
		cleanContent = strings.TrimPrefix(cleanContent, "```")
		cleanContent = strings.TrimSuffix(cleanContent, "```")
		cleanContent = strings.TrimSpace(cleanContent)
	}

	// Try to parse as JSON
	if err := json.Unmarshal([]byte(cleanContent), &jsonResponse); err == nil {
		// Successfully parsed as JSON
		reasoning.Thought = jsonResponse.Thought
		reasoning.Action = strings.ToLower(strings.TrimSpace(jsonResponse.Action))
		reasoning.FinalAnswer = jsonResponse.FinalAnswer
		if jsonResponse.Confidence > 0 {
			reasoning.Confidence = jsonResponse.Confidence
		}

		// Validate action
		if reasoning.Action == "" {
			reasoning.Action = "continue"
		}
		if reasoning.Action != "continue" && reasoning.Action != "tool_call" && reasoning.Action != "final_answer" {
			reasoning.Action = "continue"
		}

		return reasoning, nil
	}

	// Fallback to legacy text parsing if JSON parsing fails
	reasoning.Action = "continue"

	// Parse thought - support multi-line content
	lines := strings.Split(content, "\n")
	var thoughtLines []string
	var inThought bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Check if this line starts a thought section
		if regexp.MustCompile(`(?i)^thought:?\s*(.*)$`).MatchString(line) {
			inThought = true
			// Extract the thought content from this line
			thoughtRegex := regexp.MustCompile(`(?i)^thought:?\s*(.*)$`)
			if match := thoughtRegex.FindStringSubmatch(line); len(match) > 1 {
				if thoughtContent := strings.TrimSpace(match[1]); thoughtContent != "" {
					thoughtLines = append(thoughtLines, thoughtContent)
				}
			}
			continue
		}
		// Check if this line starts action or final answer section
		if regexp.MustCompile(`(?i)^(action|final\s*answer):?`).MatchString(line) {
			inThought = false
			continue
		}
		// If we're in thought section and line is not empty, add it
		if inThought && line != "" {
			thoughtLines = append(thoughtLines, line)
		}
	}

	if len(thoughtLines) > 0 {
		reasoning.Thought = strings.Join(thoughtLines, "\n")
	} else {
		// If no explicit "Thought:" found, check if content has no action/final answer markers
		hasActionMarker := regexp.MustCompile(`(?i)(action|final\s*answer):`).MatchString(content)
		if !hasActionMarker && strings.TrimSpace(content) != "" {
			reasoning.Thought = strings.TrimSpace(content)
		}
	}

	// Parse action - more flexible matching
	actionRegex := regexp.MustCompile(`(?i)action:?\s*(continue|tool_call|final_answer)`)
	if actionMatch := actionRegex.FindStringSubmatch(content); len(actionMatch) > 1 {
		reasoning.Action = strings.ToLower(strings.TrimSpace(actionMatch[1]))
	}

	// Parse final answer - support multi-line content
	var finalAnswerLines []string
	var inFinalAnswer bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Check if this line starts a final answer section
		if regexp.MustCompile(`(?i)^final\s*answer:?\s*(.*)$`).MatchString(line) {
			inFinalAnswer = true
			// Extract the final answer content from this line
			finalRegex := regexp.MustCompile(`(?i)^final\s*answer:?\s*(.*)$`)
			if match := finalRegex.FindStringSubmatch(line); len(match) > 1 {
				if finalContent := strings.TrimSpace(match[1]); finalContent != "" {
					finalAnswerLines = append(finalAnswerLines, finalContent)
				}
			}
			continue
		}
		// If we're in final answer section and line is not empty, add it
		if inFinalAnswer && line != "" {
			finalAnswerLines = append(finalAnswerLines, line)
		}
	}

	if len(finalAnswerLines) > 0 {
		reasoning.FinalAnswer = strings.Join(finalAnswerLines, "\n")
		reasoning.Action = "final_answer"
	}

	return reasoning, nil
}

// ExecutionTrace represents the complete execution trace
type ExecutionTrace struct {
	StartTime    time.Time       `json:"start_time"`
	EndTime      time.Time       `json:"end_time"`
	TotalTime    time.Duration   `json:"total_time"`
	Steps        []ExecutionStep `json:"steps"`
	FinalAnswer  string          `json:"final_answer"`
	Success      bool            `json:"success"`
	ErrorMessage string          `json:"error_message,omitempty"`
}

// ExecutionStep represents a single step in the agent execution
type ExecutionStep struct {
	Iteration  int                `json:"iteration"`
	Timestamp  time.Time          `json:"timestamp"`
	Reasoning  *ReasoningDecision `json:"reasoning"`
	Decision   *DecisionResult    `json:"decision"`
	ToolResult *ToolResult        `json:"tool_result,omitempty"`
	Duration   time.Duration      `json:"duration"`
	Error      string             `json:"error,omitempty"`
}

// ReasoningDecision represents a reasoning decision
type ReasoningDecision struct {
	Thought     string            `json:"thought"`
	Action      string            `json:"action"`
	ToolCalls   []schema.ToolCall `json:"tool_call,omitempty"`
	FinalAnswer string            `json:"final_answer,omitempty"`
	Confidence  float64           `json:"confidence"`
}

// DecisionResult represents the result of a decision
type DecisionResult struct {
	Action      string                 `json:"action"`
	NextAction  string                 `json:"next_action"`
	ToolName    string                 `json:"tool_name,omitempty"`
	ToolArgs    map[string]interface{} `json:"tool_args,omitempty"`
	FinalAnswer string                 `json:"final_answer,omitempty"`
	ShouldStop  bool                   `json:"should_stop"`
	Confidence  float64                `json:"confidence"`
	Reason      string                 `json:"reason"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolName    string      `json:"tool_name"`
	Success     bool        `json:"success"`
	Result      interface{} `json:"result"`
	Error       string      `json:"error,omitempty"`
	Observation string      `json:"observation"`
}
