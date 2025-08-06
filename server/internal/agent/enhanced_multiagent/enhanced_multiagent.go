package enhanced_multiagent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/components/model"
)

// EnhancedMultiAgent 增强版多智能体系统
type EnhancedMultiAgent struct {
	config   *EnhancedMultiAgentConfig
	graph    *compose.Graph[[]*schema.Message, *schema.Message]
	runnable compose.Runnable[[]*schema.Message, *schema.Message]
	state    *EnhancedState
}

// NewEnhancedMultiAgent 创建新的增强版多智能体系统
func NewEnhancedMultiAgent(config *EnhancedMultiAgentConfig) (*EnhancedMultiAgent, error) {
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建初始状态
	state := NewEnhancedState(config.MaxRounds)

	// 创建多智能体实例
	agent := &EnhancedMultiAgent{
		config: config,
		state:  state,
	}

	// 构建执行图
	graph, err := agent.buildGraph()
	if err != nil {
		return nil, fmt.Errorf("failed to build graph: %w", err)
	}

	agent.graph = graph

	// 编译为可运行组件
	compileOpts := []compose.GraphCompileOption{
		compose.WithGraphName("EnhancedMultiAgent"),
	}
	runnable, err := graph.Compile(context.Background(), compileOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile graph: %w", err)
	}

	agent.runnable = runnable

	return agent, nil
}

// Generate 生成响应（同步）
func (e *EnhancedMultiAgent) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	// 重置状态
	e.state = NewEnhancedState(e.config.MaxRounds)

	// 执行回调：任务开始
	for _, callback := range e.config.Callbacks {
		if err := callback.OnThinking(ctx, e.state, nil); err != nil {
			return nil, fmt.Errorf("callback error: %w", err)
		}
	}

	// 执行图
	result, err := e.runnable.Invoke(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	// 执行回调：任务完成
	for _, callback := range e.config.Callbacks {
		if err := callback.OnTaskComplete(ctx, e.state, result); err != nil {
			return nil, fmt.Errorf("callback error: %w", err)
		}
	}

	return result, nil
}

// Stream 流式生成响应
func (e *EnhancedMultiAgent) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// 重置状态
	e.state = NewEnhancedState(e.config.MaxRounds)

	// 执行回调：任务开始
	for _, callback := range e.config.Callbacks {
		if err := callback.OnThinking(ctx, e.state, nil); err != nil {
			return nil, fmt.Errorf("callback error: %w", err)
		}
	}

	// 流式执行图
	stream, err := e.runnable.Stream(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("stream execution failed: %w", err)
	}

	return stream, nil
}

// GetState 获取当前状态
func (e *EnhancedMultiAgent) GetState() *EnhancedState {
	return e.state
}

// GetConfig 获取配置
func (e *EnhancedMultiAgent) GetConfig() *EnhancedMultiAgentConfig {
	return e.config
}

// buildGraph 构建执行图
func (e *EnhancedMultiAgent) buildGraph() (*compose.Graph[[]*schema.Message, *schema.Message], error) {
	// 创建图
	graph := compose.NewGraph[[]*schema.Message, *schema.Message]()

	// 添加节点
	err := e.addNodes(graph)
	if err != nil {
		return nil, fmt.Errorf("failed to add nodes: %w", err)
	}

	// 添加边
	err = e.addEdges(graph)
	if err != nil {
		return nil, fmt.Errorf("failed to add edges: %w", err)
	}

	return graph, nil
}

// addNodes 添加节点到图
func (e *EnhancedMultiAgent) addNodes(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	// 添加Host Think节点
	hostThinkLambda := compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
		// 简化实现：直接返回思考结果
		return &schema.Message{
			Role:    schema.Assistant,
			Content: "Host thinking completed",
		}, nil
	})
	err := graph.AddLambdaNode("host_think", hostThinkLambda)
	if err != nil {
		return fmt.Errorf("failed to add host_think node: %w", err)
	}

	// 添加Direct Answer节点
	directAnswerLambda := compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: "Direct answer provided",
		}, nil
	})
	err = graph.AddLambdaNode("direct_answer", directAnswerLambda)
	if err != nil {
		return fmt.Errorf("failed to add direct_answer node: %w", err)
	}

	// 添加Final Answer节点
	finalAnswerLambda := compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: "Final answer generated",
		}, nil
	})
	err = graph.AddLambdaNode("final_answer", finalAnswerLambda)
	if err != nil {
		return fmt.Errorf("failed to add final_answer node: %w", err)
	}

	return nil
}

// addEdges 添加边到图
func (e *EnhancedMultiAgent) addEdges(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	// 简化的边连接：host_think -> direct_answer
	err := graph.AddEdge("host_think", "direct_answer")
	if err != nil {
		return fmt.Errorf("failed to add edge host_think -> direct_answer: %w", err)
	}

	return nil
}

// 配置验证

// validate 验证配置
func (c *EnhancedMultiAgentConfig) validate() error {
	if c.Host.ThinkingModel == nil {
		return fmt.Errorf("host thinking model is required")
	}

	if len(c.Specialists) == 0 {
		return fmt.Errorf("at least one specialist is required")
	}

	for i, specialist := range c.Specialists {
		if err := specialist.validate(); err != nil {
			return fmt.Errorf("specialist %d validation failed: %w", i, err)
		}
	}

	if c.MaxRounds <= 0 {
		c.MaxRounds = 5 // 默认最大轮次
	}

	if c.ComplexityThreshold <= 0 {
		c.ComplexityThreshold = 0.5 // 默认复杂度阈值
	}

	return nil
}

// validate 验证专家配置
func (s *EnhancedSpecialist) validate() error {
	if s.Name == "" {
		return fmt.Errorf("specialist name is required")
	}

	if s.ChatModel == nil {
		return fmt.Errorf("specialist chat model is required")
	}

	return nil
}

// 默认回调实现

// DefaultCallback 默认回调实现
type DefaultCallback struct{}

// OnThinking 思考阶段回调
func (d *DefaultCallback) OnThinking(ctx context.Context, state *EnhancedState, thinking *ThinkingResult) error {
	return nil
}

// OnPlanning 规划阶段回调
func (d *DefaultCallback) OnPlanning(ctx context.Context, state *EnhancedState, plan *TaskPlan) error {
	return nil
}

// OnSpecialistCall 专家调用回调
func (d *DefaultCallback) OnSpecialistCall(ctx context.Context, state *EnhancedState, specialist string, result *SpecialistResult) error {
	return nil
}

// OnFeedback 反馈阶段回调
func (d *DefaultCallback) OnFeedback(ctx context.Context, state *EnhancedState, feedback *FeedbackResult) error {
	return nil
}

// OnTaskComplete 任务完成回调
func (d *DefaultCallback) OnTaskComplete(ctx context.Context, state *EnhancedState, finalAnswer *schema.Message) error {
	return nil
}