package conclusion

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/PGshen/thinking-map/server/internal/agent/base/multiagent"
	"github.com/PGshen/thinking-map/server/internal/agent/base/react"
	"github.com/PGshen/thinking-map/server/internal/agent/llmmodel"
	"github.com/PGshen/thinking-map/server/internal/agent/tool/node"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// BuildConclusionAgent 构建问题结论Agent
func BuildConclusionAgent(ctx context.Context, option ...base.AgentOption) (compose.Runnable[[]*schema.Message, *schema.Message], error) {
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

	// 创建结论生成Agent specialist
	conclusionGenerationAgent, err := buildConclusionGenerationAgent(ctx, option...)
	if err != nil {
		return nil, err
	}

	// 创建结论优化Agent specialist
	conclusionOptimizationAgent, err := buildConclusionOptimizationAgent(ctx, option...)
	if err != nil {
		return nil, err
	}

	// react
	reactAgent, err := react.NewAgent(ctx, react.ReactAgentConfig{
		ToolCallingModel: cm,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: allTools,
		},
	}, option...)
	if err != nil {
		return nil, err
	}

	// 配置MultiAgent
	config := &multiagent.MultiAgentConfig{
		Name:        "ConclusionAgent",
		Description: "负责问题结论生成和优化的多智能体系统",
		Host: multiagent.Host{
			Model:      cm,
			ReactAgent: reactAgent,
			Planning: multiagent.PlanningConfig{
				PlanningPrompt: buildConclusionPlanningPrompt(),
			},
		},
		Specialists: []*multiagent.Specialist{
			conclusionGenerationAgent,
			conclusionOptimizationAgent,
		},
		MaxRounds: 15,
	}

	// 创建MultiAgent实例
	agent, err := multiagent.NewMultiAgent(ctx, config)
	if err != nil {
		return nil, err
	}

	return agent.Runnable, nil
}

// buildConclusionGenerationAgent 创建结论生成Agent specialist
func buildConclusionGenerationAgent(ctx context.Context, option ...base.AgentOption) (*multiagent.Specialist, error) {
	// 创建模型
	cm, err := llmmodel.NewOpenAIModel(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 添加节点操作工具
	allTools := []tool.BaseTool{}
	nodeTools, err := node.GetAllNodeTools()
	if err != nil {
		return nil, err
	}
	allTools = append(allTools, nodeTools...)

	agent, err := react.NewAgent(ctx, react.ReactAgentConfig{
		ToolCallingModel: cm,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: allTools,
		},
	}, option...)
	if err != nil {
		return nil, err
	}

	return &multiagent.Specialist{
		Name:         "ConclusionGenerationAgent",
		IntendedUse:  "基于背景上下文和用户指令生成初始结论，适用于首次结论生成场景",
		SystemPrompt: buildConclusionGenerationPrompt(),
		ReactAgent:   agent,
	}, nil
}

// buildConclusionOptimizationAgent 创建结论优化Agent specialist
func buildConclusionOptimizationAgent(ctx context.Context, option ...base.AgentOption) (*multiagent.Specialist, error) {
	// 创建模型
	cm, err := llmmodel.NewOpenAIModel(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 添加节点操作工具
	allTools := []tool.BaseTool{}
	nodeTools, err := node.GetAllNodeTools()
	if err != nil {
		return nil, err
	}
	allTools = append(allTools, nodeTools...)

	agent, err := react.NewAgent(ctx, react.ReactAgentConfig{
		ToolCallingModel: cm,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: allTools,
		},
	}, option...)
	if err != nil {
		return nil, err
	}

	return &multiagent.Specialist{
		Name:         "ConclusionOptimizationAgent",
		IntendedUse:  "基于现有结论、引用内容和优化指令进行结论的局部优化改写",
		SystemPrompt: buildConclusionOptimizationPrompt(),
		ReactAgent:   agent,
	}, nil
}
