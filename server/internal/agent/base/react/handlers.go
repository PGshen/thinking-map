package react

import (
    "context"
    "fmt"
    "strings"

    "github.com/cloudwego/eino/schema"
)

type InitHandler struct {
	config ReactAgentConfig
}

func NewInitHandler(config ReactAgentConfig) *InitHandler {
	return &InitHandler{
		config: config,
	}
}

func (h *InitHandler) PreHandler(ctx context.Context, input []*schema.Message, state *AgentState) ([]*schema.Message, error) {
	// Initialize state
	state.Messages = append(state.Messages, input...)
	state.Iteration = 0
	state.Completed = false
	state.FinalAnswer = ""
	state.ReturnDirectlyToolCallID = ""
	return input, nil
}

type ReasoningHandler struct {
	config ReactAgentConfig
}

func NewReasoningHandler(config ReactAgentConfig) *ReasoningHandler {
	return &ReasoningHandler{
		config: config,
	}
}

func (h *ReasoningHandler) PreHandler(ctx context.Context, input []*schema.Message, state *AgentState) ([]*schema.Message, error) {
    state.Iteration++
    // Build system prompt
    systemPrompt := buildReasoningSystemPrompt()
    // If we are in the forced final iteration, append strict constraints
    if state.ForceFinalAnswer {
        systemPrompt = systemPrompt + `

## Final Iteration Constraint
You have reached or exceeded the maximum number of iterations. This is the last reasoning step.
- You MUST set "action" to "final_answer".
- You MUST provide "final_answer".
- You MUST NOT return "tool_call" nor "continue".
- You MUST strictly follow the JSON response format described above.
If you fail to provide a final answer, the system will treat your latest content as the final answer.`
    }

	// Prepare messages with system prompt
	messages := []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
	}
	messages = append(messages, state.Messages...)
	fmt.Printf("---START1---\n")
	for i, message := range messages {
		fmt.Printf("ReasoningHandler: %d %s %s\n", i, message.Role, message.Content)
	}
	fmt.Printf("---START2---\n")

	return messages, nil
}

func (h *ReasoningHandler) PostHandler(ctx context.Context, output *schema.Message, state *AgentState) (*schema.Message, error) {
    // Parse reasoning response
    reasoning, err := parseReasoningResponse(output)
    if err != nil {
        return nil, fmt.Errorf("failed to parse reasoning response: %w", err)
    }
    // Enforce final answer on forced final iteration
    if state.ForceFinalAnswer {
        if strings.ToLower(strings.TrimSpace(reasoning.Action)) != "final_answer" {
            // Convert any non-final action to final_answer
            reasoning.Action = "final_answer"
            if strings.TrimSpace(reasoning.FinalAnswer) == "" {
                // Prefer structured thought, otherwise fallback to raw content
                if strings.TrimSpace(reasoning.Thought) != "" {
                    reasoning.FinalAnswer = strings.TrimSpace(reasoning.Thought)
                } else {
                    reasoning.FinalAnswer = strings.TrimSpace(output.Content)
                }
            }
            // Clear tool calls to avoid further execution
            reasoning.ToolCalls = nil
        }
    }
	// if len(reasoning.ToolCalls) > 0 {
	// 	for _, toolCall := range reasoning.ToolCalls {
	// 		fmt.Printf("ReasoningHandler: tool call %s %s\n", toolCall.Function.Name, toolCall.Function.Arguments)
	// 	}
	// }
	fmt.Printf("---END1---\n")
	fmt.Println(output.Content)
	fmt.Printf("---END2---\n")

    // Add to reasoning history
    state.ReasoningHistory = append(state.ReasoningHistory, *reasoning)

	// Add output to messages
	state.Messages = append(state.Messages, output)

	return output, nil
}

type ToolHandler struct {
	config ReactAgentConfig
}

func NewToolHandler(config ReactAgentConfig) *ToolHandler {
	return &ToolHandler{
		config: config,
	}
}

func (h *ToolHandler) PostHandler(ctx context.Context, output []*schema.Message, state *AgentState) ([]*schema.Message, error) {
	// Update state with tool result
	state.Messages = append(state.Messages, output...)

	// Check if any tool should return directly
	if h.config.ToolReturnDirectly != nil {
		for _, msg := range output {
			if msg.ToolCallID != "" && len(state.ReasoningHistory) > 0 {
				// Find the corresponding tool call to get tool name
				reasoning := state.ReasoningHistory[len(state.ReasoningHistory)-1]
				for _, toolCall := range reasoning.ToolCalls {
					if toolCall.ID == msg.ToolCallID {
						if shouldReturn, exists := h.config.ToolReturnDirectly[toolCall.Function.Name]; exists && shouldReturn {
							state.ReturnDirectlyToolCallID = msg.ToolCallID
							break
						}
					}
				}
			}
		}
	}
	return output, nil
}

type CompleteHandler struct {
	config ReactAgentConfig
}

func NewCompleteHandler(config ReactAgentConfig) *CompleteHandler {
	return &CompleteHandler{
		config: config,
	}
}

func (h *CompleteHandler) PostHandler(ctx context.Context, output *schema.Message, state *AgentState) (*schema.Message, error) {
	// Create final message based on state
	var finalMessage *schema.Message

	if state.FinalAnswer != "" {
		finalMessage = &schema.Message{
			Role:    schema.Assistant,
			Content: state.FinalAnswer,
		}
	} else if len(state.Messages) > 0 {
		finalMessage = state.Messages[len(state.Messages)-1]
	} else {
		finalMessage = &schema.Message{
			Role:    schema.Assistant,
			Content: "I apologize, but I was unable to provide a response.",
		}
	}

	return finalMessage, nil
}
