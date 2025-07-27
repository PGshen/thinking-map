package react

import (
	"github.com/cloudwego/eino/schema"
)

// AgentState represents the global state of the agent used with Eino's state management
type AgentState struct {
	Messages                 []*schema.Message   `json:"messages"`
	ReasoningHistory         []ReasoningDecision `json:"reasoning_history"`
	Iteration                int                 `json:"iteration"`
	MaxIterations            int                 `json:"max_iterations"`
	Completed                bool                `json:"completed"`
	FinalAnswer              string              `json:"final_answer"`
	ReturnDirectlyToolCallID string              `json:"return_directly_tool_call_id"`
}

// State management is now handled by Eino's global state support
// All helper types are defined in react_agent.go
