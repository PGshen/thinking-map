package react

import (
	"context"

	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	ub "github.com/cloudwego/eino/utils/callbacks"
)

// WithToolOptions returns an agent option that specifies tool.Option for the tools in agent.
func WithToolOptions(opts ...tool.Option) base.AgentOption {
	return base.WithComposeOptions(compose.WithToolsNodeOption(compose.WithToolOption(opts...)))
}

// WithChatModelOptions returns an agent option that specifies model.Option for the chat model in agent.
func WithChatModelOptions(opts ...model.Option) base.AgentOption {
	return base.WithComposeOptions(compose.WithChatModelOption(opts...))
}

// WithToolList returns an agent option that specifies the list of tools can be called which are BaseTool but must implement InvokableTool or StreamableTool.
func WithToolList(tools ...tool.BaseTool) base.AgentOption {
	return base.WithComposeOptions(compose.WithToolsNodeOption(compose.WithToolList(tools...)))
}

type MessageHandler interface {
	OnMessage(ctx context.Context, message *schema.Message) (context.Context, error)
	OnStreamMessage(ctx context.Context, message *schema.StreamReader[*schema.Message]) (context.Context, error)
}

// WithMessageHandler returns an agent option that specifies the message handler.
func WithMessageHandler(handler MessageHandler) base.AgentOption {
	cmHandler := &ub.ModelCallbackHandler{
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
			ctx, _ = handler.OnMessage(ctx, output.Message)
			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, runInfo *callbacks.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
			c := func(output *model.CallbackOutput) (*schema.Message, error) {
				return output.Message, nil
			}
			s := schema.StreamReaderWithConvert(output, c)
			ctx, _ = handler.OnStreamMessage(ctx, s)
			return ctx
		},
	}
	cb := ub.NewHandlerHelper().ChatModel(cmHandler).Handler()
	option := base.WithComposeOptions(compose.WithCallbacks(cb).DesignateNode(nodeKeyReasoning))
	return option
}
