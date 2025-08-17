package react

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// Node keys for the ReAct agent graph
const (
	nodeKeyInit         = "init"
	nodeKeyReasoning    = "reasoning"
	nodeKeyTools        = "tools"
	nodeKeyToolsChecker = "tools_checker"
	nodeKeyToReasoning  = "to_reasoning"
	nodeKeyComplete     = "complete"
)

// NewAgent creates a new ReAct agent with the given configuration
func NewAgent(ctx context.Context, config ReactAgentConfig, opts ...base.AgentOption) (*ReactAgent, error) {
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	agent := &ReactAgent{
		Config:       config,
		AgentOptions: opts,
	}

	// Setup chat model
	chatModel, err := agent.setupChatModel(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup chat model: %w", err)
	}

	// Build graph
	graph, err := agent.buildGraph(ctx, chatModel)
	if err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}

	agent.Graph = graph

	// Compile graph to runnable
	runnable, err := graph.Compile(ctx, compose.WithMaxRunSteps(config.MaxStep))
	if err != nil {
		return nil, fmt.Errorf("failed to compile graph: %w", err)
	}

	agent.Runnable = runnable

	return agent, nil
}

// setupChatModel sets up the chat model for the agent
func (a *ReactAgent) setupChatModel(ctx context.Context) (model.BaseChatModel, error) {

	// Use the tool calling model as the base chat model
	chatModel := a.Config.ToolCallingModel
	if chatModel == nil {
		return nil, fmt.Errorf("ToolCallingModel cannot be nil")
	}
	// Generate tool infos for model binding
	toolInfos, err := genToolInfos(ctx, a.Config.ToolsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tool infos: %w", err)
	}

	// Bind tools to chat model (required for tool calling capability)
	if len(toolInfos) > 0 {
		chatModel, err = chatModel.WithTools(toolInfos)
		if err != nil {
			return nil, fmt.Errorf("failed to bind tools: %w", err)
		}
	}
	return chatModel, nil
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

// buildGraph builds the execution graph for the ReAct agent
func (a *ReactAgent) buildGraph(ctx context.Context, chatModel model.BaseChatModel) (*compose.Graph[[]*schema.Message, *schema.Message], error) {
	// Create graph with state enabled
	graph := compose.NewGraph[[]*schema.Message, *schema.Message](compose.WithGenLocalState(func(ctx context.Context) *AgentState {
		return &AgentState{
			Messages:                 make([]*schema.Message, 0),
			ReasoningHistory:         make([]Reasoning, 0),
			Iteration:                0,
			MaxIterations:            a.Config.MaxStep,
			Completed:                false,
			FinalAnswer:              "",
			ReturnDirectlyToolCallID: "",
		}
	}))

	// Add nodes
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

	return graph, nil
}

// addInitNode adds the initialization node to the graph
func (a *ReactAgent) addInitNode(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	initHandler := NewInitHandler(a.Config)
	return graph.AddLambdaNode(nodeKeyInit, compose.InvokableLambda(
		func(ctx context.Context, input []*schema.Message) ([]*schema.Message, error) {
			return input, nil
		}),
		compose.WithStatePreHandler(initHandler.PreHandler),
	)
}

// addReasoningNode adds the reasoning node to the graph
func (a *ReactAgent) addReasoningNode(graph *compose.Graph[[]*schema.Message, *schema.Message], chatModel model.BaseChatModel) error {
	reasoningHandler := NewReasoningHandler(a.Config)
	return graph.AddChatModelNode(nodeKeyReasoning, chatModel,
		compose.WithStatePreHandler(reasoningHandler.PreHandler),
		compose.WithStatePostHandler(reasoningHandler.PostHandler),
	)
}

// addToReasoningNode adds the to-reasoning conversion node
func (a *ReactAgent) addToReasoningNode(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	return graph.AddLambdaNode(nodeKeyToReasoning, compose.ToList[*schema.Message]())
}

// addToolsNode adds the tools execution node to the graph
func (a *ReactAgent) addToolsNode(ctx context.Context, graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	toolsNodeConfig := &a.Config.ToolsConfig
	toolsNode, err := compose.NewToolNode(ctx, toolsNodeConfig)
	if err != nil {
		return fmt.Errorf("failed to create tools node: %w", err)
	}

	toolHandler := NewToolHandler(a.Config)
	return graph.AddToolsNode(nodeKeyTools, toolsNode,
		compose.WithStatePostHandler(toolHandler.PostHandler),
	)
}

// addToolsCheckerNode adds the tools checker node to the graph
func (a *ReactAgent) addToolsCheckerNode(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	return graph.AddLambdaNode(nodeKeyToolsChecker, compose.InvokableLambda(
		func(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
			return a.toolsCheckerNodeHandler(ctx, input)
		}),
	)
}

// toolsCheckerNodeHandler handles the tools checker node logic
func (a *ReactAgent) toolsCheckerNodeHandler(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
	// Find the message that should return directly
	var directReturnMsg *schema.Message
	err := compose.ProcessState(ctx, func(_ context.Context, state *AgentState) error {
		if state.ReturnDirectlyToolCallID != "" {
			// Find the message with the matching tool call ID
			for _, msg := range input {
				if msg.ToolCallID == state.ReturnDirectlyToolCallID {
					directReturnMsg = msg
					state.FinalAnswer = msg.Content
					state.Completed = true
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if directReturnMsg != nil {
		return directReturnMsg, nil
	}

	// If no direct return, return the first message for further processing
	if len(input) > 0 {
		return input[0], nil
	}

	return &schema.Message{Role: schema.Assistant, Content: ""}, nil
}

// addCompleteNode adds the completion node to the graph
func (a *ReactAgent) addCompleteNode(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	completeHandler := NewCompleteHandler(a.Config)
	return graph.AddLambdaNode(nodeKeyComplete, compose.InvokableLambda(
		func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
			return input, nil
		}),
		compose.WithStatePostHandler(completeHandler.PostHandler),
	)
}

// addDecisionBranch adds the decision branch to the graph
func (a *ReactAgent) addDecisionBranch(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
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
func (a *ReactAgent) decisionBranchHandler(ctx context.Context, msgsStream *schema.StreamReader[*schema.Message]) (endNode string, err error) {
	msgsStream.Close()

	// Default to continue reasoning
	endNode = nodeKeyToReasoning

	err = compose.ProcessState(ctx, func(_ context.Context, state *AgentState) error {
		if len(state.ReasoningHistory) == 0 {
			// No reasoning history, continue to reasoning
			endNode = nodeKeyToReasoning
			return nil
		}

		reasoning := &state.ReasoningHistory[len(state.ReasoningHistory)-1]

		// Check if max iterations reached
		if state.Iteration > state.MaxIterations {
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
			if len(reasoning.ToolCalls) > 0 {
				// Route to tools node
				endNode = nodeKeyTools
			} else {
				// No tool calls, continue reasoning
				endNode = nodeKeyToReasoning
			}
		case "final_answer":
			// Set final answer and mark as completed
			state.FinalAnswer = reasoning.FinalAnswer
			state.Completed = true
			endNode = nodeKeyComplete
		default:
			// Route back to reasoning for retry
			endNode = nodeKeyToReasoning
		}

		return nil
	})
	return endNode, err
}

// addToolsCheckerBranch adds the tools checker branch to the graph
func (a *ReactAgent) addToolsCheckerBranch(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	return graph.AddBranch(nodeKeyToolsChecker, compose.NewStreamGraphBranch(
		a.toolsCheckerBranchHandler,
		map[string]bool{
			nodeKeyComplete:    true,
			nodeKeyToReasoning: true,
		},
	))
}

// toolsCheckerBranchHandler handles the tools checker branch logic
func (a *ReactAgent) toolsCheckerBranchHandler(ctx context.Context, msgStream *schema.StreamReader[*schema.Message]) (endNode string, err error) {
	msgStream.Close()

	err = compose.ProcessState(ctx, func(_ context.Context, state *AgentState) error {
		if state.ReturnDirectlyToolCallID != "" && state.Completed {
			endNode = nodeKeyComplete
		} else {
			endNode = nodeKeyToReasoning
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return endNode, nil
}

// addGraphEdges adds all edges to connect the graph nodes
func (a *ReactAgent) addGraphEdges(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
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

	return nil
}

// Generate executes the agent with comprehensive error handling and monitoring
func (a *ReactAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...base.AgentOption) (*schema.Message, error) {
	// Validate input
	if len(messages) == 0 {
		return nil, fmt.Errorf("input messages cannot be empty")
	}

	options := base.GetComposeOptions(opts...)
	options = append(options, base.GetComposeOptions(a.AgentOptions...)...) // 合并option
	result, err := a.Runnable.Invoke(ctx, messages, options...)
	if err != nil {
		return nil, fmt.Errorf("agent execution failed: %w", err)
	}
	return result, nil
}

// Stream executes the agent with streaming support and comprehensive monitoring
func (a *ReactAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...base.AgentOption) (*schema.StreamReader[*schema.Message], error) {
	// Validate input
	if len(messages) == 0 {
		return nil, fmt.Errorf("input messages cannot be empty")
	}

	// Execute streaming with error handling
	options := base.GetComposeOptions(opts...)
	options = append(options, base.GetComposeOptions(a.AgentOptions...)...) // 合并option
	stream, err := a.Runnable.Stream(ctx, messages, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to start streaming: %w", err)
	}

	return stream, nil
}

// parseReasoningResponse parses the reasoning response from the model
func parseReasoningResponse(message *schema.Message) (*Reasoning, error) {
	reasoning := &Reasoning{
		Confidence: 0.8, // Default confidence
	}

	content := message.Content

	// Check if there are tool calls in the message first
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
	reasoning.Thought = cleanContent
	return reasoning, nil
}
