package react

import (
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// MessageModifier is a function type for modifying messages
type MessageModifier func([]*schema.Message) []*schema.Message

// DefaultConfig returns a default agent configuration
func DefaultConfig() *AgentConfig {
	return &AgentConfig{
		MaxStep:   10,
		DebugMode: false,
	}
}

// WithModel sets the chat model for the agent
func (c *AgentConfig) WithModel(model model.ToolCallingChatModel) *AgentConfig {
	c.ToolCallingModel = model
	return c
}

// WithTools sets the tools configuration for the agent
func (c *AgentConfig) WithTools(tools compose.ToolsNodeConfig) *AgentConfig {
	c.ToolsConfig = tools
	return c
}

// WithMessageModifier sets the message modifier for the agent
func (c *AgentConfig) WithMessageModifier(modifier MessageModifier) *AgentConfig {
	c.MessageModifier = modifier
	return c
}

// WithMaxStep sets the maximum number of reasoning steps
func (c *AgentConfig) WithMaxStep(maxStep int) *AgentConfig {
	c.MaxStep = maxStep
	return c
}

// WithDebugMode enables or disables debug mode
func (c *AgentConfig) WithDebugMode(debug bool) *AgentConfig {
	c.DebugMode = debug
	return c
}

// WithGraphOptions sets the graph options
func (c *AgentConfig) WithGraphOptions(opts ...compose.GraphAddNodeOpt) *AgentConfig {
	c.GraphOptions = opts
	return c
}
