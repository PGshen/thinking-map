package enhanced_multiagent

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/schema"
)

// ExampleUsage 演示如何使用EnhancedMultiAgent
func ExampleUsage() {
	// 1. 创建配置
	config := &EnhancedMultiAgentConfig{
		Name:      "示例多智能体系统",
		MaxRounds: 5,
		Host: EnhancedHost{
			// 注意：这里需要实际的模型实例
			// ToolCallingModel: yourToolCallingModel,
			// ThinkingModel: yourThinkingModel,
			SystemPrompt:    "你是一个智能助手，能够分析问题并协调多个专家来解决复杂任务。",
			ThinkingPrompt:  "请仔细分析用户的问题，评估其复杂度，并决定下一步行动。",
			PlanningPrompt:  "基于问题分析，制定详细的执行计划。",
			ReflectionPrompt: "反思执行结果，提供改进建议。",
		},
		Specialists: []*EnhancedSpecialist{
			{
				AgentMeta: AgentMeta{
					Name:        "技术专家",
					Description: "专门处理技术相关问题",
					Version:     "1.0",
				},
				// ChatModel: yourChatModel,
				SystemPrompt: "你是一个技术专家，擅长解决编程、架构设计等技术问题。",
				Capabilities: []string{"编程", "架构设计", "技术咨询"},
			},
			{
				AgentMeta: AgentMeta{
					Name:        "业务专家",
					Description: "专门处理业务相关问题",
					Version:     "1.0",
				},
				// ChatModel: yourChatModel,
				SystemPrompt: "你是一个业务专家，擅长分析业务需求和流程优化。",
				Capabilities: []string{"需求分析", "流程优化", "业务咨询"},
			},
		},
		ComplexityThreshold: 0.7,
		Callbacks: []EnhancedMultiAgentCallback{
			&ExampleCallback{},
		},
	}
	
	// 2. 创建多智能体实例
	agent, err := NewEnhancedMultiAgent(config)
	if err != nil {
		log.Fatalf("创建多智能体失败: %v", err)
	}
	
	// 3. 准备输入消息
	messages := []*schema.Message{
		{
			Role:    schema.User,
			Content: "我需要设计一个高并发的微服务架构，能够处理每秒10万次请求，请帮我制定详细的技术方案。",
		},
	}
	
	// 4. 执行任务
	ctx := context.Background()
	result, err := agent.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("执行任务失败: %v", err)
	}
	
	// 5. 输出结果
	fmt.Printf("最终答案: %s\n", result.Content)
	
	// 6. 获取执行状态
	state := agent.GetState()
	fmt.Printf("执行轮次: %d\n", state.CurrentRound)
	fmt.Printf("任务复杂度: %s\n", state.CurrentThinkingResult.Complexity.String())
	if state.CurrentPlan != nil {
		fmt.Printf("执行计划版本: %d\n", state.CurrentPlan.Version)
		fmt.Printf("计划步骤数: %d\n", len(state.CurrentPlan.Steps))
	}
}

// ExampleCallback 示例回调实现
type ExampleCallback struct{}

func (c *ExampleCallback) OnThinking(ctx context.Context, state *EnhancedState, thinking *ThinkingResult) error {
	fmt.Printf("[思考] %s\n", thinking.Thought)
	return nil
}

func (c *ExampleCallback) OnPlanning(ctx context.Context, state *EnhancedState, plan *TaskPlan) error {
	fmt.Printf("[规划] 创建了包含 %d 个步骤的执行计划\n", len(plan.Steps))
	return nil
}

func (c *ExampleCallback) OnSpecialistCall(ctx context.Context, state *EnhancedState, specialist string, result *SpecialistResult) error {
	fmt.Printf("[专家] %s 执行完成，状态: %s\n", specialist, result.Status.String())
	return nil
}

func (c *ExampleCallback) OnFeedback(ctx context.Context, state *EnhancedState, feedback *FeedbackResult) error {
	fmt.Printf("[反馈] %s\n", feedback.Feedback)
	return nil
}

func (c *ExampleCallback) OnTaskComplete(ctx context.Context, state *EnhancedState, finalAnswer *schema.Message) error {
	fmt.Printf("[完成] 任务执行完成\n")
	return nil
}

// CreateTestConfig 创建测试配置（用于单元测试）
func CreateTestConfig() *EnhancedMultiAgentConfig {
	return &EnhancedMultiAgentConfig{
		Name:      "测试多智能体系统",
		MaxRounds: 3,
		Host: EnhancedHost{
			SystemPrompt:     "测试系统提示",
			ThinkingPrompt:   "测试思考提示",
			PlanningPrompt:   "测试规划提示",
			ReflectionPrompt: "测试反思提示",
		},
		Specialists: []*EnhancedSpecialist{
			{
				AgentMeta: AgentMeta{
					Name:        "测试专家",
					Description: "用于测试的专家",
					Version:     "1.0",
				},
				SystemPrompt: "你是一个测试专家",
				Capabilities: []string{"测试"},
			},
		},
		ComplexityThreshold: 0.5,
		Callbacks:           []EnhancedMultiAgentCallback{},
	}
}