package intent

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/agent/llmmodel"
	"github.com/PGshen/thinking-map/server/internal/agent/tool/messaging"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// 意图识别Agent
func BuildIntentRecognitionAgent(ctx context.Context) (r compose.Runnable[[]*schema.Message, *schema.Message], err error) {
	cm, err := llmmodel.NewOpenAIModel(ctx, nil)
	if err != nil {
		return nil, err
	}
	userChoiceTool, err := messaging.UserChoiceTool()
	if err != nil {
		return nil, err
	}
	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		Model: cm,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{
				userChoiceTool,
			},
		},
		ToolReturnDirectly: map[string]struct{}{
			"user_choice": {},
		},
	})
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
		systemMsg := schema.SystemMessage(systemPrompt)
		return append([]*schema.Message{systemMsg}, input...), nil
	})).AppendLambda(lbaAgent)
	return chain.Compile(ctx, compose.WithGraphName("intent"))
}
