package react

import (
	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
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