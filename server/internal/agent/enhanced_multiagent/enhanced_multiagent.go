package enhanced_multiagent

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// EnhancedMultiAgent 增强版多智能体系统
type EnhancedMultiAgent struct {
	// 配置
	config *EnhancedMultiAgentConfig

	// 主控Agent
	host *EnhancedHost

	// 专家Agent映射
	specialists map[string]*EnhancedSpecialist

	// 回调处理器
	callbacks []EnhancedMultiAgentCallback

	// 执行图
	graph compose.Runnable[[]*schema.Message, *schema.Message]
}

// NewEnhancedMultiAgent 创建增强版多智能体系统
func NewEnhancedMultiAgent(config *EnhancedMultiAgentConfig) (*EnhancedMultiAgent, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建专家映射
	specialists := make(map[string]*EnhancedSpecialist)
	for _, specialist := range config.Specialists {
		if specialist.Status == SpecialistStatusActive {
			specialists[specialist.Name] = specialist
		}
	}

	return &EnhancedMultiAgent{
		config:      config,
		host:        &config.Host,
		specialists: specialists,
		callbacks:   config.Callbacks,
		graph:       nil, // 将在buildGraph中初始化
	}, nil
}

// RegisterSpecialist 注册专家Agent
func (ema *EnhancedMultiAgent) RegisterSpecialist(specialist *EnhancedSpecialist) error {
	if specialist == nil {
		return fmt.Errorf("specialist cannot be nil")
	}

	if specialist.Status != SpecialistStatusActive {
		return fmt.Errorf("specialist must be active to register")
	}

	ema.specialists[specialist.Name] = specialist
	return nil
}

// UnregisterSpecialist 注销专家Agent
func (ema *EnhancedMultiAgent) UnregisterSpecialist(name string) error {
	if _, exists := ema.specialists[name]; !exists {
		return fmt.Errorf("specialist %s not found", name)
	}

	delete(ema.specialists, name)
	return nil
}

// GetSpecialist 获取专家Agent
func (ema *EnhancedMultiAgent) GetSpecialist(name string) (*EnhancedSpecialist, error) {
	specialist, exists := ema.specialists[name]
	if !exists {
		return nil, fmt.Errorf("specialist %s not found", name)
	}
	return specialist, nil
}

// ListSpecialists 列出所有专家Agent
func (ema *EnhancedMultiAgent) ListSpecialists() []*EnhancedSpecialist {
	specialists := make([]*EnhancedSpecialist, 0, len(ema.specialists))
	for _, specialist := range ema.specialists {
		specialists = append(specialists, specialist)
	}
	return specialists
}

// Invoke 同步调用增强版多智能体系统
func (ema *EnhancedMultiAgent) Invoke(ctx context.Context, input []*schema.Message, opts ...compose.Option) (*schema.Message, error) {
	// 构建执行图
	if ema.graph == nil {
		if err := ema.buildGraph(); err != nil {
			return nil, fmt.Errorf("failed to build graph: %w", err)
		}
	}

	// 触发思考开始回调
	for _, callback := range ema.callbacks {
		ctx = callback.OnThinkingStart(ctx, &ThinkingStartInfo{
			SessionID: fmt.Sprintf("session_%d", time.Now().Unix()),
			Messages:  input,
			Timestamp: time.Now(),
		})
	}

	// 调用执行图
	return ema.graph.Invoke(ctx, input, opts...)
}

// Stream 流式调用增强版多智能体系统
func (ema *EnhancedMultiAgent) Stream(ctx context.Context, input []*schema.Message, opts ...compose.Option) (*schema.StreamReader[*schema.Message], error) {
	// 构建执行图
	if ema.graph == nil {
		if err := ema.buildGraph(); err != nil {
			return nil, fmt.Errorf("failed to build graph: %w", err)
		}
	}

	// 触发思考开始回调
	for _, callback := range ema.callbacks {
		ctx = callback.OnThinkingStart(ctx, &ThinkingStartInfo{
			SessionID: fmt.Sprintf("session_%d", time.Now().Unix()),
			Messages:  input,
			Timestamp: time.Now(),
		})
	}

	// 调用执行图
	return ema.graph.Stream(ctx, input, opts...)
}

// buildGraph 构建执行图
func (ema *EnhancedMultiAgent) buildGraph() error {
	// 创建图构建器
	graphBuilder := compose.NewGraph[[]*schema.Message, *schema.Message]()

	// 添加主处理节点
	err := graphBuilder.AddLambdaNode("main_processor", compose.InvokableLambda(
		func(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
			// 简单的处理逻辑，返回第一条消息
			if len(input) > 0 {
				return input[0], nil
			}
			return &schema.Message{Content: "No input received"}, nil
		},
	))
	if err != nil {
		return fmt.Errorf("failed to add main processor node: %w", err)
	}

	// 添加边
	err = graphBuilder.AddEdge(compose.START, "main_processor")
	if err != nil {
		return fmt.Errorf("failed to add start edge: %w", err)
	}

	err = graphBuilder.AddEdge("main_processor", compose.END)
	if err != nil {
		return fmt.Errorf("failed to add end edge: %w", err)
	}

	// 编译图
	graph, err := graphBuilder.Compile(context.Background())
	if err != nil {
		return fmt.Errorf("failed to compile graph: %w", err)
	}

	ema.graph = graph
	return nil
}

// validateConfig 验证配置
func validateConfig(config *EnhancedMultiAgentConfig) error {
	if config.Host.Model == nil {
		return fmt.Errorf("host model cannot be nil")
	}
	if config.MaxRounds <= 0 {
		config.MaxRounds = 5 // 默认值
	}
	return nil
}

// 占位符方法，用于后续扩展

// addThinkingNode 添加思考节点
func (ema *EnhancedMultiAgent) addThinkingNode(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	// TODO: 实现思考节点
	return nil
}

// addPlanningNode 添加规划节点
func (ema *EnhancedMultiAgent) addPlanningNode(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	// TODO: 实现规划节点
	return nil
}

// addExecutionNodes 添加执行节点
func (ema *EnhancedMultiAgent) addExecutionNodes(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	// TODO: 实现执行节点
	return nil
}

// addFeedbackNode 添加反馈节点
func (ema *EnhancedMultiAgent) addFeedbackNode(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	// TODO: 实现反馈节点
	return nil
}

// addEdges 添加边
func (ema *EnhancedMultiAgent) addEdges(graph *compose.Graph[[]*schema.Message, *schema.Message]) error {
	// TODO: 实现边连接
	return nil
}
