package understanding

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/agent/llmmodel"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func BuildUnderstandingAgent(ctx context.Context) (r compose.Runnable[[]*schema.Message, *schema.Message], err error) {
	cm, err := llmmodel.NewOpenAIModel(ctx, nil)
	if err != nil {
		return nil, err
	}
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendLambda(compose.InvokableLambdaWithOption(func(ctx context.Context, input []*schema.Message, opts ...any) (output []*schema.Message, err error) {
		systemMsg := schema.SystemMessage(systemPrompt)
		return append([]*schema.Message{systemMsg}, input...), nil
	})).AppendChatModel(cm)
	return chain.Compile(ctx, compose.WithGraphName("understanding"))
}
