package react

import (
	"github.com/cloudwego/eino/schema"
)

// AgentState represents the global state of the agent used with Eino's state management
type AgentState struct {
    Messages                 []*schema.Message `json:"messages"`
    ReasoningHistory         []Reasoning       `json:"reasoning_history"`
    Iteration                int               `json:"iteration"`
    MaxIterations            int               `json:"max_iterations"`
    Completed                bool              `json:"completed"`
    FinalAnswer              string            `json:"final_answer"`
    ReturnDirectlyToolCallID string            `json:"return_directly_tool_call_id"`
    // ForceFinalAnswer indicates the next reasoning step must produce a final answer
    // and disallow tool calls or continued thinking.
    ForceFinalAnswer         bool              `json:"force_final_answer"`
}

// State management is now handled by Eino's global state support
// All helper types are defined in types.go
