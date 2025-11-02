package conclusionv3

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/PGshen/thinking-map/server/internal/agent/base/react"
	"github.com/PGshen/thinking-map/server/internal/agent/llmmodel"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func BuildGenerationAgent(ctx context.Context, option ...base.AgentOption) (r compose.Runnable[[]*schema.Message, *schema.Message], err error) {
	cm, err := llmmodel.NewOpenAIModel(ctx, nil)
	if err != nil {
		return nil, err
	}
	agent, err := react.NewAgent(ctx, react.ReactAgentConfig{
		ToolCallingModel: cm,
	}, option...)
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
		systemMsg := schema.SystemMessage(buildConclusionGenerationPrompt())
		return append([]*schema.Message{systemMsg}, input...), nil
	})).AppendLambda(lbaAgent)
	return chain.Compile(ctx, compose.WithGraphName("conclusion_generation"))
}
