package react

import (
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

// ReactAgentConfig is the configuration for the ReAct agent
type ReactAgentConfig struct {
	// Model for reasoning and tool calling
	ToolCallingModel model.ToolCallingChatModel

	// ReasoningPrompt is the prompt template for reasoning
	ReasoningPrompt string

	// IsSupportStructuredOutput is whether the model supports structured output
	IsSupportStructuredOutput bool

	// Tools available to the agent
	ToolsConfig compose.ToolsNodeConfig

	// Message modifier for preprocessing
	MessageModifier MessageModifier

	// Maximum number of reasoning iterations
	MaxStep int

	// Enable debug mode
	DebugMode bool

	// Custom graph options
	GraphOptions []compose.GraphAddNodeOpt

	// Tools that should return directly without further reasoning
	// Map of tool name to whether it should return directly
	ToolReturnDirectly map[string]bool
}

// DefaultConfig returns a default agent configuration
func DefaultConfig() *ReactAgentConfig {
	return &ReactAgentConfig{
		MaxStep:   10,
		DebugMode: false,
	}
}

// validateConfig validates the agent configuration
func validateConfig(config *ReactAgentConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if config.ToolCallingModel == nil {
		return fmt.Errorf("ToolCallingModel cannot be nil")
	}
	if config.MaxStep <= 0 {
		config.MaxStep = 10 // default value
	}
	return nil
}
