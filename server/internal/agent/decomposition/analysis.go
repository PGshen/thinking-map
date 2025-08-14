package decomposition

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/PGshen/thinking-map/server/internal/agent/base/react"
	"github.com/PGshen/thinking-map/server/internal/agent/llmmodel"
	"github.com/PGshen/thinking-map/server/internal/agent/tool/messaging"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// 分析Agent
func BuildAnalysisAgent(ctx context.Context, option base.AgentOption) (r compose.Runnable[[]*schema.Message, *schema.Message], err error) {
	cm, err := llmmodel.NewOpenAIModel(ctx, nil)
	if err != nil {
		return nil, err
	}
	sendActionMsg, err := messaging.ActionTool()
	if err != nil {
		return nil, err
	}
	agent, err := react.NewAgent(ctx, react.ReactAgentConfig{
		ToolCallingModel: cm,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{
				sendActionMsg,
			},
		},
		ToolReturnDirectly: map[string]bool{
			"sendActionMsg": true,
		},
	}, option)
	if err != nil {
		return nil, err
	}
	lbaAgent, err := compose.AnyLambda(agent.Generate, agent.Stream, nil, nil)
	if err != nil {
		return nil, err
	}
	// 构建链
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendLambda(compose.InvokableLambdaWithOption(func(ctx context.Context, input []*schema.Message, opts ...any) (output []*schema.Message, err error) {
		systemMsg := schema.SystemMessage(analysisPrompt)
		return append([]*schema.Message{systemMsg}, input...), nil
	})).AppendLambda(lbaAgent)
	return chain.Compile(ctx, compose.WithGraphName("analysis"))
}
