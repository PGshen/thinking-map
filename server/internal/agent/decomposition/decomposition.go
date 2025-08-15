package decomposition

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/PGshen/thinking-map/server/internal/agent/base/multiagent"
	"github.com/PGshen/thinking-map/server/internal/agent/base/react"
	"github.com/PGshen/thinking-map/server/internal/agent/llmmodel"
	"github.com/PGshen/thinking-map/server/internal/agent/tool/node"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

func BuildDecompositionAgent(ctx context.Context, option base.AgentOption) (*multiagent.MultiAgent, error) {
	// 创建主模型
	cm, err := llmmodel.NewOpenAIModel(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 创建拆解决策Agent specialist
	decompositionDecisionAgent, err := buildDecompositionDecisionAgent(ctx, option)
	if err != nil {
		return nil, err
	}

	// 创建问题拆解Agent specialist
	problemDecompositionAgent, err := buildProblemDecompositionAgent(ctx, option)
	if err != nil {
		return nil, err
	}

	// 配置MultiAgent
	config := &multiagent.MultiAgentConfig{
		Name:        "DecompositionAgent",
		Description: "负责问题拆解的多智能体系统",
		Host: multiagent.Host{
			Model:        cm,
			SystemPrompt: buildHostSystemPrompt(),
		},
		Specialists: []*multiagent.Specialist{
			decompositionDecisionAgent,
			problemDecompositionAgent,
		},
		MaxRounds: 10,
	}

	// 创建MultiAgent实例
	agent, err := multiagent.NewMultiAgent(ctx, config)
	if err != nil {
		return nil, err
	}

	return agent, nil
}

// buildDecompositionDecisionAgent 创建拆解决策Agent specialist
func buildDecompositionDecisionAgent(ctx context.Context, option base.AgentOption) (*multiagent.Specialist, error) {
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
	}, option)
	if err != nil {
		return nil, err
	}

	return &multiagent.Specialist{
		Name:         "DecompositionDecisionAgent",
		IntendedUse:  "分析问题复杂度并决定拆解策略类型（顺序型、并行型、层次型、探索型）",
		// ChatModel:    cm,
		SystemPrompt: buildDecompositionDecisionPrompt(),
		Invokable:    agent.Generate,
		Streamable:   agent.Stream,
	}, nil
}

// buildProblemDecompositionAgent 创建问题拆解Agent specialist
func buildProblemDecompositionAgent(ctx context.Context, option base.AgentOption) (*multiagent.Specialist, error) {
	// 创建模型
	cm, err := llmmodel.NewOpenAIModel(ctx, nil)
	if err != nil {
		return nil, err
	}

	// 添加节点创建工具、知识检索工具
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
	}, option)
	if err != nil {
		return nil, err
	}

	return &multiagent.Specialist{
		Name:        "ProblemDecompositionAgent",
		IntendedUse: "基于拆解策略将复杂问题分解为可管理的子问题，创建子节点并设置依赖关系",
		// ChatModel:    cm,
		SystemPrompt: buildProblemDecompositionPrompt(),
		Invokable:    agent.Generate,
		Streamable:   agent.Stream,
	}, nil
}
