package react

import (
	"github.com/PGshen/thinking-map/server/internal/agent/base"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// MessageModifier is a function type for modifying messages
type MessageModifier func([]*schema.Message) []*schema.Message

// ReactAgent represents the ReAct agent
type ReactAgent struct {
	Runnable         compose.Runnable[[]*schema.Message, *schema.Message]
	Graph            *compose.Graph[[]*schema.Message, *schema.Message]
	GraphAddNodeOpts []compose.GraphAddNodeOpt
	AgentOptions     []base.AgentOption
	Config           ReactAgentConfig
}

// ReasoningOutput represents the output of reasoning process
type ReasoningOutput struct {
	Thought     string  `json:"thought"`
	Action      string  `json:"action"`
	FinalAnswer string  `json:"final_answer"`
	Confidence  float64 `json:"confidence"`
}

// Reasoning represents the result of reasoning process
type Reasoning struct {
	Thought     string            `json:"thought"`
	Action      string            `json:"action"`
	ToolCalls   []schema.ToolCall `json:"tool_call,omitempty"`
	FinalAnswer string            `json:"final_answer,omitempty"`
	Confidence  float64           `json:"confidence"`
}
